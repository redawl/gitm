package socks5

import (
	"slices"
)

const (
	MethodNoAuthRequired   = 0x00
	MethodGSSAPI           = 0x01
	MethodUsernamePassword = 0x02
	// 0x03 to 0x7F are IANA assigned
	// 0x80 to 0xFE are reserved for private methods
	MethodNoAcceptableMethods = 0xFF
)

const (
	SocksVer5 = 0x05
)

const (
	StatusSucceeded               = 0x00
	StatusGeneralFailure          = 0x01
	StatusConnectionNotAllowed    = 0x02
	StatusNetworkUnreachable      = 0x03
	StatusHostUnreachable         = 0x04
	StatusConnectionRefused       = 0x05
	StatusTTLExpired              = 0x06
	StatusCommandNotSupported     = 0x07
	StatusAddressTypeNotSupported = 0x08
	// 0x09 to 0xFF unassigned
)

const (
	AddressTypeIPv4       = 0x01
	AddressTypeDomainName = 0x03
	AddressTypeIPv6       = 0x04
)

const (
	CmdConnect      = 0x01
	CmdBind         = 0x02
	CmdUDPAssociate = 0x03
)

type ClientGreeting struct {
	Ver   byte
	Nauth uint8
	Auth  []byte
}

func (g *ClientGreeting) CanHandle() bool {
	return g.Ver == SocksVer5 &&
		slices.Contains(g.Auth, MethodNoAuthRequired)
}

type ClientConnRequest struct {
	Ver       byte
	Cmd       byte
	Rsv       byte
	DstIPType byte
	DstIP     string
	DstPort   uint16
}
