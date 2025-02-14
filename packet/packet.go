package packet


type HttpPacket struct {
    ClientIp string
    ServerIp string
    Method string
    Status int
    Path string
    RespHeaders map[string][]string
    RespContent []byte
    ReqHeaders map[string][]string
    ReqContent []byte
}

func CreatePacket (
    clientIp string, 
    serverIp string, 
    method string,
    status int, 
    path string,
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
        RespHeaders: respHeaders,
        RespContent: respContent,
        ReqHeaders: reqHeaders,
        ReqContent: reqContent,
    }
}

