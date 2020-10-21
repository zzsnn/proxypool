package proxy

import "strings"

/* 基础的接口类为Proxy，Base为Proxy的多态实现与信息补充，Vmess等继承Base，实现更多的多态与信息补充*/
type Base struct {
	Name    string `yaml:"name" json:"name" gorm:"index"`
	Server  string `yaml:"server" json:"server" gorm:"index"`
	Port    int    `yaml:"port" json:"port" gorm:"index"`
	Type    string `yaml:"type" json:"type" gorm:"index"`
	UDP     bool   `yaml:"udp,omitempty" json:"udp,omitempty"`
	Country string `yaml:"country,omitempty" json:"country,omitempty" gorm:"index"`
	// 这个单词的原作者拼写是错误的，但我不想改了，我也没有早点发现这件事，在写where查询老写错，非常的无奈
	Useable bool   `yaml:"useable,omitempty" json:"useable,omitempty" gorm:"index"`
}

// Note: Go只有值传递，必需传入指针才能改变传入的结构体

func (b *Base) TypeName() string {
	if b.Type == "" {
		return "unknown"
	}
	return b.Type
}

func (b *Base) SetName(name string) {
	b.Name = name
}

func (b *Base) SetIP(ip string) {
	b.Server = ip
}

// 返回传入的参数（约等于无效果只是为了接口规范）
func (b *Base) BaseInfo() *Base {
	return b
}

func (b *Base) Clone() Base {
	c := *b
	return c
}

func (b *Base) SetUseable(useable bool) {
	b.Useable = useable
}

func (b *Base) SetCountry(country string) {
	b.Country = country
}

type Proxy interface {
	String() string
	ToClash() string
	ToSurge() string
	Link() string
	Identifier() string
	SetName(name string)
	SetIP(ip string)
	TypeName() string //ss ssr vmess trojan
	BaseInfo() *Base
	Clone() Proxy
	SetUseable(useable bool)
	SetCountry(country string)
}

func ParseProxyFromLink(link string) Proxy {
	var err error
	var data Proxy
	if strings.HasPrefix(link, "ssr://") {
		data, err = ParseSSRLink(link)
	} else if strings.HasPrefix(link, "vmess://") {
		data, err = ParseVmessLink(link)
	} else if strings.HasPrefix(link, "ss://") {
		data, err = ParseSSLink(link)
	} else if strings.HasPrefix(link, "trojan://") {
		data, err = ParseTrojanLink(link)
	}
	if err != nil {
		return nil
	}
	ip, country, err := geoIp.Find(data.BaseInfo().Server)
	if err != nil {
		country = "🏁 ZZ"
	}
	data.SetCountry(country)
	// trojan依赖域名？
	if data.TypeName() != "trojan" {
		data.SetIP(ip)
	}
	return data
}