package httputils

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/textproto"
	"strconv"
	"strings"
	"time"

	"github.com/redawl/gitm/internal/packet"
	"github.com/redawl/gitm/internal/util"
)

// HandleHttpRequest reads http requests from inboundConn to outboundConn,
// and then read http responses from outboundConn to inboundConn.
//
// httpPacketHandler is called first on the packet when inboundConn -> outboundConn completes,
// and again when outboundConn -> inboundConn completes.
func HandleHttpRequest(inboundConn, outboundConn net.Conn, httpPacketHandler func(packet.Packet)) error {
	bufReader := bufio.NewReader(io.TeeReader(inboundConn, outboundConn))
	reader := textproto.NewReader(bufReader)
	clientBufioReader := bufio.NewReader(io.TeeReader(outboundConn, inboundConn))
	clientReader := textproto.NewReader(clientBufioReader)

	method, uri, proto, err := ReadLine1(reader)
	if err != nil {
		return fmt.Errorf("http request line1: %w", err)
	}
	headers, err := reader.ReadMIMEHeader()
	if err != nil {
		return fmt.Errorf("http request header: %w", err)
	}

	requestBody, err := readBody(headers, bufReader)
	if err != nil {
		return fmt.Errorf("http request body: %w", err)
	}

	httpPacket := packet.CreatePacket(
		headers.Get("Host"),
		method,
		"",
		uri,
		"",
		proto,
		nil,
		nil,
		http.Header(headers),
		requestBody,
	)

	if headers.Get("Upgrade") != "websocket" {
		go httpPacketHandler(&httpPacket)
	}

	respProto, statusCode, statusCodeMessage, err := ReadLine1(clientReader)
	if err != nil {
		return fmt.Errorf("http response line1: %w", err)
	}

	responseHeaders, err := clientReader.ReadMIMEHeader()
	if err != nil {
		return fmt.Errorf("http response header: %w", err)
	}

	responseBody, err := readBody(responseHeaders, clientBufioReader)
	if err != nil {
		return fmt.Errorf("http response body: %w", err)
	}

	completedPacket := packet.CreatePacket(
		headers.Get("Host"),
		method,
		fmt.Sprintf("%s %s", statusCode, statusCodeMessage),
		uri,
		respProto,
		proto,
		http.Header(responseHeaders),
		responseBody,
		http.Header(headers),
		requestBody,
	)

	httpPacket.UpdatePacket(&completedPacket)

	if headers.Get("Upgrade") == "websocket" {
		done := make(chan bool)
		p := packet.CreateWebsocketPacket(httpPacket)
		httpPacketHandler(p)
		go func() {
			for {
				if err := handleWebsocket(bufReader, p.AddClientFrame); err != nil {
					if !errors.Is(err, io.EOF) {
						slog.Error("Error handling websocket", "error", err)
					}
					break
				}
				go httpPacketHandler(p)
			}
			done <- true
		}()
		go func() {
			for {
				if err := handleWebsocket(clientBufioReader, p.AddServerFrame); err != nil {
					if !errors.Is(err, io.EOF) {
						slog.Error("Error handling websocket", "error", err)
					}
					break
				}
				go httpPacketHandler(p)
			}
			done <- true
		}()
		<-done
		<-done
		return nil
	}

	return nil
}

func ReadLine1(reader *textproto.Reader) (string, string, string, error) {
	line1, err := reader.ReadLine()
	if err != nil {
		return "", "", "", err
	}

	line1Parts := strings.Split(line1, " ")
	return line1Parts[0], line1Parts[1], strings.Join(line1Parts[2:], " "), nil
}

func readBody(headers textproto.MIMEHeader, reader io.Reader) ([]byte, error) {
	logger := slog.With("hostname", headers.Get("Host"))
	transferEncoding := headers.Get("Transfer-Encoding")

	if transferEncoding != "" {
		// TODO: handle other encodings
		// For now, handle chunked only
		if transferEncoding == "chunked" {
			return io.ReadAll(httputil.NewChunkedReader(reader))
		} else {
			logger.Error("Not handling unknown Transfer-Encoding", "encoding", transferEncoding)
		}
	}

	contentLengthHeader := headers.Get("Content-Length")

	if contentLengthHeader == "" {
		return []byte{}, nil
	}
	contentLength, err := strconv.Atoi(contentLengthHeader)
	if err != nil {
		return nil, err
	}

	if contentLength == 0 {
		return []byte{}, nil
	}
	bytes := make([]byte, contentLength)
	limitReader := io.LimitReader(reader, int64(contentLength))
	if num, err := io.ReadAtLeast(limitReader, bytes, contentLength); err != nil {
		return nil, fmt.Errorf("reading body: %w", err)
	} else {
		if num != contentLength {
			return nil, fmt.Errorf("expected read of %d, got %d", contentLength, num)
		}
	}

	return bytes, nil
}

func handleWebsocket(reader *bufio.Reader, frameHandler func(*packet.WebsocketFrame)) error {
	header, err := util.ReadCount(reader, 2)
	if err != nil {
		return fmt.Errorf("header: %w", err)
	}
	byte1 := header[0]
	byte2 := header[1]

	fin := byte1 >> 7
	rsv1 := byte1 >> 6 & 0x01
	rsv2 := byte1 >> 5 & 0x01
	rsv3 := byte1 >> 4 & 0x01
	opcode := byte1 & 0x0F

	masked := byte2 >> 7
	payloadLength := uint64(byte2 & 0x7F)
	var payloadLengthBytes []byte
	switch payloadLength {
	case 126:
		payloadLengthBytes, err = util.ReadCount(reader, 2)
		if err != nil {
			return fmt.Errorf("length(16 bit): %w", err)
		}
		payloadLength = uint64(binary.BigEndian.Uint16(payloadLengthBytes))
	case 127:
		payloadLengthBytes, err = util.ReadCount(reader, 8)
		if err != nil {
			return fmt.Errorf("length(64 bit): %w", err)
		}

		payloadLength = binary.BigEndian.Uint64(payloadLengthBytes)
	}
	maskingKey := [4]byte{}
	if masked == 1 {
		maskingKeyBytes, err := util.ReadCount(reader, 4)
		if err != nil {
			return fmt.Errorf("masking key: %w", err)
		}
		maskingKey = [4]byte(maskingKeyBytes)
	}

	bodyBytes, err := util.ReadCount(reader, int(payloadLength))
	if err != nil {
		return fmt.Errorf("payload: %w", err)
	}

	frameHandler(
		&packet.WebsocketFrame{
			TimeStamp:     time.Now(),
			Fin:           fin != 0,
			RSV1:          rsv1 != 0,
			RSV2:          rsv2 != 0,
			RSV3:          rsv3 != 0,
			Opcode:        opcode,
			Masked:        masked != 0,
			PayloadLength: payloadLength,
			MaskingKey:    maskingKey,
			Payload:       bodyBytes,
		},
	)

	return nil
}
