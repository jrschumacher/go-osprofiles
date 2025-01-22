package platform

import (
	"context"
	"log/slog"
)

type LogHandler struct {
}

func (h *LogHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *LogHandler) Handle(_ context.Context, record slog.Record) error {
	return nil
}

func (h *LogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *LogHandler) WithGroup(name string) slog.Handler {
	return h
}

type Platform interface {
	// Get the username as known to the operating system
	GetUsername() string
	// Get the user's home directory
	UserHomeDir() string
	// Get the namespaced user-level data directory for the platform
	UserAppDataDirectory() string
	// Get the namespaced user-level config directory for the platform
	UserAppConfigDirectory() string
	// Get the namespaced system-level data directory for the platform
	SystemAppDataDirectory() string
	// Get the namespaced system-level config directory for the platform
	SystemAppConfigDirectory() string
	// Get the logger for the platform
	Logger() *slog.Logger
}

func NewPlatform(serviceNamespace string) (Platform, error) {
	return NewOSPlatform(serviceNamespace)
}
