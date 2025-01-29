package socks5

import (
	"fmt"
	"io"
	"log/slog"
	"net"

	"com.github.redawl.mitmproxy/packet"
)

func StartTransparentSocksProxy(ListenUri string, PacketHandler func(packet.Packet)) (error) {
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
        
        if err := handleConnection(client, PacketHandler); err != nil {
            slog.Error("Error handling connection", "error", err)
            continue
        }
    }
}

func handleConnection(client net.Conn, PacketHandler func(packet.Packet)) (error) {
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

        transparentProxy(client, server, PacketHandler)
    } else {
        slog.Debug("Cannot handle request")
        client.Write(FormatServerChoice(SOCKS_VER_5, METHOD_NO_ACCEPTABLE_METHODS))
    }

    return nil
}

func transparentProxy (client net.Conn, server net.Conn, packetHandler func(packet.Packet)) {
    clientToServer := make(chan []byte)
    serverToClient := make(chan []byte)

    go connToConn(client, server, clientToServer)
    go connToConn(server, client, serverToClient)

    go func() {
        data := <- clientToServer
        packetHandler(packet.Packet{
            SrcIp: client.RemoteAddr().String(),
            DstIp: server.RemoteAddr().String(),
            Data: data,
        })

        data = <- serverToClient
        packetHandler(packet.Packet{
            SrcIp: server.RemoteAddr().String(),
            DstIp: client.RemoteAddr().String(),
            Data: data,
        })
    }()
}

func connToConn(conn1 net.Conn, conn2 net.Conn, outChan chan []byte) {
    buff := make([]byte, 8192)
    out := make([]byte, 0)
    for {
        count, err := conn1.Read(buff)
        if err != nil && count == 0 {
            if err == io.EOF {
                slog.Debug("Connection terminated", "error", err, "count", count)
            } else {
                slog.Error("Connection closed unexpectedly", "error", err, "count", count)
            }

            err = conn2.Close()
            if err != nil {
                slog.Error("Error closing connection2", "error", err)
            }

            outChan <- out
            return
        } else if err != nil {
            slog.Error("Connection terminated but contained data", "error", err, "count", count)
        }
        out = append(out, buff[:count]...)
        count, err = conn2.Write(buff[:count])

        if err != nil && count != 0 {
            if err == io.EOF {
                slog.Debug("Connection terminated", "error", err, "count", count)
            } else {
                slog.Error("Connection closed unexpectedly", "error", err, "count", count)
            }
            conn1.Close()
            outChan <- out
            return
        } else if err != nil {
            slog.Error("Connection terminated but didn't write all data", "error", err, "count", count)
            err = conn1.Close()
            if err != nil {
                slog.Error("Error closing connection1", "error", err)
            }
            outChan <- out
            return
        }
    }
}

