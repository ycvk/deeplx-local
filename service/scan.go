package service

import (
	"deeplx-local/domain"
	"fmt"
	"github.com/imroc/req/v3"
	lop "github.com/samber/lo/parallel"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const Quake360Page = "200"
const YingTuPageSize = "100"

type ScanService interface {
	Scan() []string
}

type YingTuScanService struct {
	client      *req.Client
	searchParam string
	apiKey      string
}

func NewYingTuScanService(client *req.Client, apikey string) ScanService {
	return &YingTuScanService{
		client,
		"KHdlYi5ib2R5PT0iRGVlcEwgRnJlZSBBUEksIERldmVsb3BlZCBieSBzamxsZW8gYW5kIG1pc3N1by4gR28gdG8gL3RyYW5zbGF0ZSB3aXRoIFBPU1QuIGh0dHA6Ly9naXRodWIuY29tL093Ty1OZXR3b3JrL0RlZXBMWCIpJiYoaXAuY291bnRyeT09IuS4reWbvSIp",
		apikey,
	}
}

func (y *YingTuScanService) Scan() []string {
	startDate, endDate := getStartDateAndEndDate()
	address := buildHunterAPIAddress(y.apiKey, y.searchParam, startDate, endDate, 1)
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

	log.Printf("鹰图爬取 deeplx ip 成功，共爬取 %d 条数据，本次查询 %s ，当前 %s \n",
		len(urls), yingtuResp.Data.ConsumeQuota, yingtuResp.Data.RestQuota)

	return urls

}

// buildHunterAPIAddress 构建鹰图查询地址
func buildHunterAPIAddress(apiKey, searchParam, startDate, endDate string, page int) string {
	baseURL := "https://hunter.qianxin.com/openApi/search"
	params := url.Values{}
	params.Set("api-key", apiKey)
	params.Set("search", searchParam)
	params.Set("page", strconv.Itoa(page))
	params.Set("page_size", YingTuPageSize)
	params.Set("is_web", "1")
	params.Set("port_filter", "false")
	params.Set("status_code", "200")
	params.Set("start_time", startDate)
	params.Set("end_time", endDate)

	address := fmt.Sprintf("%s?%s", baseURL, params.Encode())
	return address
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

func (q *Quake360ScanService) GetCredit() bool {
	const address = "https://quake.360.net/api/v3/user/info"
	var userInfo domain.Quake360UserInfoResponse
	response, err := q.client.R().
		SetHeader("X-QuakeToken", q.apiKey).
		SetSuccessResult(&userInfo).
		SetHeader("Content-Type", "application/json").
		Get(address)

	if err != nil {
		log.Println("360用户详情接口请求失败：", err)
		return false
	}

	defer response.Body.Close()

	if userInfo.Code != 0 {
		log.Println("360用户详情接口请求失败", userInfo.Message)
		return false
	}

	// 判断月度免费次数 或 月度剩余积分 是否大于0
	if userInfo.Data.FreeQueryApiCount > 0 || userInfo.Data.MonthRemainingCredit > 0 {
		log.Printf("360用户 %s 详情接口请求成功，月度免费次数剩余：%d，月度剩余积分%d \n",
			userInfo.Data.MobilePhone, userInfo.Data.FreeQueryApiCount, userInfo.Data.MonthRemainingCredit)
		return true
	}

	return false

}

func (q *Quake360ScanService) Scan() []string {
	// 无信用点直接返回 nil
	if hasCredit := q.GetCredit(); !hasCredit {
		return nil
	}

	const address = "https://quake.360.net/api/v3/search/quake_service"
	reqParam := make(map[string]string)
	reqParam["query"] = q.searchParam
	reqParam["size"] = Quake360Page
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

// CombinedScanService 聚合扫描服务
type CombinedScanService struct {
	scanServices []ScanService
}

// NewCombinedScanService 创建聚合扫描服务
func NewCombinedScanService(scanServices ...ScanService) *CombinedScanService {
	return &CombinedScanService{scanServices: scanServices}
}

// Scan 聚合扫描服务
func (c *CombinedScanService) Scan() []string {
	var combinedResults []string
	for _, service := range c.scanServices {
		results := service.Scan()
		combinedResults = append(combinedResults, results...)
		log.Printf("%T Scan Results: %d\n", service, len(results))
	}

	return combinedResults
}
