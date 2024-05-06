package main

import (
	"context"
	"deeplx-local/channel"
	"deeplx-local/cron"
	"deeplx-local/domain"
	"deeplx-local/service"
	"fmt"
	"github.com/imroc/req/v3"
	"github.com/sourcegraph/conc/pool"
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
	client   = req.NewClient().SetTimeout(3 * time.Second)
	validReq = domain.TranslateRequest{
		Text:       "I love you",
		SourceLang: "EN",
		TargetLang: "ZH",
	}
	hunterKey   = os.Getenv("hunter_api_key")
	quakeKey    = os.Getenv("360_api_key")
	scanService service.ScanService
)

// getValidURLs 从文件中读取并处理URL
func getValidURLs() []string {
	content, err := os.ReadFile("url.txt")
	if err != nil {
		log.Fatal(err)
	}

	var urls []string
	if len(content) == 0 {
		log.Println("url.txt is empty")
		s := getScanService()
		scan := s.Scan()
		if len(scan) == 0 {
			log.Fatalln("url.txt is empty and scan failed")
			return nil
		}
		urls = processUrls(scan)
	} else {
		urls = strings.Split(string(content), "\n")
	}
	// 处理URL
	urls = processUrls(urls)

	validList := make([]string, 0, len(urls))

	// 并发检查URL可用性
	p := pool.New().WithMaxGoroutines(30)
	for _, url := range urls {
		p.Go(func() {
			if availability, err := checkURLAvailability(url); err == nil && availability {
				validList = append(validList, url)
			}
		})
	}
	p.Wait()

	log.Printf("available urls count: %d\n", len(validList))
	//os.WriteFile("url.txt", []byte(strings.Join(validList, "\n")), 0600) // 保存
	return validList
}

func processUrls(urls []string) []string {
	for i := range urls {
		urls[i] = strings.TrimSpace(urls[i])
		if !strings.HasSuffix(urls[i], "/translate") {
			if strings.HasSuffix(urls[i], "/") {
				urls[i] += "translate"
			} else {
				urls[i] += "/translate"
			}
		}
		if !strings.HasPrefix(urls[i], "http") {
			urls[i] = "http://" + urls[i]
		}
	}
	// 去重
	distinctURLs(&urls)
	// 保存处理后的URL
	os.WriteFile("url.txt", []byte(strings.Join(urls, "\n")), 0600)
	return urls
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
		if sig == syscall.SIGHUP {
			channel.Restart <- sig
			quit <- true
		} else {
			quit <- true
		}
	}()
	<-quit
	fmt.Println("服务 PID 为: ", os.Getpid())
	fmt.Println("服务已退出")
	// 查杀
	exec.Command("killall", "main", strconv.Itoa(os.Getpid())).Run()
	// 自杀
	exec.Command("kill", "-9", strconv.Itoa(os.Getpid())).Run()
}

func exitV1() {
	osSig := make(chan os.Signal, 1)
	quit := make(chan bool, 1)

	signal.Notify(osSig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		sig := <-osSig
		fmt.Println("收到退出信号: ", sig)
		channel.Quit <- sig
		quit <- true
	}()
	<-quit
}

func getScanService() service.ScanService {
	if scanService != nil {
		return scanService
	}
	var cli = req.NewClient().SetTimeout(15 * time.Second)
	if hunterKey != "" {
		return service.NewYingTuScanService(cli, hunterKey)
	}
	if quakeKey != "" {
		return service.NewQuake360ScanService(cli, quakeKey)
	}
	log.Println("未找到有效的API Key")
	return nil
}

func autoScan() {
	scanService := getScanService()
	if scanService == nil {
		return
	}
	cron.StartTimer(time.Hour*24*2, func() {
		scan := scanService.Scan()
		distinctURLs(&scan)                                         // 去重
		urls := processUrls(scan)                                   // 处理URL
		writeFile("url.txt", urls)                                  // 保存
		exec.Command("kill", "-1", strconv.Itoa(os.Getpid())).Run() // 重启
	})
}

func writeFile(path string, urls []string) {
	// 打开文件，如果文件不存在则创建
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Println("文件打开失败", err)
		return
	}
	defer file.Close()

	// 要写入的内容,增量保存
	text := "\n" + strings.Join(urls, "\n")

	// 写入文件
	if _, err = file.WriteString(text); err != nil {
		log.Println("写入文件失败", err)
	}
}
