package main

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/redawl/gitm/internal/config"
	"github.com/redawl/gitm/internal/packet"
)

func setup(t *testing.T) (*config.Config, []*packet.HttpPacket, func()) {
	conf := config.Config{
		HttpListenUri:  "127.0.0.1:8081",
		TlsListenUri:   "127.0.0.1:8444",
		SocksListenUri: "127.0.0.1:1081",
	}

	packets := make([]*packet.HttpPacket, 0)

	cleanup, err := setupBackend(conf, func(hp packet.HttpPacket) {
		packets = append(packets, &hp)
	})
	if err != nil {
		t.Errorf("Expected err = nil, got err = %v", err)
	}

	return &conf, packets, cleanup
}

func TestProxyPacIsAccessible(t *testing.T) {
	conf, _, cleanup := setup(t)

	defer cleanup()

	resp, err := http.DefaultClient.Get("http://" + conf.HttpListenUri + "/proxy.pac")
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
	conf, _, cleanup := setup(t)
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
