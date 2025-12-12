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

	"github.com/redawl/gitm/internal"
)

var _ Packet = (*HTTPPacket)(nil)

const (
	FilterHostname = "hostname"
	FilterMethod   = "method"
	FilterPath     = "path"
	FilterReqBody  = "reqbody"
	// TODO filter on version?
	FilterStatus   = "status"
	FilterRespBody = "respbody"
)

// HTTPPacket represents a captured packet from either the https or http proxy.
// An HTTPPacket contains all the information from the http request, as well as the information from the http response (once it has been captured).
type HTTPPacket struct {
	Encrypted_ bool      `json:"Encrypted"`
	TimeStamp_ time.Time `json:"TimeStamp"`
	Type_      string    `json:"Type"`
	ID         [16]byte  `json:"id"`
	Hostname   string
	Method     string
	Status     string
	Path       string
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
	encrypted bool,
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
) HTTPPacket {
	packet := HTTPPacket{
		Encrypted_:  encrypted,
		TimeStamp_:  time.Now(),
		Type_:       "http",
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

	if _, err := rand.Read(packet.ID[:]); err != nil {
		slog.Error("Error generating id", "error", err)
	}

	return packet
}

func (p *HTTPPacket) Encrypted() bool {
	return p.Encrypted_
}

func (p *HTTPPacket) TimeStamp() time.Time {
	return p.TimeStamp_
}

func (p *HTTPPacket) Type() string {
	return p.Type_
}

func (p *HTTPPacket) FindPacket(packets []Packet) Packet {
	for _, pac := range packets {
		if httpPacket, ok := pac.(*HTTPPacket); ok && httpPacket.ID == p.ID {
			return p
		}
	}

	return nil
}

func (p *HTTPPacket) FormatHostname() string {
	return p.Hostname
}

func (p *HTTPPacket) FormatRequestLine() string {
	return fmt.Sprintf("%s %s %s", p.Method, p.Path, p.ReqProto)
}

func (p *HTTPPacket) FormatResponseLine() string {
	return fmt.Sprintf("%s %s", p.RespProto, p.Status)
}

func (p *HTTPPacket) MatchesFilter(tokens []internal.FilterToken) bool {
	for _, token := range tokens {
		filterStr := ""
		switch token.FilterType {
		case FilterHostname:
			filterStr = p.Hostname
		case FilterMethod:
			filterStr = p.Method
		case FilterPath:
			filterStr = p.Path
		case FilterReqBody:
			filterStr = string(p.ReqBody)
		case FilterStatus:
			filterStr = p.Status
		case FilterRespBody:
			filterStr = string(p.RespBody)
		default:
			slog.Warn("Unknown filter specified", "filterType", token.FilterType, "filterContent", token.FilterContent)
		}

		if token.Negate == strings.Contains(filterStr, token.FilterContent) {
			return false
		}
	}

	return true
}

func (p *HTTPPacket) UpdatePacket(inPacket Packet) {
	if httpPacket, ok := inPacket.(*HTTPPacket); ok {
		p.Hostname = httpPacket.Hostname
		p.Method = httpPacket.Method
		p.Status = httpPacket.Status
		p.Path = httpPacket.Path
		p.RespProto = httpPacket.RespProto
		p.ReqProto = httpPacket.ReqProto
		p.RespHeaders = httpPacket.RespHeaders
		p.RespBody = httpPacket.RespBody
		p.ReqHeaders = httpPacket.ReqHeaders
		p.ReqBody = httpPacket.ReqBody
	}
}

func (p *HTTPPacket) FormatRequestContent() string {
	return fmt.Sprintf(
		"%s %s %s\n%s\n%s",
		p.Method,
		p.Path,
		p.ReqProto,
		formatHeaders(p.ReqHeaders),
		decodeBody(p.ReqBody, p.ReqHeaders["Content-Encoding"]),
	)
}

func (p *HTTPPacket) FormatResponseContent() string {
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
		builder.WriteByte('\n')
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
				decoded, err := gzip.NewReader(decoded)
				if err != nil {
					slog.Error("Failed decoding gzip", "error", err)
					return string(body)
				}

				ret, err = io.ReadAll(decoded)
				if err != nil {
					slog.Error("Failed reading stream", "error", err)
					return string(body)
				}
			case "deflate":
				decoded := flate.NewReader(decoded)

				var err error
				ret, err = io.ReadAll(decoded)
				if err != nil {
					slog.Error("Failed reading stream", "error", err)
					return string(body)
				}
			case "br":
				slog.Error("Brotli support will be added in the future (most likely)")
			case "UTF-8":
			case "none":
			default:
				slog.Error("Unhandled compression", "compression", contentEncoding)
				return string(body)
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
