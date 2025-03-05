package http

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"

	"github.com/redawl/gitm/internal/packet"
)

func Handler(httpPacketHandler func(packet.HttpPacket)) http.HandlerFunc {
    return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
        hostName := request.Host
        slog.Debug("Handling request", "method", request.Method, "path", "http://" + hostName + request.URL.String(), "request", request)
        requestBody, err := io.ReadAll(request.Body)

        if err != nil {
            slog.Error("Error reading request responseBody", "error", err)
            return
        }

        req := &http.Request{
            Method: request.Method,
            URL: request.URL,
            Body:  io.NopCloser(bytes.NewReader(requestBody)),
            Proto: request.Proto,
            ProtoMajor: request.ProtoMajor,
            ProtoMinor: request.ProtoMinor,
            Header: request.Header,
        }

        if request.TLS != nil {
            req.URL.Scheme = "https"
        } else {
            req.URL.Scheme = "http"
        }

        req.URL.Host = request.Host

        httpPacket := packet.CreatePacket(
            request.URL.Hostname(), 
            request.Method,
            "", 
            request.URL.RequestURI(), 
            "",
            request.Proto,
            nil, 
            nil, 
            request.Header,
            requestBody,
        )

        if httpPacketHandler != nil {
            httpPacketHandler(httpPacket)
        }
        resp, err := http.DefaultTransport.RoundTrip(req)

        if err != nil {
            // TODO: What to do here
            slog.Error("Error forwarding http request", "error", err, "request", req)
            return
        }

        for header, value := range resp.Header {
            for _, v := range value {
                w.Header().Add(header, v)
            }
        }

        w.WriteHeader(resp.StatusCode)
        
        slog.Debug("Response from proxied server", "response", resp)

        responseBody, err := io.ReadAll(resp.Body)

        if err != nil {
            slog.Error("Error forwarding http request", "error", err)
            return
        }

        if httpPacketHandler != nil {
            completedPacket := packet.CreatePacket(
                request.URL.Hostname(), 
                request.Method,
                resp.Status, 
                request.URL.RequestURI(), 
                resp.Proto,
                request.Proto,
                resp.Header, 
                responseBody, 
                request.Header,
                requestBody,
            )
            httpPacket.UpdatePacket(&completedPacket)
            httpPacketHandler(httpPacket)
        }

        w.Write(responseBody)
    })
}

