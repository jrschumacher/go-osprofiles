//go:build darwin
// +build darwin

package platform

/*
#include <os/log.h>
#include <stdlib.h>

void logInfo(const char *message) {
    os_log(OS_LOG_DEFAULT, "%{public}s", message);
}

void logError(const char *message) {
    os_log(OS_LOG_DEFAULT, "Error: %{public}s", message);
}
*/

import (
	"C"
	"context"
	"log/slog"
	"os"
	"os/user"
	"path/filepath"

	"github.com/jrschumacher/go-osprofiles/pkg"
)

// // logMessage logs an informational message to Unified Logging.
// func logMessage(message string) {
// 	cMessage := C.CString(message)
// 	defer C.free(unsafe.Pointer(cMessage)) // Free C string after use
// 	C.logInfo(cMessage)                    // Call C function to log info
// }

// // logError logs an error message to Unified Logging.
// func logError(message string) {
// 	cMessage := C.CString(message)
// 	defer C.free(unsafe.Pointer(cMessage)) // Free C string after use
// 	C.logError(cMessage)                   // Call C function to log error
// }

type UnifiedLoggingHandler struct {
	LogHandler
}

func NewUnifiedLoggingHandler() *UnifiedLoggingHandler {
	return &UnifiedLoggingHandler{}
}

func (h *UnifiedLoggingHandler) Handle(_ context.Context, record slog.Record) error {
	message := record.Message
	pkg.LogMessage(message)
	// cMessage := C.CString(message)
	// println(cMessage)
	// // defer C.free(unsafe.Pointer(cMessage))
	// // could not determine kind of name for C.log_to_unified
	// C.log_to_unified(cMessage)
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
