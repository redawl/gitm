package internal

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
)

type Config struct {
	SocksListenURI  string
	PACListenURI    string
	EnablePACServer bool
	Debug           bool
	CustomDecodings []string
	configDir       string
	Theme           string
}

const (
	SocksListenURI     = "socksListenUri"
	PACListenURI       = "pacListenUri"
	EnablePACServer    = "enablePacServer"
	EnableDebugLogging = "enableDebugLogging"
	CustomDecodings    = "customDecodings"
	ConfigDir          = "configDir"
	Theme              = "customTheme"
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

	userCfgDir = filepath.Join(userCfgDir, "gitm")
	conf := Config{
		SocksListenURI:  stringWithFallbackSave(preferences, SocksListenURI, "127.0.0.1:1080"),
		PACListenURI:    stringWithFallbackSave(preferences, PACListenURI, "127.0.0.1:8080"),
		EnablePACServer: boolWithFallbackSave(preferences, EnablePACServer, false),
		Debug:           boolWithFallbackSave(preferences, EnableDebugLogging, false),
		CustomDecodings: preferences.StringList(CustomDecodings),
		configDir:       stringWithFallbackSave(preferences, ConfigDir, userCfgDir),
		Theme:           stringWithFallbackSave(preferences, Theme, ""),
	}

	return conf
}
