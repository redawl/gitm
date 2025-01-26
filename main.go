package main

import (
	"fmt"
	"log/slog"
	"net"
	"slices"
	"strings"
	"time"

	"com.github.redawl.mitmproxy/socks5"
)

func serverToClient(conn1 net.Conn, conn2 net.Conn, outChan chan string) {
    conn1.SetDeadline(time.Now().Add(time.Second))
    conn2.SetDeadline(time.Now().Add(time.Second))
    buff := make([]byte, 1)
    out := strings.Builder{}
    for {
        count, err := conn1.Read(buff)
        if err != nil {
            slog.Debug("Connection terminated", "error", err)
            outChan <- out.String()
            return
        } else if count == 0 {
            slog.Debug("Read zero count")
            outChan <- out.String()
            return
        }
        count, err = conn2.Write(buff)

        if err != nil {
            slog.Debug("Connection terminated", "error", err)
            outChan <- out.String()
            return
        } else if count == 0 {
            slog.Debug("Read zero count")
            outChan <- out.String()
            return
        }

        out.Write(buff)
    }
}

func main () {
    slog.Info("Starting proxy on :8080")

    ln, err := net.Listen("tcp", ":8080")
                    

    outChan := make(chan string)

    go func() {
        for {
            packet := <- outChan
            slog.Info("Packet", "packet", packet)
        }
    }()
    
    if err != nil {
        // handle error
    }
    for {
        conn, err := ln.Accept()
        if err != nil {
            // handle error
        }
        go func(connection net.Conn){
            slog.Debug("Recieved connection", "RemoteAddr", connection.RemoteAddr())

            clientGreeting, err := socks5.ParseClientGreeting(connection)

            if err != nil {
                slog.Error("Error occurred parsing greeting", "error", err)
                return
            }

            slog.Debug("Parsed client greeting", "clientGreeting", clientGreeting)

            if clientGreeting.Ver == 0x05 && slices.Contains(clientGreeting.Auth, 0x00) {
                slog.Debug("Can handle request")
                connection.Write(socks5.FormatServerChoice(&socks5.ServerChoice{
                    Ver: 0x05, 
                    Cauth: 0x00,
                }))
    
                connectRequest, err := socks5.ParseClientConnRequest(connection)

                if err != nil {
                    slog.Debug("Error occurred parsing conn request", "error", err)
                    return
                }

                slog.Debug("Parsed conn request", "request", connectRequest)

                conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", connectRequest.DstIp, connectRequest.DstPort))

                if err != nil {
                    slog.Debug("Couldn't connect to requested location", "error", err)
                    connection.Write(socks5.FormatConnResponse(&socks5.ServerConnResponse{
                        Ver: 0x05,
                        Status: 0x04,
                        Rsv: 0x00,
                        BndAddr: connectRequest.DstIp,
                        BndPort: connectRequest.DstPort,
                    }))
                } else {
                    slog.Debug("Connection proxied")
                    response := socks5.FormatConnResponse(&socks5.ServerConnResponse{
                        Ver: 0x05,
                        Status: 0x00,
                        Rsv: 0x00,
                        BndAddr: connectRequest.DstIp,
                        BndPort: connectRequest.DstPort,
                    })

                    connection.Write(response)
                    slog.Debug("Message parsed", "parsed", response)
                    
                    go serverToClient(conn, connection, outChan)
                    go serverToClient(connection, conn, outChan)
                }
            } else {
                slog.Debug("Cannot handle request")
                connection.Write(socks5.FormatServerChoice(&socks5.ServerChoice{
                    Ver: 0x05,
                    Cauth: 0xFF,
                }))
            }
        }(conn)
    }
}
