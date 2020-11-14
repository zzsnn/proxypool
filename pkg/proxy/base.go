package proxy

import (
	"errors"
	"strings"
)

/* Base implements interface Proxy. It's the basic proxy struct. Vmess etc extends Base*/
type Base struct {
	Name    string `yaml:"name" json:"name" gorm:"index"`
	Server  string `yaml:"server" json:"server" gorm:"index"`
	Port    int    `yaml:"port" json:"port" gorm:"index"`
	Type    string `yaml:"type" json:"type" gorm:"index"`
	UDP     bool   `yaml:"udp,omitempty" json:"udp,omitempty"`
	Country string `yaml:"country,omitempty" json:"country,omitempty" gorm:"index"`
	// 这个单词的原作者拼写是错误的，但我不想改了，我也没有早点发现这件事，在写where查询老写错，非常的无奈
	Useable bool `yaml:"useable,omitempty" json:"useable,omitempty" gorm:"index"`
}

// Note: Go只有值传递，必需传入指针才能改变传入的结构体

// TypeName() Get specific proxy type
func (b *Base) TypeName() string {
	if b.Type == "" {
		return "unknown"
	}
	return b.Type
}

// SetName() to a proxy
func (b *Base) SetName(name string) {
	b.Name = name
}

func (b *Base) AddToName(name string) {
	b.Name = b.Name + name
}

// SetIP() to a proxy
func (b *Base) SetIP(ip string) {
	b.Server = ip
}

// BaseInfo() get basic info struct of a proxy
func (b *Base) BaseInfo() *Base {
	return b
}

// Clone() returns a new basic proxy
func (b *Base) Clone() Base {
	c := *b
	return c
}

// SetUseable() set Base info "Useable" (true or false)
func (b *Base) SetUseable(useable bool) {
	b.Useable = useable
}

// SetUseable() set Base info "Country" (string)
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
	AddToName(name string)
	SetIP(ip string)
	TypeName() string //ss ssr vmess trojan
	BaseInfo() *Base
	Clone() Proxy
	SetUseable(useable bool)
	SetCountry(country string)
}

func ParseProxyFromLink(link string) (p Proxy, err error) {
	if strings.HasPrefix(link, "ssr://") {
		p, err = ParseSSRLink(link)
	} else if strings.HasPrefix(link, "vmess://") {
		p, err = ParseVmessLink(link)
	} else if strings.HasPrefix(link, "ss://") {
		p, err = ParseSSLink(link)
	} else if strings.HasPrefix(link, "trojan://") {
		p, err = ParseTrojanLink(link)
	}
	if err != nil || p == nil {
		return nil, errors.New("link parse failed")
	}
	_, country, err := geoIp.Find(p.BaseInfo().Server) // IP库不准
	if err != nil {
		country = "🏁 ZZ"
	}
	p.SetCountry(country)
	// trojan依赖域名？<-这是啥?不管什么情况感觉都不应该替换域名为IP（主要是IP库的质量和节点质量不该挂钩）
	//if p.TypeName() != "trojan" {
	//	p.SetIP(ip)
	//}
	return
}
