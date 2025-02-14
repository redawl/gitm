package http

import (
	"io"
	"log/slog"
	"net/http"

	"com.github.redawl.mitmproxy/config"
	"com.github.redawl.mitmproxy/packet"
)

func Handler(conf config.Config, httpPacketHandler func(packet.HttpPacket)) http.HandlerFunc {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        hostName := r.Host
        slog.Debug("Handling request", "method", r.Method, "path", "http://" + hostName + r.URL.String(), "request", r)

        req := &http.Request{
            Method: r.Method,
            URL: r.URL,
            Body: r.Body,
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
            slog.Error("Error forwarding http request", "error", err)
            return
        }

        // Set headers
        for header, value := range resp.Header {
            for _, v := range value {
                w.Header().Add(header, v)
            }
        }

        // Set status code
        w.WriteHeader(resp.StatusCode)
        
        slog.Debug("Response from proxied server", "response", resp)

        // Set body
        body, err := io.ReadAll(resp.Body)

        if err != nil {
            slog.Error("Error forwarding http request", "error", err)
        }

        if httpPacketHandler != nil {
            // TODO: Get request body correctly 
            slog.Info("Request", "r", r)
            httpPacketHandler(
                packet.CreatePacket(
                    r.RemoteAddr, 
                    r.Host, 
                    r.Method,
                    resp.StatusCode, 
                    r.URL.Host + r.URL.RequestURI(), 
                    resp.Header, 
                    body, 
                    r.Header,
                    []byte{},
                ),
            )
        }

        w.Write(body)
    })
}

