package socks5

import (
	"net"
	"slices"
	"testing"
)

func TestFormatServerChoice(t *testing.T) {
	expected := []byte{0x01, 0x02}
	actual := FormatServerChoice(0x01, 0x02)
	if !slices.Equal(expected, actual) {
		t.Errorf("FormatServerChoice(0x01, 0x02) = %x, want %x", actual, expected)
	}
}

func TestFormatConnResponseSuccess(t *testing.T) {
	expected := []byte{0x05, 0x00, 0x00, 0x01, 0x01, 0x01, 0x01, 0x01, 0x00, 0x50}

	actual := FormatConnResponse(
		SocksVer5,
		StatusSucceeded,
		&net.TCPAddr{
			IP:   net.IPv4(0x01, 0x01, 0x01, 0x01),
			Port: 80,
		},
	)

	if !slices.Equal(expected, actual) {
		t.Errorf("FormatConnResponse(...) = %x, want %x", actual, expected)
	}
}
