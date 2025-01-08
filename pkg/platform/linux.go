package platform

import "path/filepath"

type PlatformLinux struct {
	username         string
	serviceNamespace string
	userHomeDir      string
}

func NewPlatformLinux(username, serviceNamespace, userHomeDir string) PlatformLinux {
	return PlatformLinux{username, serviceNamespace, userHomeDir}
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
	return filepath.Join(p.userHomeDir, ".local", "share")
}

// GetConfigDirectory returns the config directory for Linux.
func (p PlatformLinux) GetConfigDirectory() string {
	return filepath.Join(p.userHomeDir, ".config")
}
