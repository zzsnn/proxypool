package app

import (
	"fmt"
	"github.com/Sansui233/proxypool/pkg/proxy"
	"testing"
)

func TestGetters(t *testing.T) {
	proxy.InitGeoIpDB()
	InitConfigAndGetters("config/config.yaml")
	g := Getters[5]
	results := g.Get()
	fmt.Println(len(results))
}