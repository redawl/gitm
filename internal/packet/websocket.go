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
	HttpPacket
	ServerFrames []*WebsocketFrame
	ClientFrames []*WebsocketFrame
}

func CreateWebsocketPacket(httpPacket HttpPacket) *WebsocketPacket {
	packet := &WebsocketPacket{
		HttpPacket:   httpPacket,
		ServerFrames: make([]*WebsocketFrame, 0),
		ClientFrames: make([]*WebsocketFrame, 0),
	}
	packet.Type_ = "websocket"

	return packet
}

func (w *WebsocketPacket) FormatRequestContent() string {
	builder := bytes.Buffer{}
	builder.WriteString(w.HttpPacket.FormatRequestContent())
	index := 0

	slices.SortFunc(w.ServerFrames, func(a *WebsocketFrame, b *WebsocketFrame) int {
		return a.TimeStamp.Compare(b.TimeStamp)
	})
	for index < len(w.ServerFrames) {
		builder.WriteString("\n-------------\n")
		buff := bytes.Buffer{}
		frame := w.ServerFrames[index]
		if !frame.Masked {
			builder.Write(frame.Payload)
		} else {
			for i := range len(frame.Payload) {
				builder.WriteByte(frame.Payload[i] ^ frame.MaskingKey[i%4])
			}
		}
		for !w.ServerFrames[index].Fin {
			index++
			if index >= len(w.ServerFrames) {
				slog.Error("Never found ending frame")
			}
			frame = w.ServerFrames[index]
			buff.Write(frame.Payload)
			if !frame.Masked {
				builder.Write(frame.Payload)
			} else {
				for i := range len(frame.Payload) {
					builder.WriteByte(frame.Payload[i] ^ frame.MaskingKey[i%4])
				}
			}
		}
		index++
		if w.ReqHeaders["Sec-Websocket-Extensions"][0] == "permessage-deflate" && frame.RSV1 {
			compressed := append(buff.Bytes(), 0x00, 0x00, 0xff, 0xff)

			if uncompressed, err := io.ReadAll(flate.NewReader(bytes.NewBuffer(compressed))); err != nil {
				slog.Error("Decompress flate", "error", err)
			} else {
				builder.Write(uncompressed)
			}
		} else {
			builder.Write(buff.Bytes())
		}
		buff.Reset()
	}

	return builder.String()
}

func (w *WebsocketPacket) FormatResponseContent() string {
	builder := bytes.Buffer{}
	builder.WriteString(w.HttpPacket.FormatResponseContent())
	index := 0

	slices.SortFunc(w.ClientFrames, func(a *WebsocketFrame, b *WebsocketFrame) int {
		return a.TimeStamp.Compare(b.TimeStamp)
	})
	for index < len(w.ClientFrames) {
		builder.WriteString("\n-------------\n")
		buff := bytes.Buffer{}
		frame := w.ClientFrames[index]
		if !frame.Masked {
			builder.Write(frame.Payload)
		} else {
			for i := range len(frame.Payload) {
				builder.WriteByte(frame.Payload[i] ^ frame.MaskingKey[i%4])
			}
		}
		for !w.ClientFrames[index].Fin {
			index++
			if index >= len(w.ClientFrames) {
				slog.Error("Never found ending frame")
			}
			frame = w.ClientFrames[index]
			buff.Write(frame.Payload)
			if !frame.Masked {
				builder.Write(frame.Payload)
			} else {
				for i := range len(frame.Payload) {
					builder.WriteByte(frame.Payload[i] ^ frame.MaskingKey[i%4])
				}
			}
		}
		index++
		if w.ReqHeaders["Sec-Websocket-Extensions"][0] == "permessage-deflate" && frame.RSV1 {
			compressed := append(buff.Bytes(), 0x00, 0x00, 0xff, 0xff)

			if uncompressed, err := io.ReadAll(flate.NewReader(bytes.NewBuffer(compressed))); err != nil {
				slog.Error("Decompress flate", "error", err)
				builder.Write(buff.Bytes())
			} else {
				builder.Write(uncompressed)
			}
		} else {
			builder.Write(buff.Bytes())
		}
		buff.Reset()
	}

	return builder.String()
}

func (w *WebsocketPacket) FindPacket(packets []Packet) Packet {
	for _, pac := range packets {
		if websocketPacket, ok := pac.(*WebsocketPacket); ok && w.Id == websocketPacket.Id {
			return websocketPacket
		}
	}

	return nil
}

func (w *WebsocketPacket) UpdatePacket(p Packet) {
	w.HttpPacket.UpdatePacket(p)
	if wp, ok := p.(*WebsocketPacket); ok {
		w.ServerFrames = wp.ServerFrames
		w.ClientFrames = wp.ClientFrames
	}
}

type WebsocketFrame struct {
	TimeStamp     time.Time
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
