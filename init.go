package main

import (
	"context"
	"deeplx-local/channel"
	"deeplx-local/cron"
	"deeplx-local/pkg"
	"deeplx-local/service"
	"errors"
	"fmt"
	"github.com/imroc/req/v3"
	lop "github.com/samber/lo/parallel"
	"github.com/sourcegraph/conc/pool"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	urlPath     = "url.txt"
	client      = req.NewClient().SetTimeout(3 * time.Second)
	hunterKey   = os.Getenv("hunter_api_key")
	quakeKey    = os.Getenv("360_api_key")
	routePath   = os.Getenv("route")
	scanService service.ScanService
)

// readFile
func readFile(filename string) ([]byte, error) {
	_, err := os.Stat(filename)
	if err != nil {
		// file no exit, create it and return nil
		if errors.Is(err, os.ErrNotExist) {
			log.Println("url.txt is not exist")
			return nil, os.WriteFile(filename, []byte{}, 0600)
		}

		// Other error
		return nil, err
	}

	// file exist, read it
	content, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	return content, nil
}

// getValidURLs 从文件中读取并处理URL
func getValidURLs() []string {
	content, err := readFile(urlPath)
	if err != nil {
		log.Fatal(err)
	}

	var urls []string
	if len(content) == 0 { // 文件为空 去扫描
		log.Println("url.txt is empty")
		s := getScanService()
		scan := s.Scan()
		if len(scan) == 0 {
			log.Fatalln("url.txt is empty and scan failed")
			return nil
		}
		// 处理URL
		urls = processUrls(scan)
	} else {
		urls = strings.Split(string(content), "\n")
		urls = processUrls(urls)
	}
	// 保存处理后的URL
	writeFileReplace(urlPath, urls)

	validChan := make(chan string, len(urls))

	// 并发检查URL可用性
	p := pool.New().WithMaxGoroutines(30)
	for _, url := range urls {
		url := url // 创建一个新的变量，避免闭包中的变量复用
		p.Go(func() {
			if availability, err := pkg.CheckURLAvailability(client, url); err == nil && availability {
				validChan <- url
			}
		})
	}
	p.Wait()
	close(validChan)

	validList := make([]string, 0, len(validChan))
	for url := range validChan {
		validList = append(validList, url)
	}

	log.Printf("available urls count: %d\n", len(validList))

	if len(validList) == 0 {
		log.Fatalln("available urls is empty")
	}
	return validList
}

func processUrls(urls []string) []string {
	// 使用正则表达式处理 URL 后缀
	suffixPattern := regexp.MustCompile(`(?:/translate)?/?$`)
	// 使用正则表达式处理 URL 前缀
	prefixPattern := regexp.MustCompile("^(http|https)://")

	urls = lop.Map(urls, func(url string, _ int) string {
		u := strings.TrimSpace(url)
		u = suffixPattern.ReplaceAllString(u, "/translate")
		if prefixPattern.MatchString(u) {
			return u
		}
		return "http://" + u
	})

	// 去重
	distinctURLs(&urls)
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

	if hunterKey == "" && quakeKey == "" {
		log.Println("未提供有YingTu 或 360的API Key")
		return nil
	}

	var (
		cli      = req.NewClient().SetTimeout(15 * time.Second)
		services []service.ScanService
	)

	if hunterKey != "" {
		services = append(services, service.NewYingTuScanService(cli, hunterKey))
	}

	if quakeKey != "" {
		services = append(services, service.NewQuake360ScanService(cli, quakeKey))
	}

	// 返回组合扫描服务
	return service.NewCombinedScanService(services...)

}

func autoScan() {
	scanService := getScanService()
	if scanService == nil {
		return
	}
	cron.StartTimer(time.Hour*24*2, func() {
		scan := scanService.Scan()
		urls := processUrls(scan)                                   // 处理URL
		writeFileIncremental(urlPath, urls)                         // 增量写入保存
		exec.Command("kill", "-1", strconv.Itoa(os.Getpid())).Run() // 重启
	})
}

// writeFileIncremental 增量写入保存
func writeFileIncremental(path string, urls []string) {
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

// writeFileReplace 全量写入保存
func writeFileReplace(path string, urls []string) {
	// 保存处理后的URL
	os.WriteFile(path, []byte(strings.Join(urls, "\n")), 0600)
}
