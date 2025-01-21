//go:build darwin
// +build darwin

package platform

/*
#cgo LDFLAGS: -framework Foundation -framework CoreFoundation -framework os

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <os/log.h>
#include "os_log_wrapper.c"
#include <os/log.h>

void log_to_unified(const char *message) {
	os_log(OS_LOG_DEFAULT, "%{public}s", message);
}
*/

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
	cMessage := C.CString(message)
	println(cMessage)
	// defer C.free(unsafe.Pointer(cMessage))
	// could not determine kind of name for C.log_to_unified
	C.log_to_unified(cMessage)
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

// GetUserHomeDir returns the user's home directory on macOS.
func (p PlatformDarwin) GetUserHomeDir() string {
	return p.userHomeDir
}

// GetDataDirectory returns the data directory for macOS.
func (p PlatformDarwin) GetDataDirectory() string {
	return filepath.Join(p.userHomeDir, "Library", "Application Support", p.serviceNamespace)
}

// GetConfigDirectory returns the config directory for macOS.
func (p PlatformDarwin) GetConfigDirectory() string {
	return filepath.Join(p.userHomeDir, "Library", "Preferences", p.serviceNamespace)
}

// Return slog.Logger for macOS
func (p PlatformDarwin) GetLogger() *slog.Logger {
	handler := NewUnifiedLoggingHandler()
	logger := slog.New(handler)
	return logger
}
