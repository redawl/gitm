package socks5

import (
	"slices"
)

const (
	METHOD_NO_AUTH_REQUIRED  = 0x00
	METHOD_GSSAPI            = 0x01
	METHOD_USERNAME_PASSWORD = 0x02
	// 0x03 to 0x7F are IANA assigned
	// 0x80 to 0xFE are reserved for private methods
	METHOD_NO_ACCEPTABLE_METHODS = 0xFF
)

const (
	SOCKS_VER_5 = 0x05
)

const (
	STATUS_SUCCEEDED                  = 0x00
	STATUS_GENERAL_FAILURE            = 0x01
	STATUS_CONNECTION_NOT_ALLOWED     = 0x02
	STATUS_NETWORK_UNREACHABLE        = 0x03
	STATUS_HOST_UNREACHABLE           = 0x04
	STATUS_CONNECTION_REFUSED         = 0x05
	STATUS_TTL_EXPIRED                = 0x06
	STATUS_COMMAND_NOT_SUPPORTED      = 0x07
	STATUS_ADDRESS_TYPE_NOT_SUPPORTED = 0x08
	// 0x09 to 0xFF unassigned
)

const (
	ADDRESS_TYPE_IPV4       = 0x01
	ADDRESS_TYPE_DOMAINNAME = 0x03
	ADDRESS_TYPE_IPV6       = 0x04
)

const (
	CMD_CONNECT       = 0x01
	CMD_BIND          = 0x02
	CMD_UDP_ASSOCIATE = 0x03
)

type ClientGreeting struct {
	Ver   byte
	Nauth uint8
	Auth  []byte
}

func (greeting *ClientGreeting) CanHandle() bool {
	return greeting.Ver == SOCKS_VER_5 &&
		slices.Contains(greeting.Auth, METHOD_NO_AUTH_REQUIRED)
}

type ClientConnRequest struct {
	Ver       byte
	Cmd       byte
	Rsv       byte
	DstIpType byte
	DstIp     string
	DstPort   uint16
}
