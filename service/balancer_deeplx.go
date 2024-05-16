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
	isAvailable   bool
}

type LoadBalancer struct {
	Servers []*Server
	re      *regexp.Regexp
	client  *req.Client
	index   uint32
}

// NewLoadBalancer 负载均衡
func NewLoadBalancer(vlist *[]string) TranslateService {
	servers := lop.Map(*vlist, func(item string, index int) *Server {
		return &Server{URL: item, Weight: 1, CurrentWeight: 1, isAvailable: true}
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
			response, err := lb.client.R().
				SetContext(ctx).
				SetBody(trReq).
				SetSuccessResult(&trResult).
				Post(server.URL)

			if err != nil {
				return err
			}
			response.Body.Close()

			if trResult.Code == 200 && len(trResult.Data) > 0 {
				resultChan <- trResult
				cancelFunc()
			} else {
				server.isAvailable = false
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
	index := atomic.AddUint32(&lb.index, 1) - 1
	server := lb.Servers[index%uint32(len(lb.Servers))]

	for !server.isAvailable {
		index = atomic.AddUint32(&lb.index, 1) - 1
		server = lb.Servers[index%uint32(len(lb.Servers))]
	}
	return server
}
