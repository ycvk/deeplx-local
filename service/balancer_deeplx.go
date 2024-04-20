package service

import (
	"deeplx-local/domain"
	"log"
	"sync"
	"time"
)

type Server struct {
	URL           string
	Weight        int
	CurrentWeight int
	ResponseTime  time.Duration
}

type LoadBalancer struct {
	Servers       []*Server
	mutex         sync.Mutex
	deepLXService *DeepLXService
}

// NewLoadBalancer 负载均衡 装饰器模式包了一层service
func NewLoadBalancer(service *DeepLXService) TranslateService {
	servers := make([]*Server, len(*service.validList))
	for i, url := range *service.validList {
		servers[i] = &Server{URL: url, Weight: 1, CurrentWeight: 1}
	}
	return &LoadBalancer{Servers: servers, deepLXService: service}
}

func (lb *LoadBalancer) GetTranslateData(trReq domain.TranslateRequest) domain.TranslateResponse {
	count := 0
	trResult := domain.TranslateResponse{}

	for {
		count++
		if count == 10 {
			break
		}

		server := lb.getServer()
		start := time.Now()
		response, err := lb.deepLXService.client.R().SetBody(trReq).SetSuccessResult(&trResult).Post(server.URL)
		elapsed := time.Since(start)
		go lb.updateResponseTime(server, elapsed)

		if err != nil {
			log.Printf("error: %s\n", err)
			continue
		}
		response.Body.Close()

		if trResult.Code == 200 {
			return trResult
		}
	}
	return trResult
}

func (lb *LoadBalancer) getServer() *Server {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	var bestServer *Server
	total := 0

	for _, server := range lb.Servers {
		server.CurrentWeight += server.Weight
		total += server.Weight

		if bestServer == nil || server.CurrentWeight > bestServer.CurrentWeight {
			bestServer = server
		}
	}

	if bestServer != nil {
		bestServer.CurrentWeight -= total
	}

	return bestServer
}

func (lb *LoadBalancer) updateResponseTime(server *Server, responseTime time.Duration) {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	server.ResponseTime = responseTime
	server.Weight = int(time.Second / (responseTime + 1))
}
