package platform

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_PlatformMacOS_NoPublisher(t *testing.T) {
	fakeAppName := "test-macos-app"
	var publisher string
	plat, err := NewPlatform(publisher, fakeAppName, "darwin")

	require.NoError(t, err)
	require.NotNil(t, plat)

	darwin, ok := plat.(*PlatformDarwin)
	require.True(t, ok)
	require.NotNil(t, darwin)

	// user scoped
	configDir := darwin.UserAppConfigDirectory()
	assert.True(t, strings.HasSuffix(configDir, fakeAppName))
	assert.True(t, strings.Contains(configDir, fmt.Sprintf("/Library/Application Support/%s", fakeAppName)))
	assert.False(t, strings.HasPrefix(configDir, "/Library/Application Support"))

	dataDir := darwin.UserAppDataDirectory()
	assert.True(t, strings.HasSuffix(dataDir, fakeAppName))
	assert.True(t, strings.Contains(dataDir, fmt.Sprintf("/Library/Application Support/%s", fakeAppName)))
	assert.False(t, strings.HasPrefix(configDir, "/Library/Application Support"))

	// system scoped
	configDir = darwin.SystemAppConfigDirectory()
	assert.True(t, strings.HasSuffix(configDir, fakeAppName))
	assert.Equal(t, fmt.Sprintf("/Library/Application Support/%s", fakeAppName), configDir)

	dataDir = darwin.SystemAppDataDirectory()
	assert.True(t, strings.HasSuffix(dataDir, fakeAppName))
	assert.Equal(t, fmt.Sprintf("/Library/Application Support/%s", fakeAppName), dataDir)
}

func Test_PlatformMacOS_WithPublisher(t *testing.T) {
	fakeAppName := "test-with-publisher-app"
	fakePublisher := "test-publisher"
	plat, err := NewPlatform(fakePublisher, fakeAppName, "darwin")

	require.NoError(t, err)
	require.NotNil(t, plat)

	darwin, ok := plat.(*PlatformDarwin)
	require.True(t, ok)
	require.NotNil(t, darwin)

	// user scoped
	configDir := darwin.UserAppConfigDirectory()
	assert.True(t, strings.HasSuffix(configDir, fakeAppName))
	assert.True(t, strings.Contains(configDir, fmt.Sprintf("/Library/Application Support/%s/%s", fakePublisher, fakeAppName)))
	assert.False(t, strings.HasPrefix(configDir, "/Library/Application Support"))

	dataDir := darwin.UserAppDataDirectory()
	assert.True(t, strings.HasSuffix(dataDir, fakeAppName))
	assert.True(t, strings.Contains(dataDir, fmt.Sprintf("/Library/Application Support/%s/%s", fakePublisher, fakeAppName)))
	assert.False(t, strings.HasPrefix(dataDir, "/Library/Application Support"))

	// system scoped
	configDir = darwin.SystemAppConfigDirectory()
	assert.True(t, strings.HasSuffix(configDir, fakeAppName))
	assert.Equal(t, fmt.Sprintf("/Library/Application Support/%s/%s", fakePublisher, fakeAppName), configDir)

	dataDir = darwin.SystemAppDataDirectory()
	assert.True(t, strings.HasSuffix(dataDir, fakeAppName))
	assert.Equal(t, fmt.Sprintf("/Library/Application Support/%s/%s", fakePublisher, fakeAppName), dataDir)
}

func Test_PlatformLinux_NoPublisher(t *testing.T) {
	fakeAppName := "test-linux-app"
	var publisher string
	plat, err := NewPlatform(publisher, fakeAppName, "linux")

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
	assert.False(t, strings.HasPrefix(dataDir, "/usr/local"))
	assert.False(t, strings.HasPrefix(dataDir, "/.local/share"))

	// system scoped
	configDir = linux.SystemAppConfigDirectory()
	assert.True(t, strings.HasSuffix(configDir, fakeAppName))
	assert.Equal(t, fmt.Sprintf("/etc/%s", fakeAppName), configDir)

	dataDir = linux.SystemAppDataDirectory()
	assert.True(t, strings.HasSuffix(dataDir, fakeAppName))
	assert.Equal(t, fmt.Sprintf("/usr/local/%s", fakeAppName), dataDir)
}

