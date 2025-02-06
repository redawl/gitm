package http

import (
	"io"
	"log/slog"
	"net/http"

	"com.github.redawl.mitmproxy/config"
	"com.github.redawl.mitmproxy/packet"
)

func Handler(conf config.Config, httpPacketHandler func(packet.Packet)) http.HandlerFunc {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        hostName := r.Host
        slog.Debug("Handling request", "method", r.Method, "path", "http://" + hostName + r.URL.String(), "request", r)

        req, err := http.NewRequest(r.Method, "http://" + hostName + r.URL.String(), r.Body)
        if err != nil {
            // TODO: What to do here
            slog.Error("Error forwarding http request", "error", err)
            return
        }

        // Set headers
        for header, value := range r.Header {
            for _, v := range value {
                req.Header.Add(header, v)
            }
        }

        resp, err := http.DefaultTransport.RoundTrip(req)

        if err != nil {
            // TODO: What to do here
            slog.Error("Error forwarding http request", "error", err)
            return
        }

        
        // Set status code
        w.WriteHeader(resp.StatusCode)
        // Set headers
        for header, value := range resp.Header {
            for _, v := range value {
                w.Header().Add(header, v)
            }
        }
        
        slog.Debug("Response from proxied server", "response", resp)

        // Set body
        body, err := io.ReadAll(resp.Body)

        if err != nil {
            slog.Error("Error forwarding http request", "error", err)
        }

        if httpPacketHandler != nil {
            httpPacketHandler(
                packet.CreatePacket(r.RemoteAddr, r.RemoteAddr, resp.StatusCode, resp.Header, body),
            )

        }

        w.Write(body)
    })
}
