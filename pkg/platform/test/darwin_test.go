//go:build darwin
// +build darwin

package test

import (
	"strings"
	"testing"

	"github.com/jrschumacher/go-osprofiles/pkg/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_PlatformMacOS(t *testing.T) {
	fakeAppName := "test-macos-app"
	plat, err := platform.NewPlatform(fakeAppName)

	require.NoError(t, err)
	require.NotNil(t, plat)

	darwin, ok := plat.(*platform.PlatformDarwin)
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

	logger := darwin.Logger()
	assert.NotNil(t, logger)
	logger.Info("Testing macOS logger")
}
