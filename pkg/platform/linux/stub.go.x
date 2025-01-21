//go:build windows

package linux

import (
	"log/slog"
)

type PlatformLinux struct {
	username         string
	serviceNamespace string
	userHomeDir      string
}

func NewPlatformLinux(serviceNamespace string) (*PlatformLinux, error) {
	return &PlatformLinux{"", "", ""}, nil
}

// GetUsername returns the username for Linux.
func (p PlatformLinux) GetUsername() string {
	return ""
}

// GetUserHomeDir returns the user's home directory on Linux.
func (p PlatformLinux) GetUserHomeDir() string {
	return ""
}

// GetDataDirectory returns the data directory for Linux.
func (p PlatformLinux) GetDataDirectory() string {
	return ""
}

// GetConfigDirectory returns the config directory for Linux.
func (p PlatformLinux) GetConfigDirectory() string {
	return ""
}

// Return slog.Logger for Linux
func (p PlatformLinux) GetLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(nil, nil))
}
