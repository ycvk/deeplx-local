package main

import (
	"deeplx-local/service"
	"deeplx-local/web"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/ycvk/endless"
	"log"
	"net/http"
	"os"
	"time"
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

	go func() {
		for {
			time.Sleep(2 * time.Second)
			log.Println(os.Getpid())
		}
	}()

	autoScan()
	// 启动服务
	go func() {
		if err := endless.ListenAndServe("0.0.0.0:62155", r); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("web服务启动失败: ", err)
		}
	}()

	select {}
}
