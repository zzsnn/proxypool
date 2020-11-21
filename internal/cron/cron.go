package cron

import (
	"github.com/Sansui233/proxypool/config"
	"github.com/Sansui233/proxypool/internal/cache"
	"github.com/Sansui233/proxypool/pkg/healthcheck"
	"runtime"

	"github.com/Sansui233/proxypool/internal/app"
	"github.com/jasonlvhit/gocron"
)

func Cron() {
	_ = gocron.Every(config.Config.CronInterval).Minutes().Do(crawlTask)
	_ = gocron.Every(config.Config.SpeedTestInterval).Hour().Do(speedTestTask)
	_ = gocron.Every(config.Config.ActiveInterval).Hour().Do(frequentSpeedTestTask)
	<-gocron.Start()
}

func crawlTask() {
	_ = app.InitConfigAndGetters("")
	app.CrawlGo()
	app.Getters = nil
	runtime.GC()
}

func speedTestTask() {
	_ = config.Parse("")
	pl := cache.GetProxies("proxies")
	app.SpeedTest(pl)
	runtime.GC()
}

func frequentSpeedTestTask() {
	_ = config.Parse("")
	pl := cache.GetProxies("proxies")
	pl = healthcheck.PStats.ReqCountThan(config.Config.ActiveFrequency, pl, true)
	app.SpeedTest(pl)
	runtime.GC()
}
