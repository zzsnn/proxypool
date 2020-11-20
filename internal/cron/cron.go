package cron

import (
	"github.com/Sansui233/proxypool/config"
	"github.com/Sansui233/proxypool/internal/cache"
	"runtime"

	"github.com/Sansui233/proxypool/internal/app"
	"github.com/jasonlvhit/gocron"
)

func Cron() {
	_ = gocron.Every(config.Config.CronTime).Minutes().Do(crawlTask)
	_ = gocron.Every(config.Config.SpeedTestInterval).Hour().Do(speedTestTask)
	<-gocron.Start()
}

func crawlTask() {
	_ = app.InitConfigAndGetters("")
	app.CrawlGo()
	app.Getters = nil
	runtime.GC()
}

func speedTestTask() {
	pl := cache.GetProxies("proxies")
	app.SpeedTest(pl)
	runtime.GC()
}
