//go:build windows
// +build windows

package platform

import (
	"strings"
	"testing"

	"github.com/jrschumacher/go-osprofiles/pkg/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// test utility to convert path to windows format and simplify cross-platform testing
func convertToWindowsPath(path string) string {
	return strings.ReplaceAll(path, "/", `\`)
}

func Test_PlatformWindows(t *testing.T) {
	fakeDrive := "FakeCDrive:\\"
	// set up fake environment variables
	t.Setenv(platform.EnvKeyLocalAppData, fakeDrive+"Users\\test\\AppData\\Local")
	t.Setenv(platform.EnvKeyProgramData, fakeDrive+"ProgramData")
	t.Setenv(platform.EnvKeyProgramFiles, fakeDrive+"ProgramFiles")

	fakeAppName := "test-windows-app"
	plat, err := platform.NewPlatform(fakeAppName)

	require.NoError(t, err)
	require.NotNil(t, plat)

	windows, ok := plat.(*platform.PlatformWindows)
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

	logger := windows.Logger()
	assert.NotNil(t, logger)
	logger.Info("Testing Windows logger")
}
