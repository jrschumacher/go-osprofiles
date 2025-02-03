package platform

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

type PlatformWindows struct {
	username         string
	serviceNamespace string
	servicePublisher string
	userHomeDir      string
	programFiles     string
	programData      string
	localAppData     string
}

const (
	envKeyLocalAppData = "LOCALAPPDATA"
	envKeyProgramData  = "PROGRAMDATA"
	envKeyProgramFiles = "PROGRAMFILES"
	envKeyUsername     = "USERNAME"
)

func NewPlatformWindows(servicePublisher, serviceNamespace string) (*PlatformWindows, error) {
	// On Windows, use user.Current() if available, else fallback to environment variable
	usr, err := user.Current()
	if err != nil {
		usr = &user.User{Username: os.Getenv(envKeyUsername)}
		if usr.Username == "" {
			return nil, ErrGettingUserOS
		}
	}
	usrHomeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, ErrGettingUserOS
	}

	programFiles := os.Getenv(envKeyProgramFiles)
	if programFiles == "" {
		return nil, fmt.Errorf("failed to detect %%%s%% in environment: %w", envKeyProgramFiles, ErrGettingUserOS)
	}

	programData := os.Getenv(envKeyProgramData)
	if programData == "" {
		return nil, fmt.Errorf("failed to detect %%%s%% in environment: %w", envKeyProgramData, ErrGettingUserOS)
	}

	localAppData := os.Getenv(envKeyLocalAppData)
	if localAppData == "" {
		return nil, fmt.Errorf("failed to detect %%%s%% in environment: %w", envKeyLocalAppData, ErrGettingUserOS)
	}

	return &PlatformWindows{
		username:         usr.Username,
		serviceNamespace: serviceNamespace,
		servicePublisher: servicePublisher,
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
