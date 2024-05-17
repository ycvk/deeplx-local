package pkg

import (
	"deeplx-local/domain"
	"github.com/imroc/req/v3"
)

const validResp = "我爱你"

var validReq = domain.TranslateRequest{
	Text:       "I love you",
	SourceLang: "EN",
	TargetLang: "ZH",
}

// CheckURLAvailability 检查URL是否可用
func CheckURLAvailability(client *req.Client, url string) (bool, error) {
	var result domain.TranslateResponse
	response, err := client.R().SetBody(&validReq).SetSuccessResult(&result).Post(url)
	if err != nil {
		return false, err
	}
	defer response.Body.Close()
	return validResp == result.Data, nil
}
