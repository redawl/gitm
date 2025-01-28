package socks5

type ClientGreeting struct {
    Ver   byte
    Nauth uint8
    Auth  []byte
}

type ClientConnRequest struct {
    Ver       byte
    Cmd       byte
    Rsv       byte
    DstIpType byte
    DstIp     string
    DstPort   uint16
}

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

