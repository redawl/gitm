package socks5

import (
	"encoding/binary"
	"net"
)

func uint32ToIpStr(ipIn uint32) (string) {
    ipOut := make(net.IP, 4)
    binary.BigEndian.PutUint32(ipOut, ipIn)

    return ipOut.String()
}
