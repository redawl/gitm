package packet

import (
	"github.com/google/uuid"
)

type HttpPacket struct {
    id uuid.UUID
    Hostname string
    Method string
    Status string
    Path string
    ReqProto string
    RespProto string
    RespHeaders map[string][]string
    RespBody []byte
    ReqHeaders map[string][]string
    ReqBody []byte
}

func CreatePacket (
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
) (HttpPacket) {
    return HttpPacket{
        id: uuid.New(),
        Hostname: hostname,
        Method: method,
        Status: status,
        Path: path,
        ReqProto: reqProto,
        RespProto: respProto,
        RespHeaders: respHeaders,
        RespBody: respBody,
        ReqHeaders: reqHeaders,
        ReqBody: reqBody,
    }
}

func FindPacket(toFind *HttpPacket, packets []*HttpPacket) *HttpPacket {
    for _, p := range packets {
        if toFind.id == p.id {
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

