package platform

import (
	"os"
	"os/user"
	"path/filepath"
)

type PlatformDarwin struct {
	username         string
	serviceNamespace string
	userHomeDir      string
}

func NewPlatformDarwin(serviceNamespace string) (*PlatformDarwin, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, ErrGettingUserOS
	}

	usrHomeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, ErrGettingUserOS
	}

	return &PlatformDarwin{usr.Username, serviceNamespace, usrHomeDir}, nil
}

// GetUsername returns the username for macOS.
func (p PlatformDarwin) GetUsername() string {
	return p.username
}

// GetUserHomeDir returns the user's home directory on macOS.
func (p PlatformDarwin) GetUserHomeDir() string {
	return p.userHomeDir
}

// GetDataDirectory returns the data directory for macOS.
func (p PlatformDarwin) GetDataDirectory() string {
	return filepath.Join(p.userHomeDir, "Library", "Application Support", p.serviceNamespace)
}

// GetConfigDirectory returns the config directory for macOS.
func (p PlatformDarwin) GetConfigDirectory() string {
	return filepath.Join(p.userHomeDir, "Library", "Preferences", p.serviceNamespace)
}
