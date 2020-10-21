package database

import (
	"fmt"
	"github.com/Sansui233/proxypool/pkg/proxy"
	"gorm.io/gorm"
	"log"
	"time"
)

// 设置数据库字段，不使用gnorm.model中的软删除特性，和primary key与unique一起容易导致无法更新
type Proxy struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	proxy.Base
	Link       string
	Identifier string `gorm:"unique"`
}

func InitTables() {
	if DB == nil {
		err := connect()
		if err != nil {
			return
		}
	}
	// Warning: 自动迁移仅仅会创建表，缺少列和索引，并且不会改变现有列的类型或删除未使用的列以保护数据。
	// 如更改表的Column请于数据库中操作
	err := DB.AutoMigrate(&Proxy{})
	if err != nil {
		log.Println("[db/proxy.go] Database migration failed")
		panic(err)
	}
}

func SaveProxyList(pl proxy.ProxyList) {

	DB.Transaction(func(tx *gorm.DB) error {
		// Set All Usable to false
		if err := DB.Model(&Proxy{}).Where("useable = ?", true).Update("useable", "false").Error; err != nil {
			log.Println("[db/proxy.go] Reset Usable to false failed: ", err)
		}
		// Create or Update proxies
		for i := 0; i < pl.Len(); i++ {
			p := Proxy{
				Base:       *pl[i].BaseInfo(),
				Link:       pl[i].Link(),
				Identifier: pl[i].Identifier(),
			}
			p.Useable = true
			if err:= DB.Create(&p).Error; err != nil{
				if uperr := DB.Model(&Proxy{}).Where("identifier = ?",p.Identifier).Update("useable", "true").Error; uperr != nil{
					log.Println("[db/proxy.go] DB Update failed: ",
						"\n\t[db/proxy.go] When Created item: ", err,
						"\n\t[db/proxy.go] When Updated item: ", uperr)
				}
			}
		}
		fmt.Println("Database Updated!")
		return nil
	})
}

// TODO: get出来的数量不够，检查一下什么问题
// Get a proxy list consists of all proxies in database
func GetAllProxies() (proxies proxy.ProxyList) {
	proxies = make(proxy.ProxyList, 0)
	if DB == nil {
		return
	}

	proxiesDB := make([]Proxy, 0)
	DB.Select("link").Find(&proxiesDB)

	for _, proxyDB := range proxiesDB {
		if proxiesDB != nil {
			p, err := proxy.ParseProxyFromLink(proxyDB.Link)
			if err == nil && p != nil {
				proxies = append(proxies, p)
			}
		}
	}
	return
}

// Clear Old unusable proxies more than 1 week
func ClearOldItems(){
	lastWeek := time.Now().Add(-time.Hour*24*7)
	if err := DB.Where("updated_at < ? AND useable = ?", lastWeek, false).Delete(&Proxy{}); err != nil{
		var pl []Proxy
		DB.Where("updated_at < ? AND useable = ?", lastWeek, false).Find(&pl)
		if len(pl) == 0 {
			fmt.Println("Nothing old to clear!")
		}else {
			log.Println("[database test] Delete old item failed ", err)
		}
	}else {
		fmt.Println("Swept old and unusable proxies from database!")
	}
}
