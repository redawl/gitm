package socks5

import (
	"fmt"
)

type ClientGreeting struct {
    Ver   byte
    Nauth uint8
    Auth  []byte
}

type ClientConnRequest struct {
    Ver     byte
    Cmd     byte
    Rsv     byte
    DstIp   string
    DstPort uint16
}

func ParseClientGreeting (message []byte) (*ClientGreeting) {
    ver := message[0]
    nauth := message[1]
    auth := make([]byte, nauth)
    for i := range(nauth) {
        auth[i] = message[2+i]
    }

    return &ClientGreeting{
        Ver: ver,
        Nauth: nauth,
        Auth: auth,
    }
}

func ParseClientConnRequest (message []byte) (*ClientConnRequest) {
    return &ClientConnRequest{
        Ver: message[0],
        Cmd: message[1],
        Rsv: message[2],
        // Ignore Type byte, only supporting ipv4 addresses for now
        DstIp: fmt.Sprintf("%d.%d.%d.%d", message[4], message[5], message[6], message[7]),
        DstPort: uint16(message[8]) + uint16(message[9]),
    }
}

