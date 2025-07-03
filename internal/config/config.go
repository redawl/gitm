package config

import (
	"os"

	"fyne.io/fyne/v2"
)

type Config struct {
	HttpListenUri   string
	TlsListenUri    string
	SocksListenUri  string
	Debug           bool
	CustomDecodings []string
	configDir       string
}

const (
	HTTP_LISTEN_URI      = "httpListenUri"
	TLS_LISTEN_URI       = "tlsListenUri"
	SOCKS_LISTEN_URI     = "socksListenUri"
	ENABLE_DEBUG_LOGGING = "enableDebugLogging"
	CUSTOM_DECODINGS     = "customDecodings"
	CONFIGDIR            = "configDir"
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

	if !value {
		prefs.SetBool(key, defaultValue)
		return defaultValue
	}

	return value
}

// FromPreferences creates a new config from fyne preferences, and if any are missing saves the default values
// to the fyne preferences object
func FromPreferences(preferences fyne.Preferences) Config {
	userCfgDir, err := os.UserConfigDir()
	if err != nil {
		userCfgDir = ""
	}

	userCfgDir = userCfgDir + "/gitm"
	conf := Config{
		HttpListenUri:   stringWithFallbackSave(preferences, HTTP_LISTEN_URI, "127.0.0.1:8080"),
		TlsListenUri:    stringWithFallbackSave(preferences, TLS_LISTEN_URI, "127.0.0.1:5443"),
		SocksListenUri:  stringWithFallbackSave(preferences, SOCKS_LISTEN_URI, "127.0.0.1:1080"),
		Debug:           boolWithFallbackSave(preferences, ENABLE_DEBUG_LOGGING, false),
		CustomDecodings: preferences.StringList(CUSTOM_DECODINGS),
		configDir:       stringWithFallbackSave(preferences, CONFIGDIR, userCfgDir),
	}

	return conf
}
