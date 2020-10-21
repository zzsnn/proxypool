package api

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Sansui233/proxypool/config"
	C "github.com/Sansui233/proxypool/internal/cache"
	"github.com/Sansui233/proxypool/pkg/provider"
	"github.com/gin-contrib/cache"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
)

const version = "v0.3.9"

var router *gin.Engine

func setupRouter() {
	gin.SetMode(gin.ReleaseMode)
	router = gin.New()          // 没有任何中间件的路由
	store := persistence.NewInMemoryStore(time.Minute)
	router.Use(gin.Recovery(), cache.SiteCache(store, time.Minute))  // 加上处理panic的中间件，防止遇到panic退出程序
	temp, err := LoadTemplate() // 加载模板，模板源存放于html.go中的类似_assetsHtmlSurgeHtml的变量
	if err != nil {
		panic(err)
	}
	router.SetHTMLTemplate(temp) // 应用模板

	router.Static("/css", "assets/css")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"domain":               config.Config.Domain,
			"getters_count":        C.GettersCount,
			"all_proxies_count":    C.AllProxiesCount,
			"ss_proxies_count":     C.SSProxiesCount,
			"ssr_proxies_count":    C.SSRProxiesCount,
			"vmess_proxies_count":  C.VmessProxiesCount,
			"trojan_proxies_count": C.TrojanProxiesCount,
			"useful_proxies_count": C.UsefullProxiesCount,
			"last_crawl_time":      C.LastCrawlTime,
			"version":              version,
		})
	})

	router.GET("/clash", func(c *gin.Context) {
		c.HTML(http.StatusOK, "clash.html", gin.H{
			"domain": config.Config.Domain,
		})
	})

	router.GET("/surge", func(c *gin.Context) {
		c.HTML(http.StatusOK, "surge.html", gin.H{
			"domain": config.Config.Domain,
		})
	})

	router.GET("/clash/config", func(c *gin.Context) {
		c.HTML(http.StatusOK, "clash-config.yaml", gin.H{
			"domain": config.Config.Domain,
		})
	})
	// 本地测试config
	router.GET("/clash/localconfig", func(c *gin.Context) {
		c.HTML(http.StatusOK, "clash-config-local.yaml", gin.H{})
	})

	router.GET("/surge/config", func(c *gin.Context) {
		c.HTML(http.StatusOK, "surge.conf", gin.H{
			"domain": config.Config.Domain,
		})
	})

	router.GET("/clash/proxies", func(c *gin.Context) {
		proxyTypes := c.DefaultQuery("type", "")
		proxyCountry := c.DefaultQuery("c", "")
		proxyNotCountry := c.DefaultQuery("nc", "")
		text := ""
		if proxyTypes == "" && proxyCountry == "" && proxyNotCountry == "" {
			text = C.GetString("clashproxies")
			if text == "" {
				proxies := C.GetProxies("proxies")
				clash := provider.Clash{
					provider.Base{
						Proxies: &proxies,
					},
				}
				text = clash.Provide() // 根据Query筛选节点
				C.SetString("clashproxies", text)
			}
		} else if proxyTypes == "all" {
			proxies := C.GetProxies("allproxies")
			clash := provider.Clash{
				provider.Base{
					Proxies:    &proxies,
					Types:      proxyTypes,
					Country:    proxyCountry,
					NotCountry: proxyNotCountry,
				},
			}
			text = clash.Provide() // 根据Query筛选节点
		} else {
			proxies := C.GetProxies("proxies")
			clash := provider.Clash{
				provider.Base{
					Proxies:    &proxies,
					Types:      proxyTypes,
					Country:    proxyCountry,
					NotCountry: proxyNotCountry,
				},
			}
			text = clash.Provide() // 根据Query筛选节点
		}
		c.String(200, text)
	})
	router.GET("/surge/proxies", func(c *gin.Context) {
		proxyTypes := c.DefaultQuery("type", "")
		proxyCountry := c.DefaultQuery("c", "")
		proxyNotCountry := c.DefaultQuery("nc", "")
		text := ""
		if proxyTypes == "" && proxyCountry == "" && proxyNotCountry == "" {
			text = C.GetString("surgeproxies")
			if text == "" {
				proxies := C.GetProxies("proxies")
				surge := provider.Surge{
					provider.Base{
						Proxies: &proxies,
					},
				}
				text = surge.Provide()
				C.SetString("surgeproxies", text)
			}
		} else if proxyTypes == "all" {
			proxies := C.GetProxies("allproxies")
			surge := provider.Surge{
				provider.Base{
					Proxies:    &proxies,
					Types:      proxyTypes,
					Country:    proxyCountry,
					NotCountry: proxyNotCountry,
				},
			}
			text = surge.Provide()
		} else {
			proxies := C.GetProxies("proxies")
			surge := provider.Surge{
				provider.Base{
					Proxies:    &proxies,
					Types:      proxyTypes,
					Country:    proxyCountry,
					NotCountry: proxyNotCountry,
				},
			}
			text = surge.Provide()
		}
		c.String(200, text)
	})

	router.GET("/ss/sub", func(c *gin.Context) {
		proxies := C.GetProxies("proxies")
		ssSub := provider.SSSub{
			provider.Base{
				Proxies: &proxies,
				Types:   "ss",
			},
		}
		c.String(200, ssSub.Provide())
	})
	router.GET("/ssr/sub", func(c *gin.Context) {
		proxies := C.GetProxies("proxies")
		ssrSub := provider.SSRSub{
			provider.Base{
				Proxies: &proxies,
				Types:   "ssr",
			},
		}
		c.String(200, ssrSub.Provide())
	})
	router.GET("/vmess/sub", func(c *gin.Context) {
		proxies := C.GetProxies("proxies")
		vmessSub := provider.VmessSub{
			provider.Base{
				Proxies: &proxies,
				Types:   "vmess",
			},
		}
		c.String(200, vmessSub.Provide())
	})
	router.GET("/sip002/sub", func(c *gin.Context) {
		proxies := C.GetProxies("proxies")
		sip002Sub := provider.SIP002Sub{
			provider.Base{
				Proxies: &proxies,
				Types:   "ss",
			},
		}
		c.String(200, sip002Sub.Provide())
	})
	router.GET("/link/:id", func(c *gin.Context) {
		idx := c.Param("id")
		proxies := C.GetProxies("allproxies")
		id, err := strconv.Atoi(idx)
		if err != nil {
			c.String(500, err.Error())
		}
		if id >= proxies.Len() || id < 0 {
			c.String(500, "id out of range")
		}
		c.String(200, proxies[id].Link())
	})
}

func Run() {
	setupRouter()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	// Run on this server
	err := router.Run(":" + port)
	if err != nil {
		log.Fatal("[router.go] Remote server starting failed")
	}

}

// 返回页面templates
func LoadTemplate() (t *template.Template, err error) {
	/* 使用本地模板文件 */
	filePaths, err := GetAllFilePaths("assets" + string(os.PathSeparator) + "html")
	if err != nil {
		log.Fatal("[router.go] Fail to load web templates: ", err)
		return nil, err;
	}
	for _, filePath := range filePaths {
		t, _ = t.ParseFiles(filePath) // Parsefile后的模板无路径前缀
		if err != nil {
			log.Panic("[router.go] ", err)
		}
	}
	return t, nil
}

// unix directory format
// TODO: This function shouldn't be here
func GetAllFilePaths(pathname string) (filenames []string,err error) {
	rd, err := ioutil.ReadDir(pathname)
	for _, fi := range rd {
		if fi.IsDir() {
			GetAllFilePaths(pathname + string(os.PathSeparator) + fi.Name())
		} else {
			filename := pathname + string(os.PathSeparator) + fi.Name()
			filenames = append(filenames, filename)
		}
	}
	return filenames,err
}