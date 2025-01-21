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

func Test_PlatformWindows(t *testing.T) {
	fakeAppName := "test-windows-app"
	plat, err := platform.NewPlatform(fakeAppName)

	require.NoError(t, err)
	require.NotNil(t, plat)

	windows, ok := plat.(*platform.PlatformWindows)
	require.True(t, ok)
	require.NotNil(t, windows)

	configDir := windows.GetConfigDirectory()
	assert.True(t, strings.HasSuffix(configDir, fakeAppName))

	// TODO: tests according to different OS versions
	logger := windows.GetLogger()
	assert.NotNil(t, logger)
	logger.Info("Testing Windows logger")
}
