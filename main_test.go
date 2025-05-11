package main

import (
	"net/http"
	"testing"

	"github.com/redawl/gitm/internal/config"
	"github.com/redawl/gitm/internal/packet"
)

func setup () (*config.Config, []*packet.HttpPacket) {
	conf := config.Config{
		HttpListenUri: "127.0.0.1:8081",
		TlsListenUri: "127.0.0.1:8444",
		SocksListenUri: "127.0.0.1:1081",
	}

	packets := make([]*packet.HttpPacket, 0)

	setupbackend(conf, func(hp packet.HttpPacket){
		packets = append(packets, &hp)
	})

	return &conf, packets
}

func TestProxyPacIsAccessible (t *testing.T) {
	conf, _ := setup()

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
