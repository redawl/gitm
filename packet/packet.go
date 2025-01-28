package packet

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
)

type Packet struct {
    SrcIp string   `json:"src_ip"`
    DstIp string   `json:"dst_ip"`
    Data  []byte    `json:"data"`
}

func CreatePacket (data []byte, src net.Conn) (*Packet) {
    return &Packet{
        SrcIp: src.RemoteAddr().String(),
        DstIp: src.LocalAddr().String(),
        Data: data,
    }
}

func (packet *Packet) WritePacket(fd os.File) (error) {
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
