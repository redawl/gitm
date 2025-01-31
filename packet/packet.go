package packet

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"com.github.redawl.mitmproxy/tls"
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
    Data  []byte    `json:"data"`
}

type TlsPacket struct {
    packet
    Records []tls.TLSRecord `json:"records"`
}


func CreatePacket (src string, dst string, data []byte) (Packet) {
    packet := packet{
        SrcIp: src,
        DstIp: dst,
    }
    if data[0] >= 0x14 && data[0] <= 0x18 {
        return TlsPacket{
            packet: packet,
            Records: tls.ParseTLSRecords(data),
        }
    }
    return HttpPacket{
        packet: packet,
        Data: data,
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

func (packet TlsPacket) WritePacket(fd os.File) (error) {
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

func (packet TlsPacket) GetLogAttrs () []slog.Attr {
    attrs := packet.packet.GetLogAttrs()

    for _, record := range(packet.Records) {
        attrs = append(attrs, record.LogAttrs()...)
    }

    return attrs
}

func (packet HttpPacket) GetLogAttrs () []slog.Attr {
    attrs := packet.packet.GetLogAttrs()

    attrs = append(attrs, slog.Any("data", packet.Data))

    return attrs
}
