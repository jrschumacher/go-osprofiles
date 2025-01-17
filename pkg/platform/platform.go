package platform

import (
	"context"
	"log/slog"
)

type LogHandler struct {
	writer interface{}
	level  slog.Level
}

func NewLogHandler(writer interface{}, level slog.Level) *LogHandler {
	return &LogHandler{writer, level}
}

func (h *LogHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

// func (h *SyslogHandler) Handle(_ context.Context, record slog.Record) error {
// 	message := record.Message
// 	writer, ok := h.writer.(*syslog.Writer)
// 	switch record.Level {
// 	case slog.LevelDebug:
// 		return writer.Debug(message)
// 	case slog.LevelInfo:
// 		return writer.Info(message)
// 	case slog.LevelWarn:
// 		return writer.Warning(message)
// 	case slog.LevelError:
// 		return writer.Err(message)
// 	default:
// 		return writer.Info(message)
// 	}
// }

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
