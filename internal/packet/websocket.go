package packet

import (
	"bytes"
	"compress/flate"
	"io"
	"log/slog"
	"slices"
	"time"
)

var _ Packet = (*WebsocketPacket)(nil)

type WebsocketPacket struct {
	HTTPPacket
	ServerFrames []*WebsocketFrame
	ClientFrames []*WebsocketFrame
}

func CreateWebsocketPacket(httpPacket HTTPPacket) *WebsocketPacket {
	packet := &WebsocketPacket{
		HTTPPacket:   httpPacket,
		ServerFrames: make([]*WebsocketFrame, 0),
		ClientFrames: make([]*WebsocketFrame, 0),
	}
	packet.Type_ = "websocket"

	return packet
}

func (w *WebsocketPacket) FormatRequestContent() string {
	builder := bytes.Buffer{}
	builder.WriteString(w.HTTPPacket.FormatRequestContent())
	builder.WriteString("\r\n")
	builder.WriteString(w.HTTPPacket.FormatResponseContent())

	return builder.String()
}

func (w *WebsocketPacket) FormatResponseContent() string {
	packets := append(w.createPacketsFromFrames(w.ClientFrames), w.createPacketsFromFrames(w.ServerFrames)...)

	slices.SortFunc(packets, func(a, b *websocketPacket) int {
		return a.TimeStamp.Compare(b.TimeStamp)
	})

	buff := bytes.Buffer{}

	for _, p := range packets {
		buff.WriteString(p.TimeStamp.String())
		buff.WriteByte('\n')
		if p.Type == ClientFrame {
			buff.WriteString("--> ")
		} else {
			buff.WriteString("<-- ")
		}
		for _, c := range p.Payload {
			buff.WriteByte(c)
			if c == '\n' {
				buff.WriteString("    ")
			}
		}
		buff.WriteString("\n----------------------\n")
	}

	return buff.String()
}

func (w *WebsocketPacket) FindPacket(packets []Packet) Packet {
	for _, pac := range packets {
		if websocketPacket, ok := pac.(*WebsocketPacket); ok && w.ID == websocketPacket.ID {
			return websocketPacket
		}
	}

	return nil
}

func (w *WebsocketPacket) UpdatePacket(p Packet) {
	w.HTTPPacket.UpdatePacket(p)
	if wp, ok := p.(*WebsocketPacket); ok {
		w.ServerFrames = wp.ServerFrames
		w.ClientFrames = wp.ClientFrames
	}
}

func (w *WebsocketPacket) AddServerFrame(frame *WebsocketFrame) {
	frame.Type = ServerFrame
	w.ServerFrames = append(w.ServerFrames, frame)
}

func (w *WebsocketPacket) AddClientFrame(frame *WebsocketFrame) {
	frame.Type = ClientFrame
	w.ClientFrames = append(w.ClientFrames, frame)
}

func (w *WebsocketPacket) createPacketsFromFrames(frames []*WebsocketFrame) []*websocketPacket {
	index := 0
	packets := make([]*websocketPacket, 0)
	for index < len(frames) {
		buff := bytes.Buffer{}
		frame := frames[index]
		if !frame.Masked {
			buff.Write(frame.Payload)
		} else {
			for i := range len(frame.Payload) {
				buff.WriteByte(frame.Payload[i] ^ frame.MaskingKey[i%4])
			}
		}
		for !frames[index].Fin {
			index++
			if index >= len(frames) {
				slog.Error("Never found ending frame")
				break
			}
			frame = frames[index]
			if !frame.Masked {
				buff.Write(frame.Payload)
			} else {
				for i := range len(frame.Payload) {
					buff.WriteByte(frame.Payload[i] ^ frame.MaskingKey[i%4])
				}
			}
		}
		index++
		if w.ReqHeaders["Sec-Websocket-Extensions"][0] == "permessage-deflate" && frame.RSV1 {
			compressed := append(buff.Bytes(), 0x00, 0x00, 0xff, 0xff)

			if uncompressed, err := io.ReadAll(flate.NewReader(bytes.NewBuffer(compressed))); err != nil {
				slog.Error("Decompress flate", "error", err)
			} else {
				buff.Reset()
				buff.Write(uncompressed)
			}
		}
		packets = append(packets, &websocketPacket{
			TimeStamp: frames[index-1].TimeStamp,
			Type:      frames[index-1].Type,
			Payload:   buff.Bytes(),
		})
	}

	return packets
}

type frameType bool

const (
	ServerFrame frameType = frameType(false)
	ClientFrame frameType = frameType(true)
)

type WebsocketFrame struct {
	TimeStamp     time.Time
	Type          frameType
	Fin           bool
	RSV1          bool
	RSV2          bool
	RSV3          bool
	Opcode        byte
	Masked        bool
	PayloadLength uint64
	MaskingKey    [4]byte
	// TODO: Split this up into extension data and application data
	Payload []byte
}

// TODO: Better name?
type websocketPacket struct {
	TimeStamp time.Time
	Type      frameType
	Payload   []byte
}
