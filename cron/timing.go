package cron

import (
	"deeplx-local/channel"
	"log"
	"time"
)

func StartTimer(t time.Duration, f func()) {
	go func() {
		for {
			now := time.Now()
			next := now.Add(t)
			next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())
			timer := time.NewTimer(next.Sub(now))
			select {
			case <-channel.Quit:
				log.Println("定时任务已退出")
				timer.Stop()
				return
			case <-timer.C:
				f()
			}
		}
	}()
}
