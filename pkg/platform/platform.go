package platform

import (
	"os"
	"os/user"
	"runtime"
)

type Platform interface {
	// Get the username as known to the operating system
	GetUsername() string
	// Get the user's home directory
	GetUserHomeDir() string
	// Get the data directory for the platform
	GetDataDirectory() string
	// Get the config directory for the platform
	GetConfigDirectory() string
}

// NewPlatform creates a new platform object based on the current operating system
func NewPlatform(serviceNamespace, GOOS string) (Platform, error) {
	switch GOOS {
	case "linux":
		return NewPlatformLinux(serviceNamespace)
	case "windows":
		return NewPlatformWindows(serviceNamespace)
	case "darwin":
		return NewPlatformDarwin(serviceNamespace)
	default:
		return nil, ErrGettingUserOS
	}
}

// getCurrentUserOS gets the current username and home directory
func getCurrentUserOS() (string, string, error) {
	usrHomeDir, err := os.UserHomeDir()
	if err != nil {
		return "", "", ErrGettingUserOS
	}
	var usr *user.User
	// Check platform
	if runtime.GOOS == "windows" {
		// On Windows, use user.Current() if available, else fallback to environment variable
		usr, err = user.Current()
		if err != nil {
			// TODO: test this on windows
			usr = &user.User{Username: os.Getenv("USERNAME")}
			if usr.Username == "" {
				return "", "", ErrGettingUserOS
			}
		}
	} else {
		// On Unix-like systems (Linux, macOS), use user.Current()
		usr, err = user.Current()
		if err != nil {
			return "", "", ErrGettingUserOS
		}
	}
	return usr.Username, usrHomeDir, nil
}
