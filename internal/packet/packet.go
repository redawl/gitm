package packet

import (
	"crypto/rand"
	"log/slog"
)

// HttpPacket represents a captured packet from either the https or http proxy.
// An HttpPacket contains all the information from the http request, as well as the information from the http response (once it has been captured).
type HttpPacket struct {
	Id       [16]byte `json:"id"`
	Hostname string
	Method   string
	Status   string
	Path     string
	// The request protocol version, i.e. "HTTP/1.1"
	// The response protocol version, i.e. "HTTP/1.1"
	ReqProto    string
	RespProto   string
	RespHeaders map[string][]string
	RespBody    []byte
	ReqHeaders  map[string][]string
	ReqBody     []byte
}

func CreatePacket(
	hostname string,
	method string,
	status string,
	path string,
	respProto string,
	reqProto string,
	respHeaders map[string][]string,
	respBody []byte,
	reqHeaders map[string][]string,
	reqBody []byte,
) HttpPacket {
	packet := HttpPacket{
		Hostname:    hostname,
		Method:      method,
		Status:      status,
		Path:        path,
		ReqProto:    reqProto,
		RespProto:   respProto,
		RespHeaders: respHeaders,
		RespBody:    respBody,
		ReqHeaders:  reqHeaders,
		ReqBody:     reqBody,
	}

	if _, err := rand.Read(packet.Id[:]); err != nil {
		slog.Error("Error generating id", "error", err)
	}

	return packet
}

func FindPacket(toFind *HttpPacket, packets []*HttpPacket) *HttpPacket {
	for _, p := range packets {
		if toFind.Id == p.Id {
			return p
		}
	}

	return nil
}

func (p *HttpPacket) UpdatePacket(inPacket *HttpPacket) {
	p.Hostname = inPacket.Hostname
	p.Method = inPacket.Method
	p.Status = inPacket.Status
	p.Path = inPacket.Path
	p.RespProto = inPacket.RespProto
	p.ReqProto = inPacket.ReqProto
	p.RespHeaders = inPacket.RespHeaders
	p.RespBody = inPacket.RespBody
	p.ReqHeaders = inPacket.ReqHeaders
	p.ReqBody = inPacket.ReqBody
}
