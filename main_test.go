package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/redawl/gitm/internal"
	"github.com/redawl/gitm/internal/packet"
)

func setup(portOffset int, t *testing.T) (*internal.Config, []packet.Packet, func()) {
	conf := internal.Config{
		SocksListenURI:  fmt.Sprintf("127.0.0.1:%d", 1080+portOffset),
		PACListenURI:    fmt.Sprintf("127.0.0.1:%d", 8080+portOffset),
		EnablePACServer: true,
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

	resp, err := http.DefaultClient.Get("http://" + conf.PACListenURI + "/proxy.pac")
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

	proxyURL, _ := url.Parse("socks5://" + conf.SocksListenURI)
	client := http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("<h1>Hello!</h1>"))
	}))

	proxyURL, _ := url.Parse("socks5://" + conf.SocksListenURI)
	client := http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}

	resp, err := client.Get(server.URL)
	if err != nil {
		t.Errorf("Expected err = nil, got err = %v", err)
		return
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status = 200, got status = %d", resp.StatusCode)
		return
	}
}
