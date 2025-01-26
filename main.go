package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"net"
	"slices"

	"com.github.redawl.mitmproxy/socks5"
)

func ReadLen (connection net.Conn, length int) ([]byte) {
    buff := make([]byte, 100)
    var data []byte
    for _, err := connection.Read(buff); len(data) < length; _, err = connection.Read(buff) {
        if err != nil {
            slog.Error("Error occurred", "error", err)
            return nil
        }

        data = append(data, buff...)
        data := bytes.TrimLeft(data, "\x00")

        if len(data) > length {
            return data[:length]
        }
    }

    return data[:length]
}

func connToChan(channel chan []byte, conn net.Conn) {

    buff := make([]byte, 100)
    for {
        _, err := conn.Read(buff)
        if err != nil {
            slog.Info("Connection terminated", "error", err)
            return
        }
        channel <- buff
        slog.Info("connToChan", "buff", buff, "localip", conn.LocalAddr(), "remoteip", conn.RemoteAddr())
    }
}

func chanToConn(channel chan []byte, conn net.Conn) {
    for {
        buff := <- channel

        conn.Write(buff)
        slog.Info("chanToCon", "buff", buff, "localip", conn.LocalAddr(), "remoteip", conn.RemoteAddr())
    }
}

func main () {
    slog.Info("Hello world")

    ln, err := net.Listen("tcp", ":8080")
    if err != nil {
        // handle error
    }
    for {
        conn, err := ln.Accept()
        if err != nil {
            // handle error
        }
        go func(connection net.Conn){
            slog.Info("Recieved connection", "RemoteAddr", connection.RemoteAddr())

            message := ReadLen(connection, 3)

            clientGreeting := socks5.ParseClientGreeting(message)
            slog.Info("Parsed client greeting", "clientGreeting", clientGreeting)

            if clientGreeting.Ver == 0x05 && slices.Contains(clientGreeting.Auth, 0x00) {
                slog.Info("Can handle request")
                connection.Write(socks5.FormatServerChoice(&socks5.ServerChoice{
                    Ver: 0x05, 
                    Cauth: 0x00,
                }))
    
                message := ReadLen(connection, 10)
                slog.Info("Message", "message", message)
                
                connectRequest := socks5.ParseClientConnRequest(message)

                slog.Info("Client conn request", "request", connectRequest)
                conn, err := net.Dial("tcp", fmt.Sprintf("%s:80", connectRequest.DstIp))

                if err != nil {
                    slog.Info("Couldn't connect to requested location", "error", err)
                    connection.Write(socks5.FormatConnResponse(&socks5.ServerConnResponse{
                        Ver: 0x05,
                        Status: 0x04,
                        Rsv: 0x00,
                        BndAddr: connectRequest.DstIp,
                        BndPort: connectRequest.DstPort,
                    }))
                } else {
                    slog.Info("Connection proxied")
                    connection.Write(socks5.FormatConnResponse(&socks5.ServerConnResponse{
                        Ver: 0x05,
                        Status: 0x00,
                        Rsv: 0x00,
                        BndAddr: connectRequest.DstIp,
                        BndPort: connectRequest.DstPort,
                    }))

                    clientToServer := make(chan []byte)
                    serverToClient := make(chan []byte)
                    
                    // client -> chan
                    go connToChan(clientToServer, connection)

                    // chan -> server
                    go chanToConn(clientToServer, connection)

                    // server -> chan
                    go connToChan(serverToClient, conn)

                    // chan -> client
                    go chanToConn(serverToClient, conn)
                }
            } else {
                slog.Info("Cannot handle request")
                connection.Write(socks5.FormatServerChoice(&socks5.ServerChoice{
                    Ver: 0x05,
                    Cauth: 0xFF,
                }))
            }
        }(conn)
    }
}
