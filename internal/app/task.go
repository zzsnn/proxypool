package app

import (
	"fmt"
	"github.com/Sansui233/proxypool/config"
	"github.com/Sansui233/proxypool/pkg/healthcheck"
	"log"
	"sync"
	"time"

	"github.com/Sansui233/proxypool/internal/cache"
	"github.com/Sansui233/proxypool/internal/database"
	"github.com/Sansui233/proxypool/pkg/provider"
	"github.com/Sansui233/proxypool/pkg/proxy"
)

var location, _ = time.LoadLocation("PRC")

func CrawlGo() {
	wg := &sync.WaitGroup{}
	var pc = make(chan proxy.Proxy)
	for _, g := range Getters {
		wg.Add(1)
		// 并发执行抓取node并存到pc
		go g.Get2Chan(pc, wg)
	}
	proxies := cache.GetProxies("allproxies")
	db_proxies := database.GetAllProxies()
	// Show last time result when launch
	if proxies == nil && db_proxies != nil {
		cache.SetProxies("proxies", db_proxies)
		cache.LastCrawlTime = "抓取中，已载入上次数据库数据"
		fmt.Println("Database: loaded")
	}
	proxies = append(proxies, db_proxies...)

	go func() {
		wg.Wait()
		close(pc)
	}() // Note: 为何并发？可以一边抓取一边读取而非抓完再读
	for node := range pc { // Note: pc关闭后不能发送数据可以读取剩余数据
		if node != nil {
			proxies = append(proxies, node)
		}
	}

	// 节点衍生并去重
	proxies = proxies.Deduplication().Derive()
	log.Println("CrawlGo unique node count:", len(proxies))
	// 去除Clash(windows)不支持的节点
	proxies = provider.Clash{
		provider.Base{
			Proxies: &proxies,
		},
	}.CleanProxies()
	log.Println("CrawlGo clash supported node count:", len(proxies))
	// 重命名节点名称为类似US_01的格式，并按国家排序
	proxies.NameSetCounrty().Sort().NameAddIndex() //由于原作停更，暂不加.NameAddTG()，如被认为有版权问题请告知
	log.Println("Proxy rename DONE!")

	cache.SetProxies("allproxies", proxies)
	cache.AllProxiesCount = proxies.Len()
	log.Println("AllProxiesCount:", cache.AllProxiesCount)
	cache.SSProxiesCount = proxies.TypeLen("ss")
	log.Println("SSProxiesCount:", cache.SSProxiesCount)
	cache.SSRProxiesCount = proxies.TypeLen("ssr")
	log.Println("SSRProxiesCount:", cache.SSRProxiesCount)
	cache.VmessProxiesCount = proxies.TypeLen("vmess")
	log.Println("VmessProxiesCount:", cache.VmessProxiesCount)
	cache.TrojanProxiesCount = proxies.TypeLen("trojan")
	log.Println("TrojanProxiesCount:", cache.TrojanProxiesCount)
	cache.LastCrawlTime = time.Now().In(location).Format("2006-01-02 15:04:05")

	// 节点可用性检测
	log.Println("Now proceed proxy health check...")
	proxies = healthcheck.CleanBadProxiesWithGrpool(proxies)
	log.Println("CrawlGo clash usable node count:", len(proxies))
	proxies.NameReIndex() //由于原作停更，暂不加.NameAddTG()，如被认为有版权问题请告知

	// 可用节点存储
	cache.SetProxies("proxies", proxies)
	cache.UsefullProxiesCount = proxies.Len()
	database.SaveProxyList(proxies)
	database.ClearOldItems()

	log.Println("Usablility checking done. Open", config.Config.Domain+":"+config.Config.Port, "to check")

	// 测速
	speedTestNew(proxies)
}

// Speed test for new proxies
func speedTestNew(proxies proxy.ProxyList) {
	// speed check
	if config.Config.SpeedTest {
		if config.Config.Timeout > 0 {
			healthcheck.SpeedTimeout = time.Second * time.Duration(config.Config.Timeout)
		}
		healthcheck.SpeedTestNew(proxies, config.Config.Connection)
	}
	cache.SetString("clashproxies", provider.Clash{
		provider.Base{
			Proxies: &proxies,
		},
	}.Provide()) // update static string provider
	cache.SetString("surgeproxies", provider.Surge{
		provider.Base{
			Proxies: &proxies,
		},
	}.Provide())
}

// Speed test for all proxies
func SpeedTest(proxies proxy.ProxyList) {
	// speed check
	if config.Config.SpeedTest {
		if config.Config.Timeout > 0 {
			healthcheck.SpeedTimeout = time.Second * time.Duration(config.Config.Timeout)
		}
		healthcheck.SpeedTestAll(proxies, config.Config.Connection)
	}
	cache.SetString("clashproxies", provider.Clash{
		provider.Base{
			Proxies: &proxies,
		},
	}.Provide()) // update static string provider
	cache.SetString("surgeproxies", provider.Surge{
		provider.Base{
			Proxies: &proxies,
		},
	}.Provide())
}
