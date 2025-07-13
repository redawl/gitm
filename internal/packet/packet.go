package packet

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/redawl/gitm/internal"
)

type Packet interface {
	TimeStamp() time.Time
	Type() string
	FindPacket([]Packet) Packet
	UpdatePacket(Packet)
	FormatHostname() string
	FormatRequestLine() string
	FormatResponseLine() string
	FormatRequestContent() string
	FormatResponseContent() string
	MatchesFilter([]internal.FilterToken) bool
}

func MarshalPackets(p []Packet) ([]byte, error) {
	return json.Marshal(p)
}

func UnmarshalPackets(data []byte, p *[]Packet) error {
	var rawPackets []*json.RawMessage
	if err := json.Unmarshal(data, &rawPackets); err != nil {
		return err
	}

	for _, pac := range rawPackets {
		var pacMap map[string]any

		if err := json.Unmarshal(*pac, &pacMap); err != nil {
			return err
		}

		if pacMap["Type"] == nil || pacMap["Type"].(string) == "http" {
			var httpPacket HttpPacket
			if err := json.Unmarshal(*pac, &httpPacket); err != nil {
				return err
			}
			*p = append(*p, &httpPacket)
		} else if pacMap["Type"] == "websocket" {
			var websocketPacket WebsocketPacket
			if err := json.Unmarshal(*pac, &websocketPacket); err != nil {
				return err
			}
			*p = append(*p, &websocketPacket)
		} else {
			slog.Error("Unknown packet type encountered!", "type", pacMap["Type"])
		}
	}
	return nil
}
