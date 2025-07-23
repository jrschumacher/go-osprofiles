package platform

import (
	"fmt"
	"os"
	"os/user"
	
	"github.com/jrschumacher/go-osprofiles/pkg/store"
)


type PlatformLinux struct {
	username         string
	serviceNamespace string
	servicePublisher string
	userHomeDir      string
}

func NewPlatformLinux(servicePublisher, serviceNamespace string) (*PlatformLinux, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, ErrGettingUserOS
	}

	usrHomeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, ErrGettingUserOS
	}

	return &PlatformLinux{usr.Username, serviceNamespace, servicePublisher, usrHomeDir}, nil
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
// ~/.local/share/<servicePublisher>/<serviceNamespace>
// ~/.local/share/<serviceNamespace> (if no publisher)
func (p PlatformLinux) UserAppDataDirectory() string {
	return buildLinuxUserDataPath(p.userHomeDir, p.servicePublisher, p.serviceNamespace)
}

// UserAppConfigDirectory returns the config directory for Linux.
// ~/.config/<servicePublisher>/<serviceNamespace>
// ~/.config/<serviceNamespace>
func (p PlatformLinux) UserAppConfigDirectory() string {
	return buildLinuxUserConfigPath(p.userHomeDir, p.servicePublisher, p.serviceNamespace)
}

// SystemAppDataDirectory returns the system-level data directory for Linux.
// /usr/local/<servicePublisher>/<serviceNamespace>
// /usr/local/<serviceNamespace> (if no publisher)
// Uses OSPROFILES_TEST_BASE_PATH environment variable as base if set (for testing)
func (p PlatformLinux) SystemAppDataDirectory() string {
	if p.servicePublisher != "" {
		return buildLinuxSystemPath(linuxUsrLocalPath, p.servicePublisher, p.serviceNamespace)
	}
	return buildLinuxSystemPath(linuxUsrLocalPath, p.serviceNamespace)
}

// SystemAppConfigDirectory returns the system-level config directory for Linux.
// /etc/<servicePublisher>/<serviceNamespace>
// /etc/<serviceNamespace> (if no publisher)
// Uses OSPROFILES_TEST_BASE_PATH environment variable as base if set (for testing)
func (p PlatformLinux) SystemAppConfigDirectory() string {
	if p.servicePublisher != "" {
		return buildLinuxSystemPath(linuxEtcPath, p.servicePublisher, p.serviceNamespace)
	}
	return buildLinuxSystemPath(linuxEtcPath, p.serviceNamespace)
}

// MDMConfigPath returns empty string as MDM is not supported on Linux.
func (p PlatformLinux) MDMConfigPath() string {
	return "" // MDM not supported on Linux
}

// MDMConfigExists returns false as MDM is not supported on Linux.
func (p PlatformLinux) MDMConfigExists() bool {
	return false // MDM not supported on Linux
}

// GetMDMData returns error as MDM is not supported on Linux
func (p PlatformLinux) GetMDMData() ([]byte, error) {
	return nil, fmt.Errorf("MDM is not supported on Linux")
}

// GetMDMDataAsJSON returns error as MDM is not supported on Linux
func (p PlatformLinux) GetMDMDataAsJSON(expandJSONStrings bool) ([]byte, error) {
	return nil, fmt.Errorf("MDM is not supported on Linux")
}

// SystemAppDataDirectoryWithMDM returns the system directory (no MDM support on Linux)
func (p PlatformLinux) SystemAppDataDirectoryWithMDM(reverseDNS ...string) (string, []store.DriverOpt) {
	systemDir := p.SystemAppDataDirectory()
	opts := []store.DriverOpt{
		store.WithStoreDirectory(systemDir),
		// No MDM support on Linux
	}
	return systemDir, opts
}
