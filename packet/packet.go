package packet

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"log/slog"
)

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
    encoding := respHeaders["Content-Encoding"]
    contentType := respHeaders["Content-Type"][0]
    rContent := respContent
    if len(encoding) > 0 {
        switch(encoding[0]) {
            case "gzip": {
                reader, err := gzip.NewReader(bytes.NewReader(rContent))

                if err != nil {
                    slog.Error("Error decompressing gzip compressed data (NewReader)", "error", err)
                    rContent = respContent
                } else {
                    defer reader.Close()
                    rContent, err = io.ReadAll(reader)

                    if err != nil && err != io.EOF {
                        slog.Error("Error decompressing gzip compressed data (ReadAll)", "error", err)
                        rContent = respContent
                    }
                }
            }
        }
    }

    if contentType == "application/json" {
        reader := bytes.NewBuffer([]byte{})
        err := json.Indent(reader, rContent, "", "\n")

        if err != nil {
            slog.Error("Error formatting json", "error", err)
        } else {
            rContent = reader.Bytes()
        }
    }
    return HttpPacket{
        ClientIp: clientIp,
        ServerIp: serverIp,
        Status: status,
        Path: path,
        RespHeaders: respHeaders,
        RespContent: rContent,
        ReqHeaders: reqHeaders,
        ReqContent: reqContent,
    }
}

