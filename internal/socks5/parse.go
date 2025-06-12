package socks5

import (
	"fmt"
	"net"
)

func ParseClientGreeting(conn net.Conn) (*ClientGreeting, error) {
	buff, err := readCount(conn, 2)

	if err != nil {
		return nil, err
	}

	ver := buff[0]
	nauth := buff[1]
	auth, err := readCount(conn, int(nauth))

	if err != nil {
		return nil, err
	}

	return &ClientGreeting{
		Ver:   ver,
		Nauth: nauth,
		Auth:  auth,
	}, nil
}

func ParseClientConnRequest(conn net.Conn) (*ClientConnRequest, byte) {
	buff, err := readCount(conn, 4)

	if err != nil {
		return nil, STATUS_GENERAL_FAILURE
	}

	ver := buff[0]
	cmd := buff[1]
	rsv := buff[2]
	dstIpType := buff[3]
	dstIp := ""

	if cmd != CMD_CONNECT {
		return nil, STATUS_COMMAND_NOT_SUPPORTED
	}
	switch dstIpType {
	case ADDRESS_TYPE_IPV4:
		buff, err = readCount(conn, 4)
		if err != nil {
			return nil, STATUS_GENERAL_FAILURE
		}
		dstIp = fmt.Sprintf("%d.%d.%d.%d", buff[0], buff[1], buff[2], buff[3])
	case ADDRESS_TYPE_DOMAINNAME:
		domainLength, err := readCount(conn, 1)
		if err != nil {
			return nil, STATUS_GENERAL_FAILURE
		}

		domain, err := readCount(conn, int(domainLength[0]))

		if err != nil {
			return nil, STATUS_GENERAL_FAILURE
		}

		// Special handling here for our internal hostname, for /proxy.pac and /ca.crt
		if string(domain) == "gitm" {
			dstIp = conn.LocalAddr().String()
		} else {
			lookups, err := net.LookupIP(string(domain))

			if err != nil {
				return nil, STATUS_HOST_UNREACHABLE
			}
			dstIp = lookups[0].String()
		}
	default:
		return nil, STATUS_ADDRESS_TYPE_NOT_SUPPORTED
	}

	buff, err = readCount(conn, 2)

	if err != nil {
		return nil, STATUS_GENERAL_FAILURE
	}

	dstPort := uint16(buff[0])<<8 + uint16(buff[1])

	return &ClientConnRequest{
		Ver:       ver,
		Cmd:       cmd,
		Rsv:       rsv,
		DstIpType: dstIpType,
		DstIp:     dstIp,
		DstPort:   dstPort,
	}, STATUS_SUCCEEDED
}

// readCount reads at most length bytes from conn.
// If less than length bytes are read from conn, the bytes are returned along with an err
func readCount(conn net.Conn, length int) ([]byte, error) {
	buff := make([]byte, length)

	count, err := conn.Read(buff)
	if err != nil {
		return nil, err
	} else if count != length {
		return buff[:count], fmt.Errorf("expected length %d, go %d", length, count)
	}

	return buff, nil
}
