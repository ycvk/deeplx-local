package service

import (
	"deeplx-local/domain"
	"github.com/imroc/req/v3"
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
	count := 0
	for {
		url := (*d.validList)[d.randNum.Intn(len(*d.validList))]
		count++
		if count == 10 {
			break
		}

		var trResult domain.TranslateResponse
		response, err := d.client.R().SetBody(trReq).SetSuccessResult(&trResult).Post(url)
		if err != nil {
			log.Printf("error: %s\n", err)
			continue
		}
		response.Body.Close()

		if trResult.Code == 200 {
			return trResult
		}
	}
	return domain.TranslateResponse{}
}
