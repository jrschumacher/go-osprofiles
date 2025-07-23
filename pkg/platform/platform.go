package platform

import "github.com/jrschumacher/go-osprofiles/pkg/store"

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
	// Get the MDM managed preferences path (system-level, highest precedence)
	MDMConfigPath() string
	// Check if MDM config exists and is accessible
	MDMConfigExists() bool
	// Read raw MDM data from managed preferences plist
	GetMDMData() ([]byte, error)
	// Read MDM data with intelligent JSON string handling
	GetMDMDataAsJSON(expandJSONStrings bool) ([]byte, error)
	// Get system directory with MDM support enabled (for file store integration)
	SystemAppDataDirectoryWithMDM(reverseDNS ...string) (string, []store.DriverOpt)
}

// NewPlatform creates a new platform object based on the current operating system
func NewPlatform(servicePublisher, serviceNamespace, GOOS string) (Platform, error) {
	switch GOOS {
	case "linux":
		return NewPlatformLinux(servicePublisher, serviceNamespace)
	case "windows":
		return NewPlatformWindows(servicePublisher, serviceNamespace)
	case "darwin":
		return NewPlatformDarwin(servicePublisher, serviceNamespace)
	default:
		return nil, ErrGettingUserOS
	}
}
