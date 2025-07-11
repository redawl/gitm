package main

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/redawl/gitm/internal/config"
	"github.com/redawl/gitm/internal/packet"
)

func setup(portOffset int, t *testing.T) (*config.Config, []packet.Packet, func()) {
	conf := config.Config{
		SocksListenUri:  fmt.Sprintf("127.0.0.1:%d", 1080+portOffset),
		PacListenUri:    fmt.Sprintf("127.0.0.1:%d", 8080+portOffset),
		EnablePacServer: true,
	}

	packets := make([]packet.Packet, 0)

	cleanup, err := setupBackend(conf, func(hp packet.Packet) {
		packets = append(packets, hp)
	})
	if err != nil {
		t.Errorf("Expected err = nil, got err = %v", err)
	}

	return &conf, packets, cleanup
}

func TestProxyPacIsAccessible(t *testing.T) {
	conf, _, cleanup := setup(1, t)

	defer cleanup()

	resp, err := http.DefaultClient.Get("http://" + conf.PacListenUri + "/proxy.pac")
	if err != nil {
		t.Errorf("Expected err = nil, got err = %v", err)
		return
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status = 200, got status = %d", resp.StatusCode)
		return
	}
}

func TestCaCertIsAccessble(t *testing.T) {
	conf, _, cleanup := setup(2, t)
	defer cleanup()

	proxyUrl, _ := url.Parse("socks5://" + conf.SocksListenUri)
	client := http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		},
	}

	resp, err := client.Get("http://gitm/ca.crt")
	if err != nil {
		t.Errorf("Expected err = nil, got err = %v", err)
		return
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status = 200, got status = %d", resp.StatusCode)
		return
	}
}

func TestConnectivityThroughProxy(t *testing.T) {
	conf, _, cleanup := setup(3, t)
	defer cleanup()

	proxyUrl, _ := url.Parse("socks5://" + conf.SocksListenUri)
	client := http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		},
	}

	resp, err := client.Get("http://example.com")
	if err != nil {
		t.Errorf("Expected err = nil, got err = %v", err)
		return
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status = 200, got status = %d", resp.StatusCode)
		return
	}
}
