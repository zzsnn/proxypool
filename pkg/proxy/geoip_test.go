package proxy

import (
	"fmt"
	"testing"
)

func TestGeoIP_Find(t *testing.T) {
	InitGeoIpDB()
	_,country,_ := geoIp.Find("220.181.38.148")
	fmt.Println(country)
}
