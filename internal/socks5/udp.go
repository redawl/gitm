package socks5

import (
	"fmt"
	"log/slog"
	"net"
)

func handleUDP(client net.Conn, request *ClientConnRequest) error {
	slog.Info("Handling udp!", "Addr", client.RemoteAddr())
	server, err := net.Dial("udp", net.JoinHostPort(request.DstIP, fmt.Sprintf("%d", request.DstPort)))
	if err != nil {
		if _, err := client.Write(FormatConnResponse(
			SocksVer5,
			StatusHostUnreachable,
			client.RemoteAddr(),
		)); err != nil {
			return fmt.Errorf("formatting conn response: %w", err)
		}
		return fmt.Errorf("handling udp: %w", err)
	}

	if l, err := net.Listen("udp", "0.0.0.0:42069"); err != nil {
		slog.Error("Error opening udp conn for client", "error", err)
	} else {
		go func() {
			for {
				if conn, err := l.Accept(); err != nil {
					slog.Error("Error accepting udp connection", "error", err)
				} else {
					go transparentProxy(server, conn)
					go transparentProxy(conn, server)
				}
			}
		}()

		if _, err := client.Write(FormatConnResponse(
			SocksVer5,
			StatusSucceeded,
			l.Addr(),
		)); err != nil {
			return fmt.Errorf("formatting conn response: %w", err)
		}
	}

	return nil
}
