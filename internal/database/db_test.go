package database

import (
	"fmt"
	"github.com/Sansui233/proxypool/pkg/proxy"
	"log"
	"testing"
	"time"
)

func TestConnect(t *testing.T) {
	//t.SkipNow()
	//connect()
	InitTables()
}

func TestGetAllProxies(t *testing.T) {
	connect()
	proxies := GetAllProxies();
	fmt.Println(proxies.Len())
	fmt.Println(proxies[0])
}

// Note: set work dir to root in test config
func TestUpdateViaGIN(t *testing.T){
	connect()
	if DB == nil {
		return
	}
	// get a proxy
	var pDB Proxy
	DB.Select("link").First(&pDB)
	fmt.Println(pDB)
	// parse
	proxy.InitGeoIpDB()
	p,_ := proxy.ParseProxyFromLink(pDB.Link)
	// construct
	pDBnew := Proxy{
		Base:	*p.BaseInfo(),
		Link:	p.Link(),
		Identifier: p.Identifier(),
	}
	pDBnew.Useable=true
	fmt.Println("NEW to save: ", pDBnew)
	// try create
	if  err := DB.Create(&pDBnew).Error; err !=nil {
		log.Println("\n\t\t[DB test] Create failed: ",err, "\n\t\t[DB test] Trying Update")
		//try Update
		result := DB.Model(Proxy{}).Where("identifier = ?",pDBnew.Identifier).Updates(&Proxy{
			Base:	proxy.Base{Useable: false, Name: "🇦🇱 AL_01"},
		})
		if result.Error != nil {
			log.Fatal("[DB test] UPDATE failed: ",err)
		}else {
			log.Println("Update pass")
		}
	}
}

func TestSaveProxyList(t *testing.T) {
	connect()
	if err := DB.Model(&Proxy{}).Where("useable = ?", true).Update("useable", "false").Error; err != nil {
		log.Println("[db/proxy.go] Reset Usable to false failed: ", err)
	}
}

func TestDeleteProxyList(t *testing.T) {
	connect()
	if err := DB.Delete(&Proxy{},"id > ?",1); err != nil{
		log.Print("Delete failed", err)
	}
}

func TestClearOldItems(t *testing.T) {
	connect()
	timepoint := time.Now().Add(-time.Hour*24*3)
	var count int64
	//var pl []Proxy
	DB.Model(&Proxy{}).Where("created_at < ? AND useable = ?", timepoint, false).Count(&count)
	//DB.Where("created_at < ? AND useable = ?", timepoint, false).Find(&pl)
	fmt.Println(count)
	//fmt.Println(len(pl))

	//ClearOldItems()
}
