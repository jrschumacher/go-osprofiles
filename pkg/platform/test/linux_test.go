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
