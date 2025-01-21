//go:build windows
// +build windows

package platform

import (
	"log/slog"
	"os"
	"os/user"
	"path/filepath"

	"golang.org/x/sys/windows/svc/eventlog"
)

type PlatformWindows struct {
	username         string
	serviceNamespace string
	userHomeDir      string
}

func NewOSPlatform(serviceNamespace string) (*PlatformWindows, error) {
	// On Windows, use user.Current() if available, else fallback to environment variable
	usr, err := user.Current()
	if err != nil {
		// TODO: test this on windows
		usr = &user.User{Username: os.Getenv("USERNAME")}
		if usr.Username == "" {
			return nil, ErrGettingUserOS
		}
	}
	usrHomeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, ErrGettingUserOS
	}
	return &PlatformWindows{usr.Username, serviceNamespace, usrHomeDir}, nil
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

// TODO: it looks like this is different depending on OS version, so we should consider that
// https://learn.microsoft.com/en-us/windows/apps/design/app-settings/store-and-retrieve-app-data

// GetDataDirectory returns the data directory for Windows.
func (p PlatformWindows) GetDataDirectory() string {
	return filepath.Join(p.userHomeDir, "AppData", "Roaming", p.serviceNamespace)
}

// GetConfigDirectory returns the config directory for Windows.
func (p PlatformWindows) GetConfigDirectory() string {
	return filepath.Join(p.userHomeDir, "AppData", "Local", p.serviceNamespace)
}

// Return slog.Logger for Windows
func (p PlatformWindows) GetLogger() *slog.Logger {
	writer, err := eventlog.Open(p.serviceNamespace)
	if err != nil {
		panic(err)
	}
	defer writer.Close()

	handler := NewSyslogHandler(writer, slog.LevelInfo)
	logger := slog.New(handler)
	return logger
}
