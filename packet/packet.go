package packet


type HttpPacket struct {
    ClientIp string
    ServerIp string
    Method string
    Status string
    Path string
    ReqProto string
    RespProto string
    RespHeaders map[string][]string
    RespContent []byte
    ReqHeaders map[string][]string
    ReqContent []byte
}

func CreatePacket (
    clientIp string, 
    serverIp string, 
    method string,
    status string, 
    path string,
    respProto string,
    reqProto string,
    respHeaders map[string][]string, 
    respContent []byte, 
    reqHeaders map[string][]string, 
    reqContent []byte,
) (HttpPacket) {
    return HttpPacket{
        ClientIp: clientIp,
        ServerIp: serverIp,
        Method: method,
        Status: status,
        Path: path,
        ReqProto: reqProto,
        RespProto: respProto,
        RespHeaders: respHeaders,
        RespContent: respContent,
        ReqHeaders: reqHeaders,
        ReqContent: reqContent,
    }
}

