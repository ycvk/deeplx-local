package main

import (
	"context"
	"deeplx-local/domain"
	"fmt"
	"github.com/imroc/req/v3"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	client   = req.NewClient().SetTimeout(time.Second)
	validReq = domain.TranslateRequest{
		Text:       "I love you",
		SourceLang: "EN",
		TargetLang: "ZH",
	}
)

// getValidURLs 从文件中读取并处理URL
func getValidURLs() []string {
	content, err := os.ReadFile("url.txt")
	if err != nil {
		log.Fatal(err)
	}

	urls := strings.Split(string(content), "\n")
	for i := range urls {
		if !strings.HasSuffix(urls[i], "/translate") {
			urls[i] += "/translate"
		}
		if !strings.HasPrefix(urls[i], "http") {
			urls[i] = "http://" + urls[i]
		}
	}
	// 去重
	distinctURLs(&urls)
	// 保存处理后的URL
	os.WriteFile("url.txt", []byte(strings.Join(urls, "\n")), 0600)

	eg := errgroup.Group{}
	validList := make([]string, 0, len(urls))

	// 并发检查URL可用性
	for _, url := range urls {
		eg.Go(func() error {
			if availability, err := checkURLAvailability(url); err == nil && availability {
				validList = append(validList, url)
			}
			return err
		})
	}
	if err := eg.Wait(); err != nil {
		log.Printf("error: %s\n", err)
	}

	log.Printf("available urls count: %d\n", len(validList))
	return validList
}

// distinctURLs 去重
func distinctURLs(urls *[]string) {
	if len(*urls) == 0 {
		return
	}
	hashset := make(map[string]struct{})
	for i := 0; i < len(*urls); i++ {
		if _, ok := hashset[(*urls)[i]]; ok {
			copy((*urls)[i:], (*urls)[i+1:])
			*urls = (*urls)[:len(*urls)-1]
			i--
		} else {
			hashset[(*urls)[i]] = struct{}{}
		}
	}
}

// checkURLAvailability 检查URL是否可用
func checkURLAvailability(url string) (bool, error) {
	var result domain.TranslateResponse
	response, err := client.R().SetBody(&validReq).SetSuccessResult(&result).Post(url)
	if err != nil {
		//log.Printf("error: url:[%s] %s\n", url, err)
		return false, err
	}
	defer response.Body.Close()
	return "我爱你" == result.Data, nil
}

// 监听退出
func exit(engine *http.Server) {
	osSig := make(chan os.Signal, 1)
	quit := make(chan bool, 1)

	signal.Notify(osSig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		sig := <-osSig
		fmt.Println("收到退出信号: ", sig)
		// 退出web服务
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := engine.Shutdown(ctx); err != nil {
			fmt.Println("web服务退出失败: ", err)
		}
		quit <- true
	}()
	<-quit
	fmt.Println("服务 PID 为: ", os.Getpid())
	fmt.Println("服务已退出")
	// 查杀
	exec.Command("killall", "main", strconv.Itoa(os.Getpid())).Run()
	// 自杀
	exec.Command("kill", "-9", strconv.Itoa(os.Getpid())).Run()
}
