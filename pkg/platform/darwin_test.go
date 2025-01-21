//go:build darwin
// +build darwin

package platform

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_PlatformMacOS(t *testing.T) {
	fakeAppName := "test-macos-app"
	plat, err := NewPlatform(fakeAppName)

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

	logger := darwin.GetLogger()
	assert.NotNil(t, logger)
	logger.Info("Testing macOS logger VIRTRU")
}
