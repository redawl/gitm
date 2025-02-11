package packet

type HttpPacket struct {
    ClientIp string   `json:"src_ip"`
    ServerIp string   `json:"dst_ip"`
    Status int
    Path string
    RespHeaders map[string][]string
    RespContent []byte
    ReqHeaders map[string][]string
    ReqContent []byte
}

func CreatePacket (clientIp string, serverIp string, status int, path string, respHeaders map[string][]string, respContent []byte, reqHeaders map[string][]string, reqContent []byte) (HttpPacket) {
    return HttpPacket{
        ClientIp: clientIp,
        ServerIp: serverIp,
        Status: status,
        Path: path,
        RespHeaders: respHeaders,
        RespContent: respContent,
        ReqHeaders: reqHeaders,
        ReqContent: reqContent,
    }
}

