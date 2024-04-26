package service

import (
	"deeplx-local/domain"
	"fmt"
	"github.com/imroc/req/v3"
	lop "github.com/samber/lo/parallel"
	"log"
	"strings"
	"time"
)

type ScanService interface {
	Scan() []string
}

type YingTuScanService struct {
	client      *req.Client
	searchParam string
	apiKey      string
}

func NewYingTuScanService(client *req.Client, apikey string) ScanService {
	return &YingTuScanService{client: client, apiKey: apikey, searchParam: "KHdlYi5ib2R5PT0iRGVlcEwgRnJlZSBBUEksIERldmVsb3BlZCBieSBzamxsZW8gYW5kIG1pc3N1by4gR28gdG8gL3RyYW5zbGF0ZSB3aXRoIFBPU1QuIGh0dHA6Ly9naXRodWIuY29tL093Ty1OZXR3b3JrL0RlZXBMWCIpJiYoaXAuY291bnRyeT09IuS4reWbvSIp"}
}

func (y *YingTuScanService) Scan() []string {
	startDate, endDate := getStartDateAndEndDate()
	address := fmt.Sprintf("https://hunter.qianxin.com/openApi/search?api-key=%s&search=%s&page=1&page_size=200&is_web=1&port_filter=false&status_code=200&start_time=%s&end_time=%s",
		y.apiKey, y.searchParam, startDate, endDate)
	var yingtuResp domain.YingTuResponse

	response, err := y.client.
		EnableInsecureSkipVerify().
		R().
		SetSuccessResult(&yingtuResp).
		Get(address)
	if err != nil {
		log.Println("鹰图爬取 deeplx ip 失败:", err)
		return nil
	}
	defer response.Body.Close()

	if yingtuResp.Code != 200 || yingtuResp.Data.Total == 0 {
		log.Println("鹰图爬取 deeplx ip 失败")
		return nil

	}
	urls := lop.Map(yingtuResp.Data.Arr, func(item domain.YingTuResponseArr, _ int) string {
		return item.Url
	})

	return urls

}

func getStartDateAndEndDate() (string, string) {
	now := time.Now()
	year, month, day := now.Date()
	endDate := fmt.Sprintf("%d-%02d-%02d", year, month, day)
	// 上一月
	lastMonth := now.AddDate(0, 0, -30)
	year, month, day = lastMonth.Date()
	startDate := fmt.Sprintf("%d-%02d-%02d", year, month, day)
	return startDate, endDate
}

type Quake360ScanService struct {
	client      *req.Client
	apiKey      string
	searchParam string
}

func NewQuake360ScanService(client *req.Client, apiKey string) ScanService {
	return &Quake360ScanService{client: client, apiKey: apiKey, searchParam: "response:\"DeepL Free API, Developed by sjlleo and missuo. Go to /translate with POST. http://github.com/OwO-Network/DeepLX\" AND country: \"China\""}
}

func (q *Quake360ScanService) Scan() []string {
	const address = "https://quake.360.net/api/v3/search/quake_service"
	reqParam := make(map[string]string)
	reqParam["query"] = q.searchParam
	reqParam["size"] = "100"
	reqParam["start"] = "0"
	var quakeResp domain.Quake360Response
	response, err := q.client.R().
		SetHeader("X-QuakeToken", q.apiKey).
		SetHeader("Content-Type", "application/json").
		SetBody(reqParam).
		SetSuccessResult(&quakeResp).
		Post(address)

	if err != nil {
		log.Println("360爬取 deeplx ip 失败:", err)
		return nil
	}
	defer response.Body.Close()
	if quakeResp.Code != 0 {
		log.Println("360爬取 deeplx ip 失败", quakeResp.Message)
		return nil
	}

	urls := lop.Map(quakeResp.Data, func(item domain.Quake360ResponseData, _ int) string {
		return getQuakeScanUrl(item)
	})
	return urls
}

func getQuakeScanUrl(data domain.Quake360ResponseData) string {
	if data.Domain != "" {
		return data.Domain
	}
	split := strings.Split(data.Id, "_") // 形如 116.204.90.243_2001_tcp
	// 去掉最后一个元素
	split = split[:len(split)-1]
	// 拼接ip和端口
	return strings.Join(split, ":")
}
