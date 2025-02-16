package http

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"

	"com.github.redawl.mitmproxy/packet"
)

func Handler(httpPacketHandler func(packet.HttpPacket)) http.HandlerFunc {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        hostName := r.Host
        slog.Debug("Handling request", "method", r.Method, "path", "http://" + hostName + r.URL.String(), "request", r)
        requestBody, err := io.ReadAll(r.Body)

        if err != nil {
            slog.Error("Error reading request body", "error", err)
            return
        }

        req := &http.Request{
            Method: r.Method,
            URL: r.URL,
            Body:  io.NopCloser(bytes.NewReader(requestBody)),
            Proto: r.Proto,
            ProtoMajor: r.ProtoMajor,
            ProtoMinor: r.ProtoMinor,
            Header: r.Header,
        }

        if r.TLS != nil {
            req.URL.Scheme = "https"
        } else {
            req.URL.Scheme = "http"
        }

        req.URL.Host = r.Host
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

        body, err := io.ReadAll(resp.Body)

        if err != nil {
            slog.Error("Error forwarding http request", "error", err)
        }

        if httpPacketHandler != nil {
            httpPacketHandler(
                packet.CreatePacket(
                    r.RemoteAddr, 
                    r.Host, 
                    r.Method,
                    resp.Status, 
                    r.URL.Host + r.URL.RequestURI(), 
                    resp.Proto,
                    r.Proto,
                    resp.Header, 
                    body, 
                    r.Header,
                    requestBody,
                ),
            )
        }

        w.Write(body)
    })
}

