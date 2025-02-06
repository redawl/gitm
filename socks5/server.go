package socks5

import (
	"fmt"
	"io"
	"log/slog"
	"net"

	"com.github.redawl.mitmproxy/config"
)

func StartTransparentSocksProxy(ListenUri string, conf config.Config) (error) {
    ln, err := net.Listen("tcp", ListenUri)

    if err != nil {
        return err
    }

    for {
        client, err := ln.Accept()
        if err != nil {
            slog.Error("Error accepting connection", "error", err)
            continue
        }
        
        if err := handleConnection(client, conf); err != nil {
            slog.Error("Error handling connection", "error", err)
            continue
        }
    }
}

func handleConnection(client net.Conn, conf config.Config) (error) {
    slog.Debug("Received connection", "Address", client.RemoteAddr())

    greeting, err := ParseClientGreeting(client)

    if err != nil {
        return err
    }

    slog.Debug("Parsed client greeting", "greeting", greeting)
    
    if greeting.CanHandle() {
        slog.Debug("Handling Request")
        client.Write(
            FormatServerChoice(SOCKS_VER_5, METHOD_NO_AUTH_REQUIRED),
        )

        request, status := ParseClientConnRequest(client)

        if status != STATUS_SUCCEEDED {
            client.Write(FormatConnResponse(
                SOCKS_VER_5,
                status,
                client.LocalAddr(),
            ))
            return fmt.Errorf("Error parsing client connection request: %d", status)
        }
        
        slog.Debug("Parsed conn request", "request", request)
        if request.DstPort == 80 {
            server, err := net.Dial("tcp", conf.HttpListenUri)
            if err != nil {
                slog.Error("Error contacting http proxy server", "error", err)
                client.Write(FormatConnResponse(
                    SOCKS_VER_5,
                    STATUS_HOST_UNREACHABLE,
                    server.LocalAddr(),
                ))
                return err
            }

            slog.Debug("Proxy success")
            client.Write(FormatConnResponse(
                SOCKS_VER_5,
                STATUS_SUCCEEDED,
                server.LocalAddr(),
            ))

            transparentProxy(client, server)

        } else if request.DstPort == 443 {
            server, err := net.Dial("tcp", conf.TlsListenUri)
            if err != nil {
                slog.Error("Error contacting https proxy server", "error", err)
                client.Write(FormatConnResponse(
                    SOCKS_VER_5,
                    STATUS_HOST_UNREACHABLE,
                    server.LocalAddr(),
                ))
                return err
            }

            slog.Debug("Proxy success")
            client.Write(FormatConnResponse(
                SOCKS_VER_5,
                STATUS_SUCCEEDED,
                server.LocalAddr(),
            ))

            transparentProxy(client, server)
        } else {
            slog.Error("Unrecognized port, forwarding without logging", "request", request)
            server, err := net.Dial("tcp", fmt.Sprintf("%s:%d", request.DstIp, request.DstPort))
            if err != nil {
                client.Write(FormatConnResponse(
                    SOCKS_VER_5,
                    STATUS_HOST_UNREACHABLE,
                    server.LocalAddr(),
                ))
                return err
            }

            slog.Debug("Proxy success")
            client.Write(FormatConnResponse(
                SOCKS_VER_5,
                STATUS_SUCCEEDED,
                server.LocalAddr(),
            ))

            transparentProxy(client, server)
        }

    } else {
        slog.Debug("Cannot handle request")
        client.Write(FormatServerChoice(SOCKS_VER_5, METHOD_NO_ACCEPTABLE_METHODS))
    }

    return nil
}

func transparentProxy (client net.Conn, server net.Conn) {
    go connToConn(client, server)
    go connToConn(server, client)
}

func connToConn(conn1 net.Conn, conn2 net.Conn) {
    buff := make([]byte, 8192)
    for {
        count, err := conn1.Read(buff)
        if err != nil && count == 0 {
            if err == io.EOF {
                slog.Debug("Connection terminated", "error", err, "count", count)
            } else {
                slog.Error("Connection closed unexpectedly", "error", err, "count", count)
            }
            return
        } else if err != nil {
            slog.Error("Connection terminated but contained data", "error", err, "count", count)
        }
        count, err = conn2.Write(buff[:count])

        if err != nil && count != 0 {
            if err == io.EOF {
                slog.Debug("Connection terminated", "error", err, "count", count)
            } else {
                slog.Error("Connection closed unexpectedly", "error", err, "count", count)
            }
        } else if err != nil {
            slog.Error("Connection terminated but didn't write all data", "error", err, "count", count)
            return
        }
    }
}

