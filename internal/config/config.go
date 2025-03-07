package config

import (
	"fyne.io/fyne/v2"
)

type Config struct {
    HttpListenUri string
    TlsListenUri string
    CacertListenUri string
    SocksListenUri string
    Debug bool
}

const (
    HTTP_LISTEN_URI = "httpListenUri"
    TLS_LISTEN_URI  = "tlsListenUri"
    CACERT_LISTEN_URI = "cacertListenUri"
    SOCKS_LISTEN_URI  = "socksListenUri"
    ENABLE_DEBUG_LOGGING = "enableDebugLogging"
)

func stringWithFallbackSave(prefs fyne.Preferences, key string, defaultValue string) string {
    value := prefs.String(key)

    if value == "" {
        prefs.SetString(key, defaultValue)
        return defaultValue
    }

    return value
}

func boolWithFallbackSave(prefs fyne.Preferences, key string, defaultValue bool) bool {
    value := prefs.Bool(key)

    if value == false {
        prefs.SetBool(key, defaultValue)
        return defaultValue
    }

    return value
}

func ParseFlags (preferences fyne.Preferences) Config {
    conf := Config{
        HttpListenUri: stringWithFallbackSave(preferences, HTTP_LISTEN_URI, "127.0.0.1:8080"), 
        TlsListenUri: stringWithFallbackSave(preferences, TLS_LISTEN_URI, "127.0.0.1:5443"),
        CacertListenUri: stringWithFallbackSave(preferences, CACERT_LISTEN_URI, "127.0.0.1:9090"),
        SocksListenUri: stringWithFallbackSave(preferences, SOCKS_LISTEN_URI, "127.0.0.1:1080"),
        Debug: boolWithFallbackSave(preferences, ENABLE_DEBUG_LOGGING, false),
    }

    return conf
}

