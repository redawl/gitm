package socks5

import (
	"net"
    "errors"
    "fmt"
)

func read(conn net.Conn, length int) ([]byte, error) {
    buff := make([]byte, length)

    count, err := conn.Read(buff)
    if err != nil {
        return nil, err
    } else if count != length {
        return nil, errors.New(fmt.Sprintf("Expected length %d, go %d", length, count))
    }

    return buff, nil
}

