package platform

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type LinuxTestSuite struct {
	suite.Suite
	tempDir          string
	originalEnv      string
	originalHomeDir  string
	fakeHomeDir      string
}

func (suite *LinuxTestSuite) SetupSuite() {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "linux-platform-test-*")
	suite.Require().NoError(err)
	suite.tempDir = tempDir

	// Store original environment variable
	suite.originalEnv = os.Getenv("OSPROFILES_TEST_BASE_PATH")

	// Set environment variable to use temp directory
	os.Setenv("OSPROFILES_TEST_BASE_PATH", tempDir)

	// Create fake home directory for user tests
	suite.fakeHomeDir = filepath.Join(tempDir, "home", "testuser")
	err = os.MkdirAll(suite.fakeHomeDir, 0755)
	suite.Require().NoError(err)

	// Create test directory structure for system paths
	err = os.MkdirAll(filepath.Join(tempDir, "usr", "local"), 0755)
	suite.Require().NoError(err)
	err = os.MkdirAll(filepath.Join(tempDir, "etc"), 0755)
	suite.Require().NoError(err)

	// Create user config and data directories
	err = os.MkdirAll(filepath.Join(suite.fakeHomeDir, ".config"), 0755)
	suite.Require().NoError(err)
	err = os.MkdirAll(filepath.Join(suite.fakeHomeDir, ".local", "share"), 0755)
	suite.Require().NoError(err)
}

func (suite *LinuxTestSuite) TearDownSuite() {
	// Restore original environment variable
	if suite.originalEnv != "" {
		os.Setenv("OSPROFILES_TEST_BASE_PATH", suite.originalEnv)
	} else {
		os.Unsetenv("OSPROFILES_TEST_BASE_PATH")
	}

	// Clean up temp directory
	os.RemoveAll(suite.tempDir)
}

func (suite *LinuxTestSuite) createTestPlatform(publisher, namespace string) *PlatformLinux {
	platform := &PlatformLinux{
		username:         "testuser",
		serviceNamespace: namespace,
		servicePublisher: publisher,
		userHomeDir:      suite.fakeHomeDir,
	}
	return platform
}

func TestLinuxTestSuite(t *testing.T) {
	suite.Run(t, new(LinuxTestSuite))
}

// Test basic functionality
func (suite *LinuxTestSuite) TestGetUsername() {
	platform := suite.createTestPlatform("", "testapp")
	assert.Equal(suite.T(), "testuser", platform.GetUsername())
}

func (suite *LinuxTestSuite) TestUserHomeDir() {
	platform := suite.createTestPlatform("", "testapp")
	assert.Equal(suite.T(), suite.fakeHomeDir, platform.UserHomeDir())
}

// Test user directories without publisher
func (suite *LinuxTestSuite) TestUserAppDataDirectory_NoPublisher() {
	platform := suite.createTestPlatform("", "myapp")
	
	dataDir := platform.UserAppDataDirectory()
	expected := filepath.Join(suite.fakeHomeDir, ".local", "share", "myapp")
	
	assert.Equal(suite.T(), expected, dataDir)
	assert.True(suite.T(), strings.HasSuffix(dataDir, "myapp"))
	assert.True(suite.T(), strings.Contains(dataDir, ".local/share"))
}

func (suite *LinuxTestSuite) TestUserAppConfigDirectory_NoPublisher() {
	platform := suite.createTestPlatform("", "myapp")
	
	configDir := platform.UserAppConfigDirectory()
	expected := filepath.Join(suite.fakeHomeDir, ".config", "myapp")
	
	assert.Equal(suite.T(), expected, configDir)
	assert.True(suite.T(), strings.HasSuffix(configDir, "myapp"))
	assert.True(suite.T(), strings.Contains(configDir, ".config"))
}

