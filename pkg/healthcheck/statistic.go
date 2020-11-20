package healthcheck

import "github.com/Sansui233/proxypool/pkg/proxy"

// Statistic for a proxy
type Stat struct {
	Speed    float64
	Delay    uint16
	ReqCount uint16
	Id       string
}

// Statistic array for proxies
type StatList []Stat

// Global var
var PStats StatList

func init() {
	PStats = make(StatList, 0)
}

// Update speed for a Stat
func (ps *Stat) UpdatePSSpeed(speed float64) {
	if ps.Speed < 60 && ps.Speed != 0 {
		ps.Speed = 0.3*ps.Speed + 0.7*speed
	} else {
		ps.Speed = speed
	}
}

// Update delay for a Stat
func (ps *Stat) UpdatePSDelay(delay uint16) {
	ps.Delay = delay
}

// Count + 1 for a Stat
func (ps *Stat) UpdatePSCount() {
	ps.ReqCount++
}

// Find a proxy's Stat in StatList
func (pss StatList) Find(p proxy.Proxy) (*Stat, bool) {
	s := p.Identifier()
	for i, _ := range pss {
		if pss[i].Id == s {
			return &pss[i], true
		}
	}
	return nil, false
}

// Return proxies that request count more than a given nubmer
// todo 不该人为指定n，不同时段不同服务器都有差别，计算比例有可能更好，流量也
func (pss StatList) ReqCountThan(n uint16, pl []proxy.Proxy, reset bool) []proxy.Proxy {
	proxies := make([]proxy.Proxy, 0)
	for _, p := range pl {
		for j, _ := range pss {
			if pss[j].ReqCount > n && p.Identifier() == pss[j].Id {
				proxies = append(proxies, p)
			}
		}
	}
	if reset {
		for i, _ := range pss {
			pss[i].ReqCount = 0
		}
	}
	return proxies
}
