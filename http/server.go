package http

import (
	"log/slog"
	"net/http"
	"os"

	"com.github.redawl.mitmproxy/config"
	"com.github.redawl.mitmproxy/packet"
)

func ListenAndServe(conf config.Config, httpPacketHandler func(packet.Packet)) error {
    err := http.ListenAndServe(conf.HttpListenUri, Handler(conf, httpPacketHandler))

    slog.Error("Error serving http proxy server", "error", err)

    return err
}

func ListenAndServeTls(conf config.Config, httpPacketHandler func(packet.Packet)) error {
    userCfgDir, err := os.UserConfigDir()
    if err != nil {
        slog.Error("Error getting config dir", "error", err)
        return err
    }
    configDir := userCfgDir + "/mitmproxy"

    err = http.ListenAndServeTLS(conf.TlsListenUri, configDir + "/server.pem", configDir + "/privkey.pem", Handler(conf, httpPacketHandler))


    slog.Error("Error serving https proxy server", "error", err)
    return err
}