// Test user directories with publisher
func (suite *LinuxTestSuite) TestUserAppDataDirectory_WithPublisher() {
	platform := suite.createTestPlatform("com.company", "myapp")
	
	dataDir := platform.UserAppDataDirectory()
	expected := filepath.Join(suite.fakeHomeDir, ".local", "share", "com.company", "myapp")
	
	assert.Equal(suite.T(), expected, dataDir)
	assert.True(suite.T(), strings.HasSuffix(dataDir, "myapp"))
	assert.True(suite.T(), strings.Contains(dataDir, "com.company"))
	assert.True(suite.T(), strings.Contains(dataDir, ".local/share"))
}

func (suite *LinuxTestSuite) TestUserAppConfigDirectory_WithPublisher() {
	platform := suite.createTestPlatform("com.company", "myapp")
	
	configDir := platform.UserAppConfigDirectory()
	expected := filepath.Join(suite.fakeHomeDir, ".config", "com.company", "myapp")
	
	assert.Equal(suite.T(), expected, configDir)
	assert.True(suite.T(), strings.HasSuffix(configDir, "myapp"))
	assert.True(suite.T(), strings.Contains(configDir, "com.company"))
	assert.True(suite.T(), strings.Contains(configDir, ".config"))
}

// Test system directories without publisher
func (suite *LinuxTestSuite) TestSystemAppDataDirectory_NoPublisher() {
	platform := suite.createTestPlatform("", "myapp")
	
	dataDir := platform.SystemAppDataDirectory()
	expected := filepath.Join(suite.tempDir, "usr", "local", "myapp")
	
	assert.Equal(suite.T(), expected, dataDir)
	assert.True(suite.T(), strings.HasSuffix(dataDir, "myapp"))
	assert.True(suite.T(), strings.Contains(dataDir, "usr/local"))
}

func (suite *LinuxTestSuite) TestSystemAppConfigDirectory_NoPublisher() {
	platform := suite.createTestPlatform("", "myapp")
	
	configDir := platform.SystemAppConfigDirectory()
	expected := filepath.Join(suite.tempDir, "etc", "myapp")
	
	assert.Equal(suite.T(), expected, configDir)
	assert.True(suite.T(), strings.HasSuffix(configDir, "myapp"))
	assert.True(suite.T(), strings.Contains(configDir, "etc"))
}

// Test system directories with publisher
func (suite *LinuxTestSuite) TestSystemAppDataDirectory_WithPublisher() {
	platform := suite.createTestPlatform("com.company", "myapp")
	
	dataDir := platform.SystemAppDataDirectory()
	expected := filepath.Join(suite.tempDir, "usr", "local", "com.company", "myapp")
	
	assert.Equal(suite.T(), expected, dataDir)
	assert.True(suite.T(), strings.HasSuffix(dataDir, "myapp"))
	assert.True(suite.T(), strings.Contains(dataDir, "com.company"))
	assert.True(suite.T(), strings.Contains(dataDir, "usr/local"))
}

func (suite *LinuxTestSuite) TestSystemAppConfigDirectory_WithPublisher() {
	platform := suite.createTestPlatform("com.company", "myapp")
	
	configDir := platform.SystemAppConfigDirectory()
	expected := filepath.Join(suite.tempDir, "etc", "com.company", "myapp")
	
	assert.Equal(suite.T(), expected, configDir)
	assert.True(suite.T(), strings.HasSuffix(configDir, "myapp"))
	assert.True(suite.T(), strings.Contains(configDir, "com.company"))
	assert.True(suite.T(), strings.Contains(configDir, "etc"))
}

// Test MDM functionality (should return empty/false since not supported)
func (suite *LinuxTestSuite) TestMDMConfigPath() {
	platform := suite.createTestPlatform("com.company", "myapp")
	
	mdmPath := platform.MDMConfigPath()
	assert.Empty(suite.T(), mdmPath)
}

func (suite *LinuxTestSuite) TestMDMConfigExists() {
	platform := suite.createTestPlatform("com.company", "myapp")
	
	exists := platform.MDMConfigExists()
	assert.False(suite.T(), exists)
}

