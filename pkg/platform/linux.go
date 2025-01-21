package platform

import (
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

// GetUsername returns the username for the Linux OS.
func (p PlatformLinux) GetUsername() string {
	return p.username
}

// UserHomeDir returns the user's home directory on the Linux OS.
func (p PlatformLinux) UserHomeDir() string {
	return p.userHomeDir
}

// UserAppDataDirectory returns the data directory for Linux.
// i.e. ~/.local/share/<serviceNamespace>
func (p PlatformLinux) UserAppDataDirectory() string {
	return filepath.Join(p.userHomeDir, ".local", "share", p.serviceNamespace)
}

// UserAppConfigDirectory returns the config directory for Linux.
// i.e. ~/.config/<serviceNamespace>
func (p PlatformLinux) UserAppConfigDirectory() string {
	return filepath.Join(p.userHomeDir, ".config", p.serviceNamespace)
}

// SystemAppDataDirectory returns the system-level data directory for Linux.
// i.e. /var/lib/<serviceNamespace>
func (p PlatformLinux) SystemAppDataDirectory() string {
	return filepath.Join("/", "var", "lib", p.serviceNamespace)
}

// SystemAppConfigDirectory returns the system-level config directory for Linux.
// i.e. /etc/<serviceNamespace>
func (p PlatformLinux) SystemAppConfigDirectory() string {
	return filepath.Join("/", "etc", p.serviceNamespace)
}
