package platform

import "path/filepath"

type PlatformWindows struct {
	username         string
	serviceNamespace string
	userHomeDir      string
}

func NewPlatformWindows(username, serviceNamespace, userHomeDir string) PlatformWindows {
	return PlatformWindows{username, serviceNamespace, userHomeDir}
}

// TODO: validate these are correct

// GetUsername returns the username for Windows.
func (p PlatformWindows) GetUsername() string {
	return p.username
}

// GetUserHomeDir returns the user's home directory on Windows.
func (p PlatformWindows) GetUserHomeDir() string {
	return p.userHomeDir
}

// GetDataDirectory returns the data directory for Windows.
func (p PlatformWindows) GetDataDirectory() string {
	return filepath.Join(p.userHomeDir, "AppData", "Roaming")
}

// GetConfigDirectory returns the config directory for Windows.
func (p PlatformWindows) GetConfigDirectory() string {
	return filepath.Join(p.userHomeDir, "AppData", "Local")
}
