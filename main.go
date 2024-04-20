package main

import (
	"deeplx-local/service"
	"deeplx-local/web"
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func main() {
	// 从文件中读取并处理URL
	urLs := getValidURLs()

	// 注册服务
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	lxService := service.NewDeepLXService(&urLs)
	balancerService := service.NewLoadBalancer(lxService.(*service.DeepLXService))
	lxHandler := web.NewDeepLXHandler(balancerService)
	lxHandler.RegisterRoutes(r)

	// 启动服务
	server := &http.Server{
		Addr:    "0.0.0.0:62155",
		Handler: r,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Println("web服务启动失败: ", err)
		}
	}()

	// 监听退出
	exit(server)
}
