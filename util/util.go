package util

import (
	"net"
    "fmt"
)
// Read reads at most length bytes from conn.
// If less than length bytes are read from conn, the bytes are returned along with an err
func Read(conn net.Conn, length int) ([]byte, error) {
    buff := make([]byte, length)

    count, err := conn.Read(buff)
    if err != nil {
        return nil, err
    } else if count != length {
        return buff[:count], fmt.Errorf("Expected length %d, go %d", length, count)
    }

    return buff, nil
}

