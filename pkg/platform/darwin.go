package platform

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	
	"github.com/jrschumacher/go-osprofiles/pkg/store"
)

type PlatformDarwin struct {
	username         string
	serviceNamespace string
	servicePublisher string
	userHomeDir      string
}

const (
	darwinLibrary            = "Library"
	darwinAppSupport         = "Application Support"
	darwinManagedPreferences = "Managed Preferences"
)

func NewPlatformDarwin(servicePublisher, serviceNamespace string) (*PlatformDarwin, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, ErrGettingUserOS
	}

	usrHomeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, ErrGettingUserOS
	}

	return &PlatformDarwin{usr.Username, serviceNamespace, servicePublisher, usrHomeDir}, nil
}

// GetUsername returns the username for macOS.
func (p PlatformDarwin) GetUsername() string {
	return p.username
}

// UserHomeDir returns the user's home directory on macOS.
func (p PlatformDarwin) UserHomeDir() string {
	return p.userHomeDir
}

// UserAppDataDirectory returns the namespaced user-level data directory for macOS.
// ~/Library/Application Support/<servicePublisher>/<serviceNamespace>
// ~/Library/Application Support/<serviceNamespace> (if no pubisher)
func (p PlatformDarwin) UserAppDataDirectory() string {
	path := filepath.Join(p.userHomeDir, darwinLibrary, darwinAppSupport)
	if p.servicePublisher != "" {
		path = filepath.Join(path, p.servicePublisher)
	}
	return filepath.Join(path, p.serviceNamespace)
}

// UserAppConfigDirectory returns the namespaced user-level config directory for macOS.
// ~/Library/Application Support/<servicePublisher>/<serviceNamespace>
// ~/Library/Application Support/<serviceNamespace> (if no publisher)
func (p PlatformDarwin) UserAppConfigDirectory() string {
	path := filepath.Join(p.userHomeDir, darwinLibrary, darwinAppSupport)
	if p.servicePublisher != "" {
		path = filepath.Join(path, p.servicePublisher)
	}
	return filepath.Join(path, p.serviceNamespace)
}

// SystemAppDataDirectory returns the namespaced system-level data directory for macOS.
// If serviceNamespace is reverse DNS format, automatically enables MDM checking.
// /Library/Application Support/<servicePublisher>/<serviceNamespace>
// /Library/Application Support/<serviceNamespace> (if no publisher)
func (p PlatformDarwin) SystemAppDataDirectory() string {
	path := filepath.Join("/", darwinLibrary, darwinAppSupport)
	if p.servicePublisher != "" {
		path = filepath.Join(path, p.servicePublisher)
	}
	return filepath.Join(path, p.serviceNamespace)
}

// SystemAppConfigDirectory returns the namespaced system-level config directory for macOS.
// /Library/Application Support/<servicePublisher>/<serviceNamespace>
// /Library/Application Support/<serviceNamespace> (if no publisher)
func (p PlatformDarwin) SystemAppConfigDirectory() string {
	path := filepath.Join("/", darwinLibrary, darwinAppSupport)
	if p.servicePublisher != "" {
		path = filepath.Join(path, p.servicePublisher)
	}
	return filepath.Join(path, p.serviceNamespace)
}

// MDMConfigPath returns the path to the MDM managed preferences plist file.
// It uses RDNS format from publisher+namespace, or just namespace if it contains dots.
func (p PlatformDarwin) MDMConfigPath() string {
	var identifier string
	
	// Determine RDNS identifier for MDM
	if strings.Contains(p.serviceNamespace, ".") {
		// serviceNamespace is already RDNS format (e.g., "com.company.app")
		identifier = p.serviceNamespace
	} else if p.servicePublisher != "" && strings.Contains(p.servicePublisher, ".") {
		// Publisher is RDNS, append namespace (e.g., "com.company" + "myapp")
		identifier = fmt.Sprintf("%s.%s", p.servicePublisher, p.serviceNamespace)
	} else if p.servicePublisher != "" {
		// Neither is RDNS, combine them (e.g., "company" + "myapp")
		identifier = fmt.Sprintf("%s.%s", p.servicePublisher, p.serviceNamespace)
	} else {
		// Just use namespace as-is
		identifier = p.serviceNamespace
	}
	
	plistName := fmt.Sprintf("%s.plist", identifier)
	return filepath.Join("/", darwinLibrary, darwinManagedPreferences, plistName)
}

// MDMConfigExists checks if the MDM managed preferences file exists and is accessible.
func (p PlatformDarwin) MDMConfigExists() bool {
	path := p.MDMConfigPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	// Check if file is readable
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	file.Close()
	return true
}

// SystemAppDataDirectoryWithMDM returns the system directory with MDM support enabled
// If no explicit reverseDNS is provided and serviceNamespace contains dots, uses serviceNamespace as MDM identifier
func (p PlatformDarwin) SystemAppDataDirectoryWithMDM(reverseDNS ...string) (string, []store.DriverOpt) {
	systemDir := p.SystemAppDataDirectory()
	
	// Determine MDM identifier
	var mdmID string
	if len(reverseDNS) > 0 && reverseDNS[0] != "" {
		// Explicit MDM identifier provided
		mdmID = reverseDNS[0]
	} else if strings.Contains(p.serviceNamespace, ".") {
		// serviceNamespace is already reverse DNS format
		mdmID = p.serviceNamespace
	} else if p.servicePublisher != "" && strings.Contains(p.servicePublisher, ".") {
		// Publisher is reverse DNS, combine with namespace
		mdmID = fmt.Sprintf("%s.%s", p.servicePublisher, p.serviceNamespace)
	} else if p.servicePublisher != "" {
		// Neither is reverse DNS, but combine them anyway
		mdmID = fmt.Sprintf("%s.%s", p.servicePublisher, p.serviceNamespace)
	} else {
		// Just use namespace as-is
		mdmID = p.serviceNamespace
	}
	
	opts := []store.DriverOpt{
		store.WithStoreDirectory(systemDir),
	}
	
	// Only enable MDM if we have a valid identifier
	if mdmID != "" {
		opts = append(opts, store.WithMDMSupport(mdmID))
	}
	
	return systemDir, opts
}

// SystemAppDataDirectoryWithAutoMDM returns system directory with automatic MDM detection
// If serviceNamespace is reverse DNS format, automatically enables MDM checking
func (p PlatformDarwin) SystemAppDataDirectoryWithAutoMDM() (string, []store.DriverOpt) {
	if strings.Contains(p.serviceNamespace, ".") {
		// Namespace is reverse DNS - enable MDM automatically
		return p.SystemAppDataDirectoryWithMDM()
	} else {
		// Regular namespace - no MDM
		systemDir := p.SystemAppDataDirectory()
		return systemDir, []store.DriverOpt{store.WithStoreDirectory(systemDir)}
	}
}
