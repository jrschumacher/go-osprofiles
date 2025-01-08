package platform

import "path/filepath"

type PlatformDarwin struct {
	username         string
	serviceNamespace string
	userHomeDir      string
}

func NewPlatformDarwin(username, serviceNamespace, userHomeDir string) Platform {
	return PlatformDarwin{username, serviceNamespace, userHomeDir}
}

// TODO: validate these are correct

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
	return filepath.Join(p.userHomeDir, "Library", "Application Support")
}

// GetConfigDirectory returns the config directory for macOS.
func (p PlatformDarwin) GetConfigDirectory() string {
	return filepath.Join(p.userHomeDir, "Library", "Preferences")
}
