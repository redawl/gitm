package packet

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"
)

// HttpPacket represents a captured packet from either the https or http proxy.
// An HttpPacket contains all the information from the http request, as well as the information from the http response (once it has been captured).
type HttpPacket struct {
	TimeStamp time.Time
	Id        [16]byte `json:"id"`
	Hostname  string
	Method    string
	Status    string
	Path      string
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
		TimeStamp:   time.Now(),
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

func (p *HttpPacket) FormatRequestContent() string {
	return fmt.Sprintf(
		"%s %s %s\n%s\n%s",
		p.Method,
		p.Path,
		p.ReqProto,
		formatHeaders(p.ReqHeaders),
		decodeBody(p.ReqBody, p.ReqHeaders["Content-Encoding"]),
	)
}

func (p *HttpPacket) FormatResponseContent() string {
	return fmt.Sprintf(
		"%s %s\n%s\n%s",
		p.RespProto,
		p.Status,
		formatHeaders(p.RespHeaders),
		decodeBody(p.RespBody, p.RespHeaders["Content-Encoding"]),
	)
}

func formatHeaders(headers map[string][]string) string {
	builder := strings.Builder{}

	for header, values := range headers {
		builder.WriteString(header)
		builder.WriteString(": ")
		for i, value := range values {
			builder.WriteString(value)
			if i != len(values)-1 {
				builder.WriteString(", ")
			}
		}
		builder.WriteString("\n")
	}

	return builder.String()
}

func decodeBody(body []byte, contentEncodings []string) string {
	ret := body
	if len(contentEncodings) > 0 {
		decoded := bytes.NewReader(body)
		for _, contentEncoding := range contentEncodings {
			switch contentEncoding {
			case "gzip":
				{
					decoded, err := gzip.NewReader(decoded)
					if err != nil {
						slog.Error("Failed decoding gzip", "error", err)
						break
					}

					ret, err = io.ReadAll(decoded)
					if err != nil {
						slog.Error("Failed reading stream", "error", err)
						break
					}
				}
			case "deflate":
				{
					decoded := flate.NewReader(decoded)

					var err error
					ret, err = io.ReadAll(decoded)
					if err != nil {
						slog.Error("Failed reading stream", "error", err)
						break
					}
				}
			case "UTF-8":
			case "none":
			default:
				{
					slog.Error("Unhandled compression", "compression", contentEncoding)
					break
				}
			}
		}
	}

	if json.Valid(ret) {
		buff := new(bytes.Buffer)
		err := json.Indent(buff, ret, "", "    ")

		if err != nil {
			slog.Error("Failed indenting json", "error", err)
		} else {
			ret = buff.Bytes()
		}
	}

	return string(ret)
}
