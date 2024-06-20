package main

import (
	"deeplx-local/service"
	"deeplx-local/web"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/ycvk/endless"
	"log"
	"net/http"
)

func main() {
	initServer() // 初始化服务
	autoScan()   // 自动扫描
	exitV1()     // 监听退出
	select {}
}

func initServer() {
	// 从文件中读取并处理URL
	urLs := getValidURLs()

	// 注册服务
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	balancerService := service.NewLoadBalancer(&urLs)
	lxHandler := web.NewDeepLXHandler(balancerService, routePath)
	lxHandler.RegisterRoutes(r)

	go func() {
		if err := endless.ListenAndServe("0.0.0.0:62155", r); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("web服务启动失败: ", err)
		}
	}()
}
