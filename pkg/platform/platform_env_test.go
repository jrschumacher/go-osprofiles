package platform

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type EnvironmentTestSuite struct {
	suite.Suite
	originalEnvValue string
	testBasePath     string
}

func (suite *EnvironmentTestSuite) SetupTest() {
	// Save original environment value
	suite.originalEnvValue = os.Getenv("OSPROFILES_TEST_BASE_PATH")
	
	// Set up test base path
	suite.testBasePath = "/tmp/osprofiles-test"
}

func (suite *EnvironmentTestSuite) TearDownTest() {
	// Restore original environment value
	if suite.originalEnvValue == "" {
		os.Unsetenv("OSPROFILES_TEST_BASE_PATH")
	} else {
		os.Setenv("OSPROFILES_TEST_BASE_PATH", suite.originalEnvValue)
	}
}

func (suite *EnvironmentTestSuite) TestDarwinWithoutTestBasePath() {
	if runtime.GOOS != "darwin" {
		suite.T().Skip("Skipping Darwin-specific test on non-macOS")
	}
	
	// Ensure environment variable is not set
	os.Unsetenv("OSPROFILES_TEST_BASE_PATH")
	
	platform, err := NewPlatform("com.company", "myapp", "darwin")
	assert.NoError(suite.T(), err)
	
	// Should use default system paths
	systemDir := platform.SystemAppDataDirectory()
	expectedSystemDir := "/Library/Application Support/com.company/myapp"
	assert.Equal(suite.T(), expectedSystemDir, systemDir)
	
	mdmPath := platform.MDMConfigPath()
	expectedMDMPath := "/Library/Managed Preferences/com.company.myapp.plist"
	assert.Equal(suite.T(), expectedMDMPath, mdmPath)
}

func (suite *EnvironmentTestSuite) TestDarwinWithTestBasePath() {
	if runtime.GOOS != "darwin" {
		suite.T().Skip("Skipping Darwin-specific test on non-macOS")
	}
	
	// Set test base path
	os.Setenv("OSPROFILES_TEST_BASE_PATH", suite.testBasePath)
	
	platform, err := NewPlatform("com.company", "myapp", "darwin")
	assert.NoError(suite.T(), err)
	
	// Should use test base path
	systemDir := platform.SystemAppDataDirectory()
	expectedSystemDir := filepath.Join(suite.testBasePath, "Library", "Application Support", "com.company", "myapp")
	assert.Equal(suite.T(), expectedSystemDir, systemDir)
	
	mdmPath := platform.MDMConfigPath()
	expectedMDMPath := filepath.Join(suite.testBasePath, "Library", "Managed Preferences", "com.company.myapp.plist")
	assert.Equal(suite.T(), expectedMDMPath, mdmPath)
}

func (suite *EnvironmentTestSuite) TestLinuxWithoutTestBasePath() {
	// Ensure environment variable is not set
	os.Unsetenv("OSPROFILES_TEST_BASE_PATH")
	
	platform, err := NewPlatform("com.company", "myapp", "linux")
	assert.NoError(suite.T(), err)
	
	// Should use default system paths
	systemDir := platform.SystemAppDataDirectory()
	expectedSystemDir := "/usr/local/com.company/myapp"
	assert.Equal(suite.T(), expectedSystemDir, systemDir)
	
	configDir := platform.SystemAppConfigDirectory()
	expectedConfigDir := "/etc/com.company/myapp"
	assert.Equal(suite.T(), expectedConfigDir, configDir)
	
	// MDM not supported on Linux
	mdmPath := platform.MDMConfigPath()
	assert.Equal(suite.T(), "", mdmPath)
}

func (suite *EnvironmentTestSuite) TestLinuxWithTestBasePath() {
	// Set test base path
	os.Setenv("OSPROFILES_TEST_BASE_PATH", suite.testBasePath)
	
	platform, err := NewPlatform("com.company", "myapp", "linux")
	assert.NoError(suite.T(), err)
	
	// Should use test base path
	systemDir := platform.SystemAppDataDirectory()
	expectedSystemDir := filepath.Join(suite.testBasePath, "usr", "local", "com.company", "myapp")
	assert.Equal(suite.T(), expectedSystemDir, systemDir)
	
	configDir := platform.SystemAppConfigDirectory()
	expectedConfigDir := filepath.Join(suite.testBasePath, "etc", "com.company", "myapp")
	assert.Equal(suite.T(), expectedConfigDir, configDir)
	
	// MDM not supported on Linux
	mdmPath := platform.MDMConfigPath()
	assert.Equal(suite.T(), "", mdmPath)
}

