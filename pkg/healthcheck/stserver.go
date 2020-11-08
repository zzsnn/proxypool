package healthcheck

import (
	"bytes"
	"encoding/xml"
	C "github.com/Dreamacro/clash/constant"
	"math"
	"sort"
	"strconv"
	"sync"
	"time"
)

// Server information
type Server struct {
	URL      string `xml:"url,attr"`
	Lat      string `xml:"lat,attr"`
	Lon      string `xml:"lon,attr"`
	Name     string `xml:"name,attr"`
	Country  string `xml:"country,attr"`
	Sponsor  string `xml:"sponsor,attr"`
	ID       string `xml:"id,attr"`
	URL2     string `xml:"url2,attr"`
	Host     string `xml:"host,attr"`
	Distance float64
	DLSpeed  float64
}

// ServerList : List of Server. for xml decoding
type ServerList struct {
	Servers []Server `xml:"servers>server"`
}

// Servers : For sorting servers.
type Servers []Server

// ByDistance : For sorting servers.
type ByDistance struct {
	Servers
}

// Len : length of servers. For sorting servers.
func (s Servers) Len() int {
	return len(s)
}

// Swap : swap i-th and j-th. For sorting servers.
func (s Servers) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less : compare the distance. For sorting servers.
func (b ByDistance) Less(i, j int) bool {
	return b.Servers[i].Distance < b.Servers[j].Distance
}

func fetchServerList(user User, clashProxy C.Proxy) (*ServerList, error) {
	url := "http://www.speedtest.net/speedtest-servers-static.php"
	body, err := HTTPGetBodyViaProxy(clashProxy, url)
	if err != nil {
		return nil, err
	}

	if len(body) == 0 {
		url = "http://c.speedtest.net/speedtest-servers-static.php"
		body, err = HTTPGetBodyViaProxy(clashProxy, url)
		if err != nil {
			return nil, err
		}
	}

	// Decode xml
	decoder := xml.NewDecoder(bytes.NewReader(body))
	list := ServerList{}
	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			decoder.DecodeElement(&list, &se)
		}
	}

	// Calculate distance
	for i := range list.Servers {
		server := &list.Servers[i]
		sLat, _ := strconv.ParseFloat(server.Lat, 64)
		sLon, _ := strconv.ParseFloat(server.Lon, 64)
		uLat, _ := strconv.ParseFloat(user.Lat, 64)
		uLon, _ := strconv.ParseFloat(user.Lon, 64)
		server.Distance = distance(sLat, sLon, uLat, uLon)
	}

	// Sort by distance
	sort.Sort(ByDistance{list.Servers})

	return &list, nil
}

func distance(lat1 float64, lon1 float64, lat2 float64, lon2 float64) float64 {
	radius := 6378.137

	a1 := lat1 * math.Pi / 180.0
	b1 := lon1 * math.Pi / 180.0
	a2 := lat2 * math.Pi / 180.0
	b2 := lon2 * math.Pi / 180.0

	x := math.Sin(a1)*math.Sin(a2) + math.Cos(a1)*math.Cos(a2)*math.Cos(b2-b1)
	return radius * math.Acos(x)
}

// StartTest : start testing to the servers.
func (svrs Servers) StartTest(clashProxy C.Proxy) {
	var wg sync.WaitGroup
	for i, s := range svrs {
		wg.Add(1)
		latency := pingTest(clashProxy, s.URL)
		if latency == time.Second*5 { // fail to get latency, skip
			continue
		}
		dlSpeed := downloadTest(clashProxy, s.URL, latency)
		svrs[i].DLSpeed = dlSpeed
	}
}

// GetResult : return testing result. -1 for no effective result
func (svrs Servers) GetResult() float64 {
	if len(svrs) == 1 {
		return svrs[0].DLSpeed
	} else {
		avgDL := 0.0
		count := 0
		for _, s := range svrs {
			if s.DLSpeed != 0 {
				avgDL = avgDL + s.DLSpeed
				count++
			}
		}
		if count == 0 {
			return -1
		}
		//fmt.Printf("Download Avg: %5.2f Mbit/s\n", avgDL/float64(len(svrs)))
		return avgDL / float64(count)
	}

}
