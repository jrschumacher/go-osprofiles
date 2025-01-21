//go:build !windows

package windows

import (
	"log/slog"
)

type PlatformWindows struct {
	username         string
	serviceNamespace string
	userHomeDir      string
}

func NewPlatformWindows(serviceNamespace string) (*PlatformWindows, error) {
	return &PlatformWindows{"", "", ""}, nil
}

// TODO: validate these are correct

// GetUsername returns the username for Windows.
func (p PlatformWindows) GetUsername() string {
	return ""
}

// GetUserHomeDir returns the user's home directory on Windows.
func (p PlatformWindows) GetUserHomeDir() string {
	return ""
}

// TODO: it looks like this is different depending on OS version, so we should consider that
// https://learn.microsoft.com/en-us/windows/apps/design/app-settings/store-and-retrieve-app-data

// GetDataDirectory returns the data directory for Windows.
func (p PlatformWindows) GetDataDirectory() string {
	return ""
}

// GetConfigDirectory returns the config directory for Windows.
func (p PlatformWindows) GetConfigDirectory() string {
	return ""
}

// Return slog.Logger for Windows
func (p PlatformWindows) GetLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(nil, nil))
}
