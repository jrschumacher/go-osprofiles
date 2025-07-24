package platform

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"strings"
	
	"github.com/micromdm/plist"
	"github.com/jrschumacher/go-osprofiles/pkg/store"
)


type PlatformDarwin struct {
	username         string
	serviceNamespace string
	servicePublisher string
	userHomeDir      string
}


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
	return buildDarwinUserPath(p.userHomeDir, p.servicePublisher, p.serviceNamespace)
}

// UserAppConfigDirectory returns the namespaced user-level config directory for macOS.
// ~/Library/Application Support/<servicePublisher>/<serviceNamespace>
// ~/Library/Application Support/<serviceNamespace> (if no publisher)
func (p PlatformDarwin) UserAppConfigDirectory() string {
	return buildDarwinUserPath(p.userHomeDir, p.servicePublisher, p.serviceNamespace)
}

// SystemAppDataDirectory returns the namespaced system-level data directory for macOS.
// If serviceNamespace is reverse DNS format, automatically enables MDM checking.
// /Library/Application Support/<servicePublisher>/<serviceNamespace>
// /Library/Application Support/<serviceNamespace> (if no publisher)
// Uses OSPROFILES_TEST_BASE_PATH environment variable as base if set (for testing)
func (p PlatformDarwin) SystemAppDataDirectory() string {
	if p.servicePublisher != "" {
		return buildDarwinSystemPath(darwinAppSupportPath, p.servicePublisher, p.serviceNamespace)
	}
	return buildDarwinSystemPath(darwinAppSupportPath, p.serviceNamespace)
}

// SystemAppConfigDirectory returns the namespaced system-level config directory for macOS.
// /Library/Application Support/<servicePublisher>/<serviceNamespace>
// /Library/Application Support/<serviceNamespace> (if no publisher)
// Uses OSPROFILES_TEST_BASE_PATH environment variable as base if set (for testing)
func (p PlatformDarwin) SystemAppConfigDirectory() string {
	if p.servicePublisher != "" {
		return buildDarwinSystemPath(darwinAppSupportPath, p.servicePublisher, p.serviceNamespace)
	}
	return buildDarwinSystemPath(darwinAppSupportPath, p.serviceNamespace)
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
	
	return buildMDMSystemPath(identifier)
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
		opts = append(opts, store.WithMDMSupport(mdmID, p)) // Pass the platform instance for MDM operations
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

// GetMDMData reads data from MDM managed preferences plist
func (p PlatformDarwin) GetMDMData() ([]byte, error) {
	mdmPath := p.MDMConfigPath()
	
	// Read the plist file
	file, err := os.Open(mdmPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open MDM plist: %w", err)
	}
	defer file.Close()
	
	// Parse the plist into a generic interface
	var plistData any
	decoder := plist.NewDecoder(file)
	if err := decoder.Decode(&plistData); err != nil {
		return nil, fmt.Errorf("failed to parse MDM plist: %w", err)
	}
	
	// Convert the parsed data to JSON for consistent handling
	jsonData, err := json.Marshal(plistData)
	if err != nil {
		return nil, fmt.Errorf("failed to convert plist data to JSON: %w", err)
	}
	
	return jsonData, nil
}

// GetMDMDataAsJSON reads MDM plist data and attempts to intelligently handle JSON strings
// If the plist root is a JSON string, it parses and returns the JSON directly
// If the plist contains JSON string values, it optionally expands them
func (p PlatformDarwin) GetMDMDataAsJSON(expandJSONStrings bool) ([]byte, error) {
	rawData, err := p.GetMDMData()
	if err != nil {
		return nil, err
	}

	// Try to detect if the root value is a JSON string
	var rootValue any
	if err := json.Unmarshal(rawData, &rootValue); err != nil {
		return rawData, nil // Return as-is if we can't parse
	}

	// Case 1: Root value is a string that might be JSON
	if rootStr, ok := rootValue.(string); ok {
		// Try to parse as JSON
		var jsonObj any
		if err := json.Unmarshal([]byte(rootStr), &jsonObj); err == nil {
			// It's valid JSON! Return the parsed JSON instead of the string wrapper
			return json.Marshal(jsonObj)
		}
		// Not valid JSON, return the original data
		return rawData, nil
	}

	// Case 2: Root value is a map/dict that might contain JSON strings
	if expandJSONStrings {
		return p.expandJSONStrings(rawData)
	}

	return rawData, nil
}

const (
	// maxJSONExpansionDepth limits recursive JSON expansion to prevent stack overflow
	maxJSONExpansionDepth = 10
)

// expandJSONStrings recursively finds string values that are valid JSON and expands them
func (p PlatformDarwin) expandJSONStrings(data []byte) ([]byte, error) {
	var obj any
	if err := json.Unmarshal(data, &obj); err != nil {
		return data, nil // Return as-is if we can't parse
	}

	expanded := p.expandJSONStringsRecursive(obj, 0)
	return json.Marshal(expanded)
}

// expandJSONStringsRecursive recursively processes data structures to expand JSON strings
// with depth limiting to prevent stack overflow
func (p PlatformDarwin) expandJSONStringsRecursive(obj any, depth int) any {
	// Prevent infinite recursion by limiting depth
	if depth >= maxJSONExpansionDepth {
		return obj
	}

	switch v := obj.(type) {
	case map[string]any:
		result := make(map[string]any)
		for key, value := range v {
			result[key] = p.expandJSONStringsRecursive(value, depth+1)
		}
		return result
		
	case []any:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = p.expandJSONStringsRecursive(item, depth+1)
		}
		return result
		
	case string:
		// Try to parse as JSON
		var jsonObj any
		if err := json.Unmarshal([]byte(v), &jsonObj); err == nil {
			// It's valid JSON! Return the parsed object instead of the string
			return p.expandJSONStringsRecursive(jsonObj, depth+1)
		}
		// Not valid JSON, return the string as-is
		return v
		
	default:
		return v
	}
}
