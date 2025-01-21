//go:build windows
// +build windows

package platform

import (
	"context"
	"log/slog"
	"math/rand"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/svc/eventlog"
)

type EventLogHandler struct {
	LogHandler
	writer *eventlog.Log
}

func NewEventLogHandler(writer *eventlog.Log) *EventLogHandler {
	return &EventLogHandler{writer: writer}
}

func (h *EventLogHandler) Handle(_ context.Context, record slog.Record) error {
	message := record.Message
	randNum := rand.Intn(1000) + 1
	eid := uint32(randNum)
	switch record.Level {
	case slog.LevelDebug:
		return h.writer.Info(eid, message)
	case slog.LevelInfo:
		return h.writer.Info(eid, message)
	case slog.LevelWarn:
		return h.writer.Warning(eid, message)
	case slog.LevelError:
		return h.writer.Error(eid, message)
	default:
		return h.writer.Info(eid, message)
	}
}

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
	// Check if the event source exists and create it if it doesn't
	if err := eventlog.InstallAsEventCreate(p.serviceNamespace, eventlog.Error|eventlog.Warning|eventlog.Info); err != nil {
		if !strings.Contains(err.Error(), "registry key already exists") {
			panic(err)
		}
	}

	// Open the event log
	writer, err := eventlog.Open(p.serviceNamespace)
	if err != nil {
		panic(err)
	}
	defer writer.Close()

	handler := NewEventLogHandler(writer)
	logger := slog.New(handler)
	return logger
}
