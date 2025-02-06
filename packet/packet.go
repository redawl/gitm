package packet

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
)

type Packet interface {
    WritePacket(fd os.File) error
    GetLogAttrs() []slog.Attr
}

type packet struct {
    SrcIp string   `json:"src_ip"`
    DstIp string   `json:"dst_ip"`
}

type HttpPacket struct {
    packet
    Status int 
    Headers map[string][]string
    Content []byte
}

func CreatePacket (src string, dst string, status int, headers map[string][]string, content []byte) (Packet) {
    packet := packet{
        SrcIp: src,
        DstIp: dst,
    }

    return HttpPacket{
        packet: packet,
        Status: status,
        Headers: headers,
        Content: content,
    }
}

func (packet packet) WritePacket(fd os.File) (error) {

    data, err := json.Marshal(packet)

    if err != nil {
        return err
    }

    count, err := fd.Write(data)

    if err != nil {
        return err
    } 

    if count != len(data) {
        return fmt.Errorf("Expected write of size %d, wrote %d", len(data), count)
    }

    return nil
}

func (packet HttpPacket) WritePacket(fd os.File) (error) {
    data, err := json.Marshal(packet)

    if err != nil {
        return err
    }

    count, err := fd.Write(data)

    if err != nil {
        return err
    } 

    if count != len(data) {
        return fmt.Errorf("Expected write of size %d, wrote %d", len(data), count)
    }

    return nil
}

func (packet packet) GetLogAttrs () []slog.Attr {
    return []slog.Attr{
        slog.String("src", packet.SrcIp),
        slog.String("dst", packet.DstIp),
    }
}

func (packet HttpPacket) GetLogAttrs () []slog.Attr {
    attrs := packet.packet.GetLogAttrs()

    return attrs
}
