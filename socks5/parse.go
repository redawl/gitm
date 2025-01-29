package socks5

import (
	"fmt"
	"net"

	"com.github.redawl.mitmproxy/util"
)

func ParseClientGreeting (conn net.Conn) (*ClientGreeting, error) {
    buff, err := util.Read(conn, 2)

    if err != nil {
        return nil, err
    }
    
    ver := buff[0]
    nauth := buff[1]
    auth, err := util.Read(conn, int(nauth))

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
    buff, err := util.Read(conn, 4)

    if err != nil {
        return nil, err
    }

    ver := buff[0]
    cmd := buff[1]
    rsv := buff[2]
    dstIpType := buff[3]
    dstIp := ""

    if dstIpType == ADDRESS_TYPE_IPV4 {
        buff, err = util.Read(conn, 4)
        if err != nil {
            return nil, err
        }
        dstIp = fmt.Sprintf("%d.%d.%d.%d", buff[0], buff[1], buff[2], buff[3])
    } else if dstIpType == ADDRESS_TYPE_DOMAINNAME {
        domainLength, err := util.Read(conn, 1)
        if err != nil {
            return nil, err
        }

        domain, err := util.Read(conn, int(domainLength[0]))
        if err != nil {
            return nil, err
        }
        lookups, err := net.LookupIP(string(domain))
        if err != nil {
            return nil, err
        }
        dstIp = lookups[0].String()
    } else {
        return nil, fmt.Errorf("Ip type %d is unsupported", dstIpType)
    }

    buff, err = util.Read(conn, 2)
    
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

