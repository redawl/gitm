package httputils

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/textproto"
	"strconv"
	"strings"

	"github.com/redawl/gitm/internal/packet"
)

// HandleHttpRequest reads http requests from inboundConn to outboundConn,
// and then read http responses from outboundConn to inboundConn.
//
// httpPacketHandler is called first on the packet when inboundConn -> outboundConn completes,
// and again when outboundConn -> inboundConn completes.
//
// TODO: Handle websockets
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

	go httpPacketHandler(&httpPacket)

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

	go httpPacketHandler(&httpPacket)

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
			logger.Debug("Handling chunking encoding request")
			return io.ReadAll(httputil.NewChunkedReader(reader))
		} else {
			logger.Error("Not handling unknown Transfer-Encoding", "encoding", transferEncoding)
		}
	}

	logger.Debug("Handling normal encoding request", "headers", headers)
	contentLengthHeader := headers.Get("Content-Length")

	if contentLengthHeader == "" {
		return []byte{}, nil
	}
	contentLength, err := strconv.Atoi(contentLengthHeader)
	if err != nil {
		return nil, err
	}
	logger.Debug("About to read data", "Content-Length", contentLength)

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
