package config

import (
	"os"

	"fyne.io/fyne/v2"
)

type Config struct {
	SocksListenUri  string
	PacListenUri    string
	EnablePacServer bool
	Debug           bool
	CustomDecodings []string
	configDir       string
}

const (
	SOCKS_LISTEN_URI     = "socksListenUri"
	PAC_LISTEN_URI       = "pacListenUri"
	ENABLE_PAC_SERVER    = "enablePacServer"
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
		SocksListenUri:  stringWithFallbackSave(preferences, SOCKS_LISTEN_URI, "127.0.0.1:1080"),
		PacListenUri:    stringWithFallbackSave(preferences, PAC_LISTEN_URI, "127.0.0.1:8080"),
		EnablePacServer: boolWithFallbackSave(preferences, ENABLE_PAC_SERVER, false),
		Debug:           boolWithFallbackSave(preferences, ENABLE_DEBUG_LOGGING, false),
		CustomDecodings: preferences.StringList(CUSTOM_DECODINGS),
		configDir:       stringWithFallbackSave(preferences, CONFIGDIR, userCfgDir),
	}

	return conf
}
