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

var _ Packet = (*HttpPacket)(nil)

const (
	FILTER_HOSTNAME = "hostname"
	FILTER_METHOD   = "method"
	FILTER_PATH     = "path"
	FILTER_REQ_BODY = "reqbody"
	// TODO filter on version?
	FILTER_STATUS    = "status"
	FILTER_RESP_BODY = "respbody"
)

// HttpPacket represents a captured packet from either the https or http proxy.
// An HttpPacket contains all the information from the http request, as well as the information from the http response (once it has been captured).
type HttpPacket struct {
	TimeStamp_ time.Time `json:"TimeStamp"`
	Type_      string    `json:"Type"`
	Id         [16]byte  `json:"id"`
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

	if _, err := rand.Read(packet.Id[:]); err != nil {
		slog.Error("Error generating id", "error", err)
	}

	return packet
}

func (p *HttpPacket) TimeStamp() time.Time {
	return p.TimeStamp_
}

func (p *HttpPacket) Type() string {
	return p.Type_
}

func (p *HttpPacket) FindPacket(packets []Packet) Packet {
	for _, pac := range packets {
		if httpPacket, ok := pac.(*HttpPacket); ok && httpPacket.Id == p.Id {
			return p
		}
	}

	return nil
}

func (p *HttpPacket) FormatHostname() string {
	return p.Hostname
}

func (p *HttpPacket) FormatRequestLine() string {
	return fmt.Sprintf("%s %s %s", p.Method, p.Path, p.ReqProto)
}

func (p *HttpPacket) FormatResponseLine() string {
	return fmt.Sprintf("%s %s", p.RespProto, p.Status)
}

func (p *HttpPacket) MatchesFilter(tokens []internal.FilterToken) bool {
	for _, token := range tokens {
		filterStr := ""
		switch token.FilterType {
		case FILTER_HOSTNAME:
			filterStr = p.Hostname
		case FILTER_METHOD:
			filterStr = p.Method
		case FILTER_PATH:
			filterStr = p.Path
		case FILTER_REQ_BODY:
			filterStr = string(p.ReqBody)
		case FILTER_STATUS:
			filterStr = p.Status
		case FILTER_RESP_BODY:
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

func (p *HttpPacket) UpdatePacket(inPacket Packet) {
	if httpPacket, ok := inPacket.(*HttpPacket); ok {
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
				decoded, err := gzip.NewReader(decoded)
				if err != nil {
					slog.Error("Failed decoding gzip", "error", err)
				}

				ret, err = io.ReadAll(decoded)
				if err != nil {
					slog.Error("Failed reading stream", "error", err)
				}
			case "deflate":
				decoded := flate.NewReader(decoded)

				var err error
				ret, err = io.ReadAll(decoded)
				if err != nil {
					slog.Error("Failed reading stream", "error", err)
				}
			case "UTF-8":
			case "none":
			default:
				slog.Error("Unhandled compression", "compression", contentEncoding)
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
