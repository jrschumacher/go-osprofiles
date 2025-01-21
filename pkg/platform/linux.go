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
	writer, err := syslog.New(syslog.LOG_INFO|syslog.LOG_USER, p.serviceNamespace)
	if err != nil {
		panic(err)
	}
	defer writer.Close()

	handler := NewSyslogHandler(writer)
	logger := slog.New(handler)
	return logger
}
