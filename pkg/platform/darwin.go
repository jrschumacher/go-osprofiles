//go:build darwin
// +build darwin

package platform

import (
	"C"
	"context"
	"log/slog"
	"os"
	"os/user"
	"path/filepath"
)

type UnifiedLoggingHandler struct {
	LogHandler
}

func NewUnifiedLoggingHandler() *UnifiedLoggingHandler {
	return &UnifiedLoggingHandler{}
}

func (h *UnifiedLoggingHandler) Handle(_ context.Context, record slog.Record) error {
	message := record.Message
	LogMessage(message)
	return nil
}

type PlatformDarwin struct {
	username         string
	serviceNamespace string
	userHomeDir      string
}

func NewOSPlatform(serviceNamespace string) (*PlatformDarwin, error) {
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

// UserHomeDir returns the user's home directory on macOS.
func (p PlatformDarwin) UserHomeDir() string {
	return p.userHomeDir
}

// UserAppDataDirectory returns the namespaced user-level data directory for macOS.
// i.e. ~/Library/Application Support/<serviceNamespace>
func (p PlatformDarwin) UserAppDataDirectory() string {
	return filepath.Join(p.userHomeDir, "Library", "Application Support", p.serviceNamespace)
}

// UserAppConfigDirectory returns the namespaced user-level config directory for macOS.
// i.e. ~/Library/Application Support/<serviceNamespace>
func (p PlatformDarwin) UserAppConfigDirectory() string {
	return filepath.Join(p.userHomeDir, "Library", "Application Support", p.serviceNamespace)
}

// SystemAppDataDirectory returns the namespaced system-level data directory for macOS.
// i.e. /Library/Application Support/<serviceNamespace>
func (p PlatformDarwin) SystemAppDataDirectory() string {
	return filepath.Join("/", "Library", "Application Support", p.serviceNamespace)
}

// SystemAppConfigDirectory returns the namespaced system-level config directory for macOS.
// i.e. /Library/Application Support/<serviceNamespace>
func (p PlatformDarwin) SystemAppConfigDirectory() string {
	return filepath.Join("/", "Library", "Application Support", p.serviceNamespace)
}

// Return slog.Logger for macOS
func (p PlatformDarwin) Logger() *slog.Logger {
	handler := NewUnifiedLoggingHandler()
	logger := slog.New(handler)
	return logger
}
