package platform

import (
	"log/slog"
	"os"
	"os/user"
	"path/filepath"
)

type PlatformLinux struct {
	username         string
	serviceNamespace string
	userHomeDir      string
}

func NewPlatformLinux(serviceNamespace string) (*PlatformLinux, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, ErrGettingUserOS
	}

	usrHomeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, ErrGettingUserOS
	}

	return &PlatformLinux{usr.Username, serviceNamespace, usrHomeDir}, nil
}

// TODO: validate these are correct

// GetUsername returns the username for the Linux OS.
func (p PlatformLinux) GetUsername() string {
	return p.username
}

// GetUserHomeDir returns the user's home directory on the Linux OS.
func (p PlatformLinux) GetUserHomeDir() string {
	return p.userHomeDir
}

// GetDataDirectory returns the data directory for Linux.
func (p PlatformLinux) GetDataDirectory() string {
	return filepath.Join(p.userHomeDir, ".local", "share", p.serviceNamespace)
}

// GetConfigDirectory returns the config directory for Linux.
func (p PlatformLinux) GetConfigDirectory() string {
	return filepath.Join(p.userHomeDir, ".config", p.serviceNamespace)
}

// Return slog.Logger for Linux
func (p PlatformLinux) GetLogger() *slog.Logger {
	// TODO: Implement logger
	return &slog.Logger{}
}
