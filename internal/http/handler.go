package http

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/redawl/gitm/internal/packet"
	"github.com/redawl/gitm/internal/util"
)

// Handler proxies requests from the socks5 proxy, to the requested server,
// and then forwards the response back to the socks5 proxy.
// httpPacketHandler is called twice; once with just the request information,
// and then a second time once the response information is available.
// Handler also handles the special host "gitm", which
func Handler(httpPacketHandler func(packet.HttpPacket), proxyUri string) http.HandlerFunc {
    return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
        hostName := request.Host
        slog.Debug("Handling request", "method", request.Method, "path", "http://" + hostName + request.URL.String(), "request", request)
        requestBody, err := io.ReadAll(request.Body)

        if err != nil {
            slog.Error("Error reading request responseBody", "error", err)
            return
        }

        // Special handling for /proxy.pac and /ca.crt
        if request.TLS == nil && request.Host == "gitm" {
            if request.URL.Path == "/ca.crt" {
                configDir, err := util.GetConfigDir()

                if err != nil {
                    slog.Error("Error getting config dir", "error", err)
                    return
                }

                certLocation := configDir + "/ca.crt"
                contents, err := os.ReadFile(certLocation)

                if err != nil {
                    slog.Error("Error getting ca cert", "error", err)
                    w.WriteHeader(http.StatusInternalServerError)
                    return
                }

                _, _ = w.Write(contents)
            } else if request.URL.Path == "/proxy.pac" {
                _, _ = fmt.Fprintf(w, "function FindProxyForURL(url, host){return \"SOCKS %s\";}", proxyUri)
            } else {
                http.Error(w, "Not found", http.StatusNotFound)
            }
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
            slog.Error("Error forwarding http request", "error", err, "request", req)

            // Attempt hijack to terminate the connection.
            hijack, ok := w.(http.Hijacker)

            if !ok {
                slog.Error("Webserver doesn't support hijacking, sending internal server error")
                http.Error(w, "Internal server error", http.StatusInternalServerError)
            }

            conn, _, _ := hijack.Hijack()

            conn.Close()
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

