package platform

type Platform interface {
	// Get the username as known to the operating system
	GetUsername() string
	// Get the user's home directory
	GetUserHomeDir() string
	// Get the data directory for the platform
	GetDataDirectory() string
	// Get the config directory for the platform
	GetConfigDirectory() string
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
