package socks5

import (
	"fmt"
	"log/slog"
)

type ServerChoice struct {
    Ver   byte
    Cauth byte
}

type ServerConnResponse struct {
    Ver      byte
    Status   byte
    Rsv     byte
    BndAddr string
    BndPort uint16
}

func FormatServerChoice(choice *ServerChoice) ([]byte) {
    return []byte{choice.Ver, choice.Cauth}
}

func FormatConnResponse(response *ServerConnResponse) ([]byte) {
    var i1, i2, i3, i4 byte
    fmt.Sscanf(response.BndAddr, "%d.%d.%d.%d", &i1, &i2, &i3, &i4)
    return []byte{
        response.Ver,
        response.Status,
        response.Rsv,
        i1, i2, i3, i4,
        byte(response.BndPort >> 8),
        byte(response.BndPort & 0xFF),
    }
}
