package util

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
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

func GetConfigDir () (string, error) {
    userCfgDir, err := os.UserConfigDir()

    if err != nil {
        return "", err
    }

    cfgDir := userCfgDir + "/mitmproxy"

    if _, err := os.Stat(cfgDir); errors.Is(err, os.ErrNotExist) {
        slog.Debug("Config dir doesn't exist, creating")
        err := os.Mkdir(cfgDir, 0700)

        if err != nil {
            return "", err
        }
    }

    return cfgDir, nil
}
