//go:build windows
// +build windows

package platform

import (
	"context"
	"fmt"
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
	programFiles     string
	programData      string
	localAppData     string
}

const (
	EnvKeyLocalAppData = "LOCALAPPDATA"
	EnvKeyProgramData  = "PROGRAMDATA"
	EnvKeyProgramFiles = "PROGRAMFILES"
	EnvKeyUsername     = "USERNAME"
)

func NewOSPlatform(serviceNamespace string) (*PlatformWindows, error) {
	// On Windows, use user.Current() if available, else fallback to environment variable
	usr, err := user.Current()
	if err != nil {
		usr = &user.User{Username: os.Getenv(EnvKeyUsername)}
		if usr.Username == "" {
			return nil, ErrGettingUserOS
		}
	}
	usrHomeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, ErrGettingUserOS
	}

	programFiles := os.Getenv(EnvKeyProgramFiles)
	if programFiles == "" {
		return nil, fmt.Errorf("failed to detect %%%s%% in environment: %w", EnvKeyProgramFiles, ErrGettingUserOS)
	}

	programData := os.Getenv(EnvKeyProgramData)
	if programData == "" {
		return nil, fmt.Errorf("failed to detect %%%s%% in environment: %w", EnvKeyProgramData, ErrGettingUserOS)
	}

	localAppData := os.Getenv(EnvKeyLocalAppData)
	if localAppData == "" {
		return nil, fmt.Errorf("failed to detect %%%s%% in environment: %w", EnvKeyLocalAppData, ErrGettingUserOS)
	}

	return &PlatformWindows{
		username:         usr.Username,
		serviceNamespace: serviceNamespace,
		userHomeDir:      usrHomeDir,
		programFiles:     programFiles,
		programData:      programData,
		localAppData:     localAppData,
	}, nil
}

// GetUsername returns the username for Windows.
func (p PlatformWindows) GetUsername() string {
	return p.username
}

// UserHomeDir returns the user's home directory on Windows.
func (p PlatformWindows) UserHomeDir() string {
	return p.userHomeDir
}

// UserAppDataDirectory returns the namespaced user-level data directory for windows.
// i.e. %LocalAppData%\<serviceNamespace>
func (p PlatformWindows) UserAppDataDirectory() string {
	return filepath.Join(p.localAppData, p.serviceNamespace)
}

// UserAppConfigDirectory returns the namespaced user-level config directory for windows.
// i.e. %LocalAppData%\<serviceNamespace>
func (p PlatformWindows) UserAppConfigDirectory() string {
	return filepath.Join(p.localAppData, p.serviceNamespace)
}

// SystemAppDataDirectory returns the namespaced system-level data directory for windows.
// %ProgramData%\<serviceNamespace>
func (p PlatformWindows) SystemAppDataDirectory() string {
	return filepath.Join(p.programData, p.serviceNamespace)
}

// SystemAppConfigDirectory returns the namespaced system-level config directory for windows.
// %ProgramFiles%\<serviceNamespace>
func (p PlatformWindows) SystemAppConfigDirectory() string {
	return filepath.Join(p.programFiles, p.serviceNamespace)
}

// Return slog.Logger for Windows
func (p PlatformWindows) Logger() *slog.Logger {
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
