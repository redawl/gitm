package socks5

import (
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"strings"
)

func FormatServerChoice(version byte, auth byte) []byte {
    return []byte{version, auth}
}

func FormatConnResponse(
    version byte,
    status  byte,
    bndAddr net.Addr,
) []byte {
    parts := strings.Split(bndAddr.String(), ":")
    ip := parts[0]
    port, err := strconv.Atoi(parts[1])

    if err != nil {
        slog.Error("Cannot parse port %d", parts[1])
    }

    var i1, i2, i3, i4 byte
    fmt.Sscanf(ip, "%d.%d.%d.%d", &i1, &i2, &i3, &i4)
    return []byte{
        version,
        status,
        0x00, // Rsv is always 0x00
        // No IPV6 support for now
        ADDRESS_TYPE_IPV4, i1, i2, i3, i4,
        byte(port >> 8),
        byte(port & 0xFF),
    }
}
