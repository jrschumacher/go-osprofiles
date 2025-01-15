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

	configDir := darwin.GetConfigDirectory()
	assert.True(t, strings.HasSuffix(configDir, fakeAppName))

	dataDir := darwin.GetDataDirectory()
	assert.True(t, strings.HasSuffix(dataDir, fakeAppName))
	assert.True(t, strings.Contains(dataDir, "Library/Application Support"))
}

func Test_PlatformLinux(t *testing.T) {
	fakeAppName := "test-linux-app"
	plat, err := NewPlatform(fakeAppName, "linux")

	require.NoError(t, err)
	require.NotNil(t, plat)

	linux, ok := plat.(*PlatformLinux)
	require.True(t, ok)
	require.NotNil(t, linux)

	configDir := linux.GetConfigDirectory()
	assert.True(t, strings.HasSuffix(configDir, fakeAppName))
	assert.True(t, strings.Contains(configDir, ".config"))

	dataDir := linux.GetDataDirectory()
	assert.True(t, strings.HasSuffix(dataDir, fakeAppName))
	assert.True(t, strings.Contains(dataDir, ".local/share"))

	logger := linux.GetLogger()
	assert.NotNil(t, logger)
	logger.Info("Testing Linux logger")
}

func Test_PlatformWindows(t *testing.T) {
	fakeAppName := "test-windows-app"
	plat, err := NewPlatform(fakeAppName, "windows")

	require.NoError(t, err)
	require.NotNil(t, plat)

	windows, ok := plat.(*PlatformWindows)
	require.True(t, ok)
	require.NotNil(t, windows)

	configDir := windows.GetConfigDirectory()
	assert.True(t, strings.HasSuffix(configDir, fakeAppName))

	// TODO: tests according to different OS versions
}
