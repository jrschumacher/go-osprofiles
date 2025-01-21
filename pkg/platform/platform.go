package platform

import (
	"context"
	"log/slog"
)

type LogHandler struct {
	// writer interface{}
	// level  slog.Level
}

func NewLogHandler(writer interface{}, level slog.Level) *LogHandler {
	return &LogHandler{}
}

func (h *LogHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *LogHandler) Handle(_ context.Context, record slog.Record) error {
	switch record.Level {
	case slog.LevelDebug:
		return nil
	case slog.LevelInfo:
		return nil
	case slog.LevelWarn:
		return nil
	case slog.LevelError:
		return nil
	default:
		return nil
	}
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
	GetUserHomeDir() string
	// Get the data directory for the platform
	GetDataDirectory() string
	// Get the config directory for the platform
	GetConfigDirectory() string
	// Get the logger for the platform
	GetLogger() *slog.Logger
}

func NewPlatform(serviceNamespace string) (Platform, error) {
	return NewOSPlatform(serviceNamespace)
}
