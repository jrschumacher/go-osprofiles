package platform

import (
	"fmt"
	"os"
	"path/filepath"
)

// Configurable variables for testing - can be overridden in tests
var (
	// Environment variable name for test base path override
	testBasePathEnvVar = "OSPROFILES_TEST_BASE_PATH"
	
	// System path constants that can be overridden in tests
	systemRootPath = "/"
	
	// Darwin-specific paths
	darwinLibraryPath       = "Library"
	darwinManagedPrefsPath  = "Managed Preferences"
	darwinAppSupportPath    = "Application Support"
	
	// Linux-specific paths  
	linuxUsrLocalPath = "usr/local"
	linuxEtcPath      = "etc"
	
	// Windows-specific paths
	windowsProgramDataPath  = "ProgramData"
	windowsProgramFilesPath = "ProgramFiles"
)

// getTestBasePath returns the test base path from environment variable if set
// This allows overriding system paths for testing without admin permissions
func getTestBasePath() string {
	return os.Getenv(testBasePathEnvVar)
}

// getSystemBasePath returns the appropriate base path for system directories
// If test base path is set, uses that as base, otherwise uses the provided default
func getSystemBasePath(defaultPath string) string {
	if testPath := getTestBasePath(); testPath != "" {
		return testPath
	}
	return defaultPath
}

// buildSystemPath constructs a system path using test base path override if available
// Equivalent to: getSystemBasePath(systemRootPath) + pathElements...
func buildSystemPath(pathElements ...string) string {
	basePath := getSystemBasePath(systemRootPath)
	allElements := append([]string{basePath}, pathElements...)
	return filepath.Join(allElements...)
}

// buildDarwinSystemPath constructs a macOS system path with Library prefix
// Uses test base path if set, otherwise /Library/...
func buildDarwinSystemPath(pathElements ...string) string {
	basePath := getSystemBasePath(systemRootPath)
	allElements := append([]string{basePath, darwinLibraryPath}, pathElements...)
	return filepath.Join(allElements...)
}

// buildLinuxSystemPath constructs a Linux system path
// Uses test base path if set, otherwise /...
func buildLinuxSystemPath(pathElements ...string) string {
	return buildSystemPath(pathElements...)
}

// buildWindowsSystemPath constructs a Windows system path
// Uses test base path if set, otherwise uses the provided base directory
func buildWindowsSystemPath(defaultBase string, subPath string, pathElements ...string) string {
	var basePath string
	if testPath := getTestBasePath(); testPath != "" {
		basePath = filepath.Join(testPath, subPath)
	} else {
		basePath = defaultBase
	}
	
	allElements := append([]string{basePath}, pathElements...)
	return filepath.Join(allElements...)
}

// buildMDMSystemPath constructs the MDM managed preferences path for macOS
// Uses test base path if set, otherwise default system path
func buildMDMSystemPath(identifier string) string {
	plistName := fmt.Sprintf("%s.plist", identifier)
	return buildDarwinSystemPath(darwinManagedPrefsPath, plistName)
}

// buildUserPath constructs a user directory path with optional publisher namespace
// If publisher is provided: basePath/publisher/namespace
// If no publisher: basePath/namespace
func buildUserPath(basePath, publisher, namespace string) string {
	if publisher != "" {
		return filepath.Join(basePath, publisher, namespace)
	}
	return filepath.Join(basePath, namespace)
}

// buildDarwinUserPath constructs a macOS user directory path
// userHomeDir/Library/Application Support/[publisher]/namespace
func buildDarwinUserPath(userHomeDir, publisher, namespace string) string {
	basePath := filepath.Join(userHomeDir, darwinLibraryPath, darwinAppSupportPath)
	return buildUserPath(basePath, publisher, namespace)
}

// buildLinuxUserDataPath constructs a Linux user data directory path  
// userHomeDir/.local/share/[publisher]/namespace
func buildLinuxUserDataPath(userHomeDir, publisher, namespace string) string {
	basePath := filepath.Join(userHomeDir, ".local", "share")
	return buildUserPath(basePath, publisher, namespace)
}

// buildLinuxUserConfigPath constructs a Linux user config directory path
// userHomeDir/.config/[publisher]/namespace  
func buildLinuxUserConfigPath(userHomeDir, publisher, namespace string) string {
	basePath := filepath.Join(userHomeDir, ".config")
	return buildUserPath(basePath, publisher, namespace)
}

// buildWindowsUserPath constructs a Windows user directory path
// localAppData/[publisher]/namespace
func buildWindowsUserPath(localAppData, publisher, namespace string) string {
	return buildUserPath(localAppData, publisher, namespace)
}