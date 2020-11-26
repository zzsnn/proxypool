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
	proxies = uniqAppend(proxies, db_proxies)

	go func() {
		wg.Wait()
		close(pc)
	}() // Note: 为何并发？可以一边抓取一边读取而非抓完再读
	for p := range pc { // Note: pc关闭后不能发送数据可以读取剩余数据
		if p != nil {
			proxies = uniqAppendProxy(proxies, p)
		}
	}

	// 节点衍生并去重 -> 去重在上面做了，衍生的必要性不太明白
	//proxies = proxies.Deduplication().Derive()
	log.Println("CrawlGo unique proxy count:", len(proxies))

	// 去除Clash(windows)不支持的节点
	proxies = provider.Clash{
		provider.Base{
			Proxies: &proxies,
		},
	}.CleanProxies()
	log.Println("CrawlGo clash supported proxy count:", len(proxies))

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

	// 节点可用性检测，使用batchsize不能降低内存占用，只是为了看性能
	log.Println("Now proceed proxy health check...")
	b := 1000
	round := len(proxies) / b
	okproxies := make(proxy.ProxyList, 0)
	for i := 0; i < round; i++ {
		okproxies = append(okproxies, healthcheck.CleanBadProxiesWithGrpool(proxies[i*b:(i+1)*b])...)
		log.Println("Checking round:", i)
	}
	okproxies = append(okproxies, healthcheck.CleanBadProxiesWithGrpool(proxies[round*b:])...)
	proxies = okproxies

	proxies = healthcheck.CleanBadProxiesWithGrpool(proxies)
	log.Println("CrawlGo clash usable proxy count:", len(proxies))

	// 重命名节点名称为类似US_01的格式，并按国家排序
	proxies.NameSetCounrty().Sort().NameAddIndex()
	//proxies.NameReIndex()
	log.Println("Proxy rename DONE!")

	// 可用节点存储
	cache.SetProxies("proxies", proxies)
	cache.UsefullProxiesCount = proxies.Len()
	database.SaveProxyList(proxies)
	database.ClearOldItems()

	log.Println("Usablility checking done. Open", config.Config.Domain+":"+config.Config.Port, "to check")

	// 测速
	speedTestNew(proxies)
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

// Speed test for new proxies
func speedTestNew(proxies proxy.ProxyList) {
	// speed check
	if config.Config.SpeedTest {
		cache.IsSpeedTest = "已开启"
		if config.Config.Timeout > 0 {
			healthcheck.SpeedTimeout = time.Second * time.Duration(config.Config.Timeout)
		}
		healthcheck.SpeedTestNew(proxies, config.Config.Connection)
	} else {
		cache.IsSpeedTest = "未开启"
	}
}

// Speed test for all proxies in proxy.ProxyList
func SpeedTest(proxies proxy.ProxyList) {
	// speed check
	if config.Config.SpeedTest {
		cache.IsSpeedTest = "已开启"
		if config.Config.Timeout > 0 {
			healthcheck.SpeedTimeout = time.Second * time.Duration(config.Config.Timeout)
		}
		healthcheck.SpeedTestAll(proxies, config.Config.Connection)
	} else {
		cache.IsSpeedTest = "未开启"
	}
}

// Append unique new proxies to old proxy.ProxyList
func uniqAppend(pl proxy.ProxyList, new proxy.ProxyList) proxy.ProxyList {
	if len(new) == 0 {
		return pl
	}
	if len(pl) == 0 {
		return new
	}
	for _, p := range new {
		for i, _ := range pl {
			if pl[i].Identifier() == p.Identifier() {
				continue
			}
		}
		pl = append(pl, p)
	}
	return pl
}

// Append unique new proxies to old proxy.ProxyList
func uniqAppendProxy(pl proxy.ProxyList, new proxy.Proxy) proxy.ProxyList {
	if len(pl) == 0 {
		pl = append(pl, new)
		return pl
	}
	for i, _ := range pl {
		if pl[i].Identifier() == new.Identifier() {
			return pl
		}
	}
	pl = append(pl, new)
	return pl
}
