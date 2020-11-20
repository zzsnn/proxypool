package healthcheck

import "github.com/Sansui233/proxypool/pkg/proxy"

type Stat struct {
	Speed    float64
	Delay    uint16
	ReqCount uint16
	Id       string
}

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
