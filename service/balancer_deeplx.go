package service

import (
	"context"
	"deeplx-local/domain"
	"github.com/imroc/req/v3"
	lop "github.com/samber/lo/parallel"
	"github.com/sourcegraph/conc/pool"
	"github.com/sourcegraph/conc/stream"
	"log"
	"regexp"
	"strings"
	"sync/atomic"
	"time"
)

const maxLength = 4096

type Server struct {
	URL           string
	Weight        int64
	CurrentWeight int64
}

type LoadBalancer struct {
	Servers []*Server
	re      *regexp.Regexp
	client  *req.Client
}

// NewLoadBalancer 负载均衡
func NewLoadBalancer(vlist *[]string) TranslateService {
	servers := lop.Map(*vlist, func(item string, index int) *Server {
		return &Server{URL: item, Weight: 1, CurrentWeight: 1}
	})
	return &LoadBalancer{
		Servers: servers,
		client:  req.NewClient().SetTimeout(2 * time.Second),
		re:      regexp.MustCompile(`[^.!?]+[.!?]`),
	}
}

func (lb *LoadBalancer) GetTranslateData(trReq domain.TranslateRequest) domain.TranslateResponse {
	text := trReq.Text
	textLength := len(text)

	if textLength <= maxLength {
		return lb.sendRequest(trReq)
	}

	var textParts []string
	var currentPart string

	sentences := lb.re.FindAllString(text, -1)

	for _, sentence := range sentences {
		if len(currentPart)+len(sentence) <= maxLength {
			currentPart += sentence
		} else {
			textParts = append(textParts, currentPart)
			currentPart = sentence
		}
	}

	if currentPart != "" {
		textParts = append(textParts, currentPart)
	}

	var results = make([]string, 0, len(textParts))
	s := stream.New()

	for _, part := range textParts {
		s.Go(func() stream.Callback {
			transReq := domain.TranslateRequest{
				Text:       part,
				SourceLang: trReq.SourceLang,
				TargetLang: trReq.TargetLang,
			}
			res := lb.sendRequest(transReq)
			return func() {
				results = append(results, res.Data)
			}
		})
	}

	s.Wait()

	return domain.TranslateResponse{
		Code: 200,
		Data: strings.Join(results, ""),
	}
}

func (lb *LoadBalancer) sendRequest(trReq domain.TranslateRequest) domain.TranslateResponse {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelFunc()
	resultChan := make(chan domain.TranslateResponse, 5)

	contextPool := pool.New().WithContext(ctx).WithMaxGoroutines(5)
	for i := 0; i < 5; i++ {
		contextPool.Go(func(ctx context.Context) error {
			server := lb.getServer()
			var trResult domain.TranslateResponse
			start := time.Now()
			response, err := lb.client.R().
				SetContext(ctx).
				SetBody(trReq).
				SetSuccessResult(&trResult).
				Post(server.URL)
			elapsed := time.Since(start)
			lb.updateResponseTime(server, elapsed)

			if err != nil {
				return err
			}
			response.Body.Close()

			if trResult.Code == 200 && len(trResult.Data) > 0 {
				resultChan <- trResult
				cancelFunc()
			}
			return nil
		})
	}

	//go func() {
	//	_ = contextPool.Wait()
	//	if _, ok := <-resultChan; !ok { // 如果通道已经关闭，直接返回
	//		return
	//	}
	//	close(resultChan)
	//}()

	select {
	case r := <-resultChan:
		defer func() {
			cancelFunc()
			close(resultChan)
		}()
		return r
	case <-ctx.Done():
		close(resultChan)
		log.Println("all requests failed")
	}
	return domain.TranslateResponse{}
}

func (lb *LoadBalancer) getServer() *Server {
	var bestServer *Server
	var total int64

	for _, server := range lb.Servers {
		currentWeight := atomic.AddInt64(&server.CurrentWeight, server.Weight)
		atomic.AddInt64(&total, server.Weight)

		if bestServer == nil || currentWeight > atomic.LoadInt64(&bestServer.CurrentWeight) {
			bestServer = server
		}
	}

	if bestServer != nil {
		atomic.AddInt64(&bestServer.CurrentWeight, -total)
	}

	return bestServer
}

func (lb *LoadBalancer) updateResponseTime(server *Server, responseTime time.Duration) {
	prevWeight := atomic.LoadInt64(&server.Weight)
	newWeight := calculateWeight(prevWeight, responseTime)
	atomic.StoreInt64(&server.Weight, newWeight)
}

// 一个更加平滑的权重计算, 指数移动平均(Exponential Moving Average, EMA)
// 使用EMA, 可以根据服务器的历史响应时间来计算当前的权重,这样可以避免权重变化过于剧烈
// k 是一个平滑因子 (0,1)
// 当responseTime很小时,k接近1,这意味着新的权重值主要由当前的响应时间决定
// 当responseTime很大时,k接近0,这意味着新的权重值主要由历史权重值决定
func calculateWeight(prevWeight int64, responseTime time.Duration) int64 {
	k := 2.0 / (1 + float64(responseTime)/float64(time.Second))
	return int64(float64(prevWeight)*k + (1-k)*float64(time.Second)/float64(responseTime+1))
}
