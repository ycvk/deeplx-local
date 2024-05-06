package service

import (
	"context"
	"deeplx-local/domain"
	"github.com/imroc/req/v3"
	"github.com/sourcegraph/conc/pool"
	"log"
	"math/rand"
	"time"
)

type TranslateService interface {
	GetTranslateData(trReq domain.TranslateRequest) domain.TranslateResponse
}

type DeepLXService struct {
	validList *[]string
	randNum   *rand.Rand
	client    *req.Client
}

func NewDeepLXService(vlist *[]string) TranslateService {
	return &DeepLXService{
		validList: vlist,
		randNum:   rand.New(rand.NewSource(time.Now().UnixNano())),
		client:    req.NewClient().SetTimeout(2 * time.Second),
	}
}

func (d *DeepLXService) GetTranslateData(trReq domain.TranslateRequest) domain.TranslateResponse {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	resultChan := make(chan domain.TranslateResponse, 5)

	contextPool := pool.New().WithContext(ctx).WithMaxGoroutines(5)
	for i := 0; i < 5; i++ {
		contextPool.Go(func(ctx context.Context) error {
			url := (*d.validList)[d.randNum.Intn(len(*d.validList))]
			var trResult domain.TranslateResponse
			response, err := d.client.R().
				SetContext(ctx).
				SetBody(trReq).
				SetSuccessResult(&trResult).
				Post(url)

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

	go func() {
		_ = contextPool.Wait()
		if _, ok := <-resultChan; !ok { // 如果通道已经关闭，直接返回
			return
		}
		close(resultChan)
	}()

	select {
	case r := <-resultChan:
		defer cancelFunc()
		return r
	case <-ctx.Done():
		log.Println("all requests failed")
	}
	return domain.TranslateResponse{}
}
