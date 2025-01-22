//go:build linux
// +build linux

package platform

import (
	"context"
	"log/slog"
	"log/syslog"
	"os"
	"os/user"
	"path/filepath"
)

type SyslogHandler struct {
	LogHandler
	writer *syslog.Writer
}

func NewSyslogHandler(writer *syslog.Writer) *SyslogHandler {
	return &SyslogHandler{writer: writer}
}

func (h *SyslogHandler) Handle(_ context.Context, record slog.Record) error {
	message := record.Message
	switch record.Level {
	case slog.LevelDebug:
		return h.writer.Debug(message)
	case slog.LevelInfo:
		return h.writer.Info(message)
	case slog.LevelWarn:
		return h.writer.Warning(message)
	case slog.LevelError:
		return h.writer.Err(message)
	default:
		return h.writer.Info(message)
	}
}

type PlatformLinux struct {
	username         string
	serviceNamespace string
	userHomeDir      string
}

func NewOSPlatform(serviceNamespace string) (*PlatformLinux, error) {
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

// Return slog.Logger for Linux
func (p PlatformLinux) Logger() *slog.Logger {
	writer, err := syslog.New(syslog.LOG_INFO|syslog.LOG_USER, p.serviceNamespace)
	if err != nil {
		panic(err)
	}
	defer writer.Close()

	handler := NewSyslogHandler(writer)
	logger := slog.New(handler)
	return logger
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