func (suite *LinuxTestSuite) TestGetMDMData() {
	platform := suite.createTestPlatform("com.company", "myapp")
	
	data, err := platform.GetMDMData()
	assert.Nil(suite.T(), data)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "MDM is not supported on Linux")
}

func (suite *LinuxTestSuite) TestGetMDMDataAsJSON() {
	platform := suite.createTestPlatform("com.company", "myapp")
	
	data, err := platform.GetMDMDataAsJSON(false)
	assert.Nil(suite.T(), data)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "MDM is not supported on Linux")
}

// Test SystemAppDataDirectoryWithMDM
func (suite *LinuxTestSuite) TestSystemAppDataDirectoryWithMDM_NoPublisher() {
	platform := suite.createTestPlatform("", "myapp")
	
	systemDir, opts := platform.SystemAppDataDirectoryWithMDM()
	expected := filepath.Join(suite.tempDir, "usr", "local", "myapp")
	
	assert.Equal(suite.T(), expected, systemDir)
	assert.Len(suite.T(), opts, 1) // Should only have WithStoreDirectory option
}

func (suite *LinuxTestSuite) TestSystemAppDataDirectoryWithMDM_WithPublisher() {
	platform := suite.createTestPlatform("com.company", "myapp")
	
	systemDir, opts := platform.SystemAppDataDirectoryWithMDM("com.company.myapp")
	expected := filepath.Join(suite.tempDir, "usr", "local", "com.company", "myapp")
	
	assert.Equal(suite.T(), expected, systemDir)
	assert.Len(suite.T(), opts, 1) // Should only have WithStoreDirectory option, no MDM on Linux
}

// Test path construction edge cases
func (suite *LinuxTestSuite) TestEmptyPublisher() {
	platform := suite.createTestPlatform("", "myapp")
	
	// User directories
	userDataDir := platform.UserAppDataDirectory()
	userConfigDir := platform.UserAppConfigDirectory()
	
	assert.False(suite.T(), strings.Contains(userDataDir, "//")) // No double slashes
	assert.False(suite.T(), strings.Contains(userConfigDir, "//"))
	
	// System directories
	systemDataDir := platform.SystemAppDataDirectory()
	systemConfigDir := platform.SystemAppConfigDirectory()
	
	assert.False(suite.T(), strings.Contains(systemDataDir, "//"))
	assert.False(suite.T(), strings.Contains(systemConfigDir, "//"))
}

func (suite *LinuxTestSuite) TestEmptyNamespace() {
	platform := suite.createTestPlatform("com.company", "")
	
	// Should still work with empty namespace
	userDataDir := platform.UserAppDataDirectory()
	assert.True(suite.T(), strings.Contains(userDataDir, "com.company"))
	assert.True(suite.T(), strings.HasSuffix(userDataDir, "com.company")) // namespace is empty, so ends with publisher
}

// Test environment variable override behavior
func (suite *LinuxTestSuite) TestEnvironmentOverride() {
	platform := suite.createTestPlatform("com.company", "myapp")
	
	// System directories should use the test environment override
	systemDataDir := platform.SystemAppDataDirectory()
	systemConfigDir := platform.SystemAppConfigDirectory()
	
	assert.True(suite.T(), strings.HasPrefix(systemDataDir, suite.tempDir))
	assert.True(suite.T(), strings.HasPrefix(systemConfigDir, suite.tempDir))
	
	// User directories should use the fake home dir (not affected by test env override)
	userDataDir := platform.UserAppDataDirectory()
	userConfigDir := platform.UserAppConfigDirectory()
	
	assert.True(suite.T(), strings.HasPrefix(userDataDir, suite.fakeHomeDir))
	assert.True(suite.T(), strings.HasPrefix(userConfigDir, suite.fakeHomeDir))
}