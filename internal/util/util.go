package util

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"github.com/redawl/gitm/internal"
)

// GetConfigDir returns the path to the user config dir on the current machine as a string.
//
// If the config dir does not exist, it is created.
//
// If the default location for config directories does not exist, or if there is an error creating
// the config dir, an error is returned.
func GetConfigDir() (string, error) {
	// TODO: Remove dependency on fyne for this function
	a := fyne.CurrentApp()
	var cfgDir string
	if a != nil {
		cfgDir = a.Preferences().String(internal.ConfigDir)
	} else {
		// TODO: Use config here instead of hardcoding os config dir
		userCfgDir, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}

		cfgDir = filepath.Join(userCfgDir, "gitm")
	}

	if _, err := os.Stat(cfgDir); errors.Is(err, os.ErrNotExist) {
		slog.Debug("Config dir doesn't exist, creating")
		if err := os.Mkdir(cfgDir, 0o700); err != nil {
			return "", err
		}
	}

	return cfgDir, nil
}

// ReadCount reads at most length bytes from reader.
// If less than length bytes are read from reader, the bytes are returned along with an err
func ReadCount(reader io.Reader, length int) ([]byte, error) {
	buff := make([]byte, length)

	count, err := io.ReadAtLeast(reader, buff, length)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	} else if count == 0 {
		return nil, fmt.Errorf("reader closed before reading any bytes")
	}

	return buff, nil
}
