package provider

import (
	"fmt"
	"github.com/Sansui233/proxypool/pkg/healthcheck"
	"strings"

	"github.com/Sansui233/proxypool/pkg/proxy"
)

type Provider interface {
	Provide() string
}

type Base struct {
	Proxies    *proxy.ProxyList `yaml:"proxies"`
	Types      string           `yaml:"type"`
	Country    string           `yaml:"country"`
	NotCountry string           `yaml:"not_country"`
	Speed      float64          `yaml:"speed"`
}

// 根据子类的的Provide()传入的信息筛选节点，结果会改变传入的proxylist。
func (b *Base) preFilter() {
	proxies := make(proxy.ProxyList, 0)

	needFilterType := true
	needFilterCountry := true
	needFilterNotCountry := true
	needFilterSpeed := true
	if b.Types == "" || b.Types == "all" {
		needFilterType = false
	}
	if b.Country == "" || b.Country == "all" {
		needFilterCountry = false
	}
	if b.NotCountry == "" {
		needFilterNotCountry = false
	}
	if b.Speed < 0 {
		needFilterSpeed = false
	}
	types := strings.Split(b.Types, ",")
	countries := strings.Split(b.Country, ",")
	notCountries := strings.Split(b.NotCountry, ",")

	bProxies := *b.Proxies
	for _, p := range bProxies {
		if needFilterType {
			typeOk := false
			for _, t := range types {
				if p.TypeName() == t {
					typeOk = true
					break
				}
			}
			if !typeOk {
				goto exclude
			}
		}

		if needFilterNotCountry {
			for _, c := range notCountries {
				if strings.Contains(p.BaseInfo().Name, c) {
					goto exclude
				}
			}
		}

		if needFilterCountry {
			countryOk := false
			for _, c := range countries {
				if strings.Contains(p.BaseInfo().Name, c) {
					countryOk = true
					break
				}
			}
			if !countryOk {
				goto exclude
			}
		}

		if needFilterSpeed && healthcheck.SpeedResults != nil {
			if speed, ok := healthcheck.SpeedResults[p.Identifier()]; ok {
				// clear history result on name
				names := strings.Split(p.BaseInfo().Name, " |")
				if len(names) > 1 {
					p.BaseInfo().Name = names[0]
				}
				// check speed
				if speed > b.Speed {
					p.AddToName(fmt.Sprintf(" |%5.2fMb", speed))
				} else {
					goto exclude
				}
			} else {
				if b.Speed != 0 {
					goto exclude
				}
			}
		} else { // clear speed tag. But I don't know why speed is stored in name while provider get proxies from cache everytime. It's name should be refreshed without speed tag. Because of gin-cache?
			names := strings.Split(p.BaseInfo().Name, " |")
			if len(names) > 1 {
				p.BaseInfo().Name = names[0]
			}
		}

		proxies = append(proxies, p)
	exclude:
	}

	b.Proxies = &proxies
}
