package packet

type HttpPacket struct {
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

