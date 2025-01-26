package socks5

import (
	"errors"
	"fmt"
	"net"
)

type ClientGreeting struct {
    Ver   byte
    Nauth uint8
    Auth  []byte
}

type ClientConnRequest struct {
    Ver       byte
    Cmd       byte
    Rsv       byte
    DstIpType byte
    DstIp     string
    DstPort   uint16
}

func ParseClientGreeting (conn net.Conn) (*ClientGreeting, error) {
    buff, err := read(conn, 2)

    if err != nil {
        return nil, err
    }
    
    ver := buff[0]
    nauth := buff[1]
    auth, err := read(conn, int(nauth))

    if err != nil {
        return nil, err
    }

    return &ClientGreeting{
        Ver: ver,
        Nauth: nauth,
        Auth: auth,
    }, nil
}

func ParseClientConnRequest (conn net.Conn) (*ClientConnRequest, error) {
    buff, err := read(conn, 4)

    if err != nil {
        return nil, err
    }

    ver := buff[0]
    cmd := buff[1]
    rsv := buff[2]
    dstIpType := buff[3]
    dstIp := ""

    if dstIpType == 0x01 {
        buff, err = read(conn, 4)
        if err != nil {
            return nil, err
        }
        dstIp = fmt.Sprintf("%d.%d.%d.%d", buff[0], buff[1], buff[2], buff[3])
    } else if dstIpType == 0x03 {
        domainLength, err := read(conn, 1)
        if err != nil {
            return nil, err
        }

        domain, err := read(conn, int(domainLength[0]))
        if err != nil {
            return nil, err
        }
        lookups, err := net.LookupIP(string(domain))
        if err != nil {
            return nil, err
        }
        dstIp = lookups[0].String()
    } else {
        return nil, errors.New(fmt.Sprintf("Ip type %d is unsupported", dstIpType))
    }

    buff, err = read(conn, 2)
    
    if err != nil {
        return nil, err
    }

    dstPort := uint16(buff[0]) << 8 + uint16(buff[1])

    return &ClientConnRequest{
        Ver: ver,
        Cmd: cmd,
        Rsv: rsv,
        DstIpType: dstIpType,
        DstIp: dstIp,
        DstPort: dstPort,
    }, nil
}

