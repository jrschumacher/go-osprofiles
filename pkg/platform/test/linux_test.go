//go:build linux
// +build linux

package test

import (
	"strings"
	"testing"

	"github.com/jrschumacher/go-osprofiles/pkg/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_PlatformLinux(t *testing.T) {
	fakeAppName := "test-linux-app"
	plat, err := platform.NewPlatform(fakeAppName)

	require.NoError(t, err)
	require.NotNil(t, plat)

	linux, ok := plat.(*platform.PlatformLinux)
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

	logger := linux.Logger()
	assert.NotNil(t, logger)
	logger.Info("Testing Linux logger")
}