func (suite *EnvironmentTestSuite) TestWindowsWithoutTestBasePath() {
	// This test can run on any platform since we're not calling Windows-specific APIs
	platform, err := NewPlatform("com.company", "myapp", "windows")
	assert.NoError(suite.T(), err)
	
	// Ensure environment variable is not set
	os.Unsetenv("OSPROFILES_TEST_BASE_PATH")
	
	// Should use default Windows paths (based on constructor values)
	systemDir := platform.SystemAppDataDirectory()
	// The exact path depends on Windows environment variables, but it should contain the namespace
	assert.Contains(suite.T(), systemDir, "com.company")
	assert.Contains(suite.T(), systemDir, "myapp")
	
	// MDM not supported on Windows
	mdmPath := platform.MDMConfigPath()
	assert.Equal(suite.T(), "", mdmPath)
}

func (suite *EnvironmentTestSuite) TestWindowsWithTestBasePath() {
	// Set test base path
	os.Setenv("OSPROFILES_TEST_BASE_PATH", suite.testBasePath)
	
	platform, err := NewPlatform("com.company", "myapp", "windows")
	assert.NoError(suite.T(), err)
	
	// Should use test base path
	systemDir := platform.SystemAppDataDirectory()
	expectedSystemDir := filepath.Join(suite.testBasePath, "ProgramData", "com.company", "myapp")
	assert.Equal(suite.T(), expectedSystemDir, systemDir)
	
	configDir := platform.SystemAppConfigDirectory()
	expectedConfigDir := filepath.Join(suite.testBasePath, "ProgramFiles", "com.company", "myapp")
	assert.Equal(suite.T(), expectedConfigDir, configDir)
	
	// MDM not supported on Windows
	mdmPath := platform.MDMConfigPath()
	assert.Equal(suite.T(), "", mdmPath)
}

func (suite *EnvironmentTestSuite) TestSystemAppDataDirectoryWithMDMUsesTestBasePath() {
	if runtime.GOOS != "darwin" {
		suite.T().Skip("Skipping Darwin-specific test on non-macOS")
	}
	
	// Set test base path
	os.Setenv("OSPROFILES_TEST_BASE_PATH", suite.testBasePath)
	
	platform, err := NewPlatform("com.company", "myapp", "darwin")
	assert.NoError(suite.T(), err)
	
	// SystemAppDataDirectoryWithMDM should also use test base path
	systemDir, opts := platform.SystemAppDataDirectoryWithMDM()
	expectedSystemDir := filepath.Join(suite.testBasePath, "Library", "Application Support", "com.company", "myapp")
	assert.Equal(suite.T(), expectedSystemDir, systemDir)
	assert.Len(suite.T(), opts, 2) // Should have WithStoreDirectory and WithMDMSupport
}

func (suite *EnvironmentTestSuite) TestGetTestBasePathFunction() {
	// Test the utility function directly
	
	// Without environment variable
	os.Unsetenv("OSPROFILES_TEST_BASE_PATH")
	result := getTestBasePath()
	assert.Equal(suite.T(), "", result)
	
	// With environment variable
	os.Setenv("OSPROFILES_TEST_BASE_PATH", suite.testBasePath)
	result = getTestBasePath()
	assert.Equal(suite.T(), suite.testBasePath, result)
}

func (suite *EnvironmentTestSuite) TestUserDirectoriesNotAffectedByTestBasePath() {
	// User directories should NOT be affected by OSPROFILES_TEST_BASE_PATH
	// Only system directories should be overridden
	
	os.Setenv("OSPROFILES_TEST_BASE_PATH", suite.testBasePath)
	
	platform, err := NewPlatform("com.company", "myapp", runtime.GOOS)
	assert.NoError(suite.T(), err)
	
	userDir := platform.UserAppDataDirectory()
	userHomeDir := platform.UserHomeDir()
	
	// User directories should still use real user paths, not test paths
	assert.NotContains(suite.T(), userDir, suite.testBasePath)
	assert.NotContains(suite.T(), userHomeDir, suite.testBasePath)
	
	// Should contain actual user path elements
	switch runtime.GOOS {
	case "darwin":
		assert.Contains(suite.T(), userDir, "Library/Application Support")
	case "linux":
		// Could be ~/.local/share or ~/.config depending on implementation
		assert.Contains(suite.T(), userDir, userHomeDir)
	case "windows":
		// Could be AppData/Local or AppData/Roaming
		assert.Contains(suite.T(), userDir, userHomeDir)
	}
}

func TestEnvironmentTestSuite(t *testing.T) {
	suite.Run(t, new(EnvironmentTestSuite))
}