func Test_PlatformLinux_WithPublisher(t *testing.T) {
	fakeAppName := "test-publisher-linux-app"
	fakeAppPublisher := "test-publisher"
	plat, err := NewPlatform(fakeAppPublisher, fakeAppName, "linux")

	require.NoError(t, err)
	require.NotNil(t, plat)

	linux, ok := plat.(*PlatformLinux)
	require.True(t, ok)
	require.NotNil(t, linux)

	// user scoped
	configDir := linux.UserAppConfigDirectory()
	assert.True(t, strings.HasSuffix(configDir, fakeAppName))
	assert.True(t, strings.Contains(configDir, fmt.Sprintf("/.config/%s/%s", fakeAppPublisher, fakeAppName)))
	assert.False(t, strings.HasPrefix(configDir, "/.config"))
	assert.False(t, strings.HasPrefix(configDir, "/etc"))

	dataDir := linux.UserAppDataDirectory()
	assert.True(t, strings.HasSuffix(dataDir, fakeAppName))
	assert.True(t, strings.Contains(dataDir, fmt.Sprintf("/.local/share/%s/%s", fakeAppPublisher, fakeAppName)))
	assert.False(t, strings.HasPrefix(dataDir, "/usr/local"))
	assert.False(t, strings.HasPrefix(dataDir, "/.local/share"))

	// system scoped
	configDir = linux.SystemAppConfigDirectory()
	assert.True(t, strings.HasSuffix(configDir, fakeAppName))
	assert.Equal(t, fmt.Sprintf("/etc/%s/%s", fakeAppPublisher, fakeAppName), configDir)

	dataDir = linux.SystemAppDataDirectory()
	assert.True(t, strings.HasSuffix(dataDir, fakeAppName))
	assert.Equal(t, fmt.Sprintf("/usr/local/%s/%s", fakeAppPublisher, fakeAppName), dataDir)
}

// test utility to convert path to windows format and simplify cross-platform testing
func convertToWindowsPath(path string) string {
	return strings.ReplaceAll(path, "/", `\`)
}

func Test_PlatformWindows_NoPublisher(t *testing.T) {
	fakeDrive := "FakeCDrive:\\"
	// set up fake environment variables
	t.Setenv(envKeyLocalAppData, fakeDrive+"Users\\test\\AppData\\Local")
	t.Setenv(envKeyProgramData, fakeDrive+"ProgramData")
	t.Setenv(envKeyProgramFiles, fakeDrive+"ProgramFiles")

	fakeAppName := "test-windows-app"
	var publisher string
	plat, err := NewPlatform(publisher, fakeAppName, "windows")

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

func Test_PlatformWindows_WithPublisher(t *testing.T) {
	fakeDrive := "FakeCDrive:\\"
	// set up fake environment variables
	t.Setenv(envKeyLocalAppData, fakeDrive+"Users\\test\\AppData\\Local")
	t.Setenv(envKeyProgramData, fakeDrive+"ProgramData")
	t.Setenv(envKeyProgramFiles, fakeDrive+"ProgramFiles")

	fakeAppName := "test-publisher-windows-app"
	fakeAppPublisher := "test-publisher"
	plat, err := NewPlatform(fakeAppPublisher, fakeAppName, "windows")

	require.NoError(t, err)
	require.NotNil(t, plat)

	windows, ok := plat.(*PlatformWindows)
	require.True(t, ok)
	require.NotNil(t, windows)

	// user scoped
	configDir := windows.UserAppConfigDirectory()
	configDir = convertToWindowsPath(configDir)
	assert.True(t, strings.HasSuffix(configDir, fakeAppName))
	assert.True(t, strings.Contains(configDir, fmt.Sprintf("AppData\\Local\\%s", fakeAppPublisher)))
	assert.True(t, strings.HasPrefix(configDir, fakeDrive))

	dataDir := windows.UserAppDataDirectory()
	dataDir = convertToWindowsPath(dataDir)
	assert.True(t, strings.HasSuffix(dataDir, fakeAppName))
	assert.True(t, strings.Contains(dataDir, fmt.Sprintf("AppData\\Local\\%s", fakeAppPublisher)))
	assert.True(t, strings.HasPrefix(dataDir, fakeDrive))

	// system scoped
	configDir = windows.SystemAppConfigDirectory()
	configDir = convertToWindowsPath(configDir)
	assert.True(t, strings.HasSuffix(configDir, fakeAppName))
	assert.Equal(t, fmt.Sprintf("%sProgramFiles\\%s\\%s", fakeDrive, fakeAppPublisher, fakeAppName), configDir)

	dataDir = windows.SystemAppDataDirectory()
	dataDir = convertToWindowsPath(dataDir)
	assert.True(t, strings.HasSuffix(dataDir, fakeAppName))
	assert.Equal(t, fmt.Sprintf("%sProgramData\\%s\\%s", fakeDrive, fakeAppPublisher, fakeAppName), dataDir)
}
