package platform

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_PlatformMacOS(t *testing.T) {
	fakeAppName := "test-macos-app"
	plat, err := NewPlatform(fakeAppName, "darwin")

	require.NoError(t, err)
	require.NotNil(t, plat)

	darwin, ok := plat.(*PlatformDarwin)
	require.True(t, ok)
	require.NotNil(t, darwin)

	// user scoped
	configDir := darwin.UserAppConfigDirectory()
	assert.True(t, strings.HasSuffix(configDir, fakeAppName))
	assert.True(t, strings.Contains(configDir, "/Library/Application Support"))
	assert.False(t, strings.HasPrefix(configDir, "/Library/Application Support"))

	dataDir := darwin.UserAppDataDirectory()
	assert.True(t, strings.HasSuffix(dataDir, fakeAppName))
	assert.True(t, strings.Contains(dataDir, "/Library/Application Support"))
	assert.False(t, strings.HasPrefix(configDir, "/Library/Application Support"))

	// system scoped
	configDir = darwin.SystemAppConfigDirectory()
	assert.True(t, strings.HasSuffix(configDir, fakeAppName))
	assert.True(t, strings.HasPrefix(configDir, "/Library/Application Support"))

	dataDir = darwin.SystemAppDataDirectory()
	assert.True(t, strings.HasSuffix(dataDir, fakeAppName))
	assert.True(t, strings.HasPrefix(dataDir, "/Library/Application Support"))
}

func Test_PlatformLinux(t *testing.T) {
	fakeAppName := "test-linux-app"
	plat, err := NewPlatform(fakeAppName, "linux")

	require.NoError(t, err)
	require.NotNil(t, plat)

	linux, ok := plat.(*PlatformLinux)
	require.True(t, ok)
	require.NotNil(t, linux)

	// user scoped
	configDir := linux.UserAppConfigDirectory()
	assert.True(t, strings.HasSuffix(configDir, fakeAppName))
	assert.True(t, strings.Contains(configDir, "/.config"))
	assert.False(t, strings.HasPrefix(configDir, "/.config"))
	assert.False(t, strings.HasPrefix(configDir, "/etc"))

	dataDir := linux.UserAppDataDirectory()
	assert.True(t, strings.HasSuffix(dataDir, fakeAppName))
	assert.True(t, strings.Contains(dataDir, "/.local/share"))
	assert.False(t, strings.HasPrefix(dataDir, "/var/lib"))
	assert.False(t, strings.HasPrefix(dataDir, "/.local/share"))

	// system scoped
	configDir = linux.SystemAppConfigDirectory()
	assert.True(t, strings.HasSuffix(configDir, fakeAppName))
	assert.True(t, strings.HasPrefix(configDir, "/etc"))

	dataDir = linux.SystemAppDataDirectory()
	assert.True(t, strings.HasSuffix(dataDir, fakeAppName))
	assert.True(t, strings.HasPrefix(dataDir, "/var/lib"))
}

// test utility to convert path to windows format and simplify cross-platform testing
func convertToWindowsPath(path string) string {
	return strings.ReplaceAll(path, "/", `\`)
}

func Test_PlatformWindows(t *testing.T) {
	fakeDrive := "FakeCDrive:\\"
	// set up fake environment variables
	t.Setenv(envKeyLocalAppData, fakeDrive+"Users\\test\\AppData\\Local")
	t.Setenv(envKeyProgramData, fakeDrive+"ProgramData")
	t.Setenv(envKeyProgramFiles, fakeDrive+"ProgramFiles")

	fakeAppName := "test-windows-app"
	plat, err := NewPlatform(fakeAppName, "windows")

	require.NoError(t, err)
	require.NotNil(t, plat)

	windows, ok := plat.(*PlatformWindows)
	require.True(t, ok)
	require.NotNil(t, windows)

	// user scoped
	configDir := windows.UserAppConfigDirectory()
	assert.True(t, strings.HasSuffix(configDir, fakeAppName))
	assert.True(t, strings.Contains(configDir, "AppData\\Local"))
	assert.True(t, strings.HasPrefix(configDir, fakeDrive))

	dataDir := windows.UserAppDataDirectory()
	assert.True(t, strings.HasSuffix(dataDir, fakeAppName))
	assert.True(t, strings.Contains(dataDir, "AppData\\Local"))
	assert.True(t, strings.HasPrefix(dataDir, fakeDrive))

	// system scoped
	configDir = windows.SystemAppConfigDirectory()
	configDir = convertToWindowsPath(configDir)
	assert.True(t, strings.HasSuffix(configDir, fakeAppName))
	assert.True(t, strings.HasPrefix(configDir, fakeDrive+"ProgramFiles\\"))

	dataDir = windows.SystemAppDataDirectory()
	dataDir = convertToWindowsPath(dataDir)
	assert.True(t, strings.HasSuffix(dataDir, fakeAppName))
	assert.True(t, strings.HasPrefix(dataDir, fakeDrive+"ProgramData\\"))
}
