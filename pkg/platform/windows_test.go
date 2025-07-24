package platform

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type WindowsTestSuite struct {
	suite.Suite
	tempDir             string
	originalEnv         string
	fakeHomeDir         string
	fakeLocalAppData    string
	fakeProgramData     string
	fakeProgramFiles    string
	originalEnvVars     map[string]string
}

func (suite *WindowsTestSuite) SetupSuite() {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "windows-platform-test-*")
	suite.Require().NoError(err)
	suite.tempDir = tempDir

	// Store original environment variables
	suite.originalEnv = os.Getenv("OSPROFILES_TEST_BASE_PATH")
	suite.originalEnvVars = map[string]string{
		envKeyLocalAppData: os.Getenv(envKeyLocalAppData),
		envKeyProgramData:  os.Getenv(envKeyProgramData),
		envKeyProgramFiles: os.Getenv(envKeyProgramFiles),
		envKeyUsername:     os.Getenv(envKeyUsername),
	}

	// Set environment variable to use temp directory
	os.Setenv("OSPROFILES_TEST_BASE_PATH", tempDir)

	// Create fake Windows directory structure
	suite.fakeHomeDir = filepath.Join(tempDir, "Users", "testuser")
	suite.fakeLocalAppData = filepath.Join(suite.fakeHomeDir, "AppData", "Local")
	suite.fakeProgramData = filepath.Join(tempDir, "ProgramData")
	suite.fakeProgramFiles = filepath.Join(tempDir, "ProgramFiles") // Note: using ProgramFiles to match util.go

	// Set up fake Windows environment variables
	os.Setenv(envKeyLocalAppData, suite.fakeLocalAppData)
	os.Setenv(envKeyProgramData, suite.fakeProgramData)
	os.Setenv(envKeyProgramFiles, suite.fakeProgramFiles)
	os.Setenv(envKeyUsername, "testuser")

	// Create test directory structure
	err = os.MkdirAll(suite.fakeLocalAppData, 0755)
	suite.Require().NoError(err)
	err = os.MkdirAll(suite.fakeProgramData, 0755)
	suite.Require().NoError(err)
	err = os.MkdirAll(suite.fakeProgramFiles, 0755)
	suite.Require().NoError(err)
	err = os.MkdirAll(suite.fakeHomeDir, 0755)
	suite.Require().NoError(err)
}

func (suite *WindowsTestSuite) TearDownSuite() {
	// Restore original environment variable
	if suite.originalEnv != "" {
		os.Setenv("OSPROFILES_TEST_BASE_PATH", suite.originalEnv)
	} else {
		os.Unsetenv("OSPROFILES_TEST_BASE_PATH")
	}

	// Restore original Windows environment variables
	for key, value := range suite.originalEnvVars {
		if value != "" {
			os.Setenv(key, value)
		} else {
			os.Unsetenv(key)
		}
	}

	// Clean up temp directory
	os.RemoveAll(suite.tempDir)
}

func (suite *WindowsTestSuite) createTestPlatform(publisher, namespace string) *PlatformWindows {
	platform := &PlatformWindows{
		username:         "testuser",
		serviceNamespace: namespace,
		servicePublisher: publisher,
		userHomeDir:      suite.fakeHomeDir,
		programFiles:     suite.fakeProgramFiles,
		programData:      suite.fakeProgramData,
		localAppData:     suite.fakeLocalAppData,
	}
	return platform
}

func TestWindowsTestSuite(t *testing.T) {
	suite.Run(t, new(WindowsTestSuite))
}

// Test basic functionality
func (suite *WindowsTestSuite) TestGetUsername() {
	platform := suite.createTestPlatform("", "testapp")
	assert.Equal(suite.T(), "testuser", platform.GetUsername())
}

func (suite *WindowsTestSuite) TestUserHomeDir() {
	platform := suite.createTestPlatform("", "testapp")
	assert.Equal(suite.T(), suite.fakeHomeDir, platform.UserHomeDir())
}

// Test user directories without publisher
func (suite *WindowsTestSuite) TestUserAppDataDirectory_NoPublisher() {
	platform := suite.createTestPlatform("", "myapp")
	
	dataDir := platform.UserAppDataDirectory()
	expected := filepath.Join(suite.fakeLocalAppData, "myapp")
	
	assert.Equal(suite.T(), expected, dataDir)
	assert.True(suite.T(), strings.HasSuffix(dataDir, "myapp"))
	assert.True(suite.T(), strings.Contains(dataDir, "AppData\\Local") || strings.Contains(dataDir, "AppData/Local"))
}

func (suite *WindowsTestSuite) TestUserAppConfigDirectory_NoPublisher() {
	platform := suite.createTestPlatform("", "myapp")
	
	configDir := platform.UserAppConfigDirectory()
	expected := filepath.Join(suite.fakeLocalAppData, "myapp")
	
	assert.Equal(suite.T(), expected, configDir)
	assert.True(suite.T(), strings.HasSuffix(configDir, "myapp"))
	assert.True(suite.T(), strings.Contains(configDir, "AppData\\Local") || strings.Contains(configDir, "AppData/Local"))
}

// Test user directories with publisher
func (suite *WindowsTestSuite) TestUserAppDataDirectory_WithPublisher() {
	platform := suite.createTestPlatform("Company", "myapp")
	
	dataDir := platform.UserAppDataDirectory()
	expected := filepath.Join(suite.fakeLocalAppData, "Company", "myapp")
	
	assert.Equal(suite.T(), expected, dataDir)
	assert.True(suite.T(), strings.HasSuffix(dataDir, "myapp"))
	assert.True(suite.T(), strings.Contains(dataDir, "Company"))
	assert.True(suite.T(), strings.Contains(dataDir, "AppData\\Local") || strings.Contains(dataDir, "AppData/Local"))
}

func (suite *WindowsTestSuite) TestUserAppConfigDirectory_WithPublisher() {
	platform := suite.createTestPlatform("Company", "myapp")
	
	configDir := platform.UserAppConfigDirectory()
	expected := filepath.Join(suite.fakeLocalAppData, "Company", "myapp")
	
	assert.Equal(suite.T(), expected, configDir)
	assert.True(suite.T(), strings.HasSuffix(configDir, "myapp"))
	assert.True(suite.T(), strings.Contains(configDir, "Company"))
	assert.True(suite.T(), strings.Contains(configDir, "AppData\\Local") || strings.Contains(configDir, "AppData/Local"))
}

// Test system directories without publisher
func (suite *WindowsTestSuite) TestSystemAppDataDirectory_NoPublisher() {
	platform := suite.createTestPlatform("", "myapp")
	
	dataDir := platform.SystemAppDataDirectory()
	expected := filepath.Join(suite.tempDir, "ProgramData", "myapp")
	
	assert.Equal(suite.T(), expected, dataDir)
	assert.True(suite.T(), strings.HasSuffix(dataDir, "myapp"))
	assert.True(suite.T(), strings.Contains(dataDir, "ProgramData"))
}

func (suite *WindowsTestSuite) TestSystemAppConfigDirectory_NoPublisher() {
	platform := suite.createTestPlatform("", "myapp")
	
	configDir := platform.SystemAppConfigDirectory()
	expected := filepath.Join(suite.tempDir, "ProgramFiles", "myapp")
	
	assert.Equal(suite.T(), expected, configDir)
	assert.True(suite.T(), strings.HasSuffix(configDir, "myapp"))
	assert.True(suite.T(), strings.Contains(configDir, "ProgramFiles"))
}

// Test system directories with publisher
func (suite *WindowsTestSuite) TestSystemAppDataDirectory_WithPublisher() {
	platform := suite.createTestPlatform("Company", "myapp")
	
	dataDir := platform.SystemAppDataDirectory()
	expected := filepath.Join(suite.tempDir, "ProgramData", "Company", "myapp")
	
	assert.Equal(suite.T(), expected, dataDir)
	assert.True(suite.T(), strings.HasSuffix(dataDir, "myapp"))
	assert.True(suite.T(), strings.Contains(dataDir, "Company"))
	assert.True(suite.T(), strings.Contains(dataDir, "ProgramData"))
}

func (suite *WindowsTestSuite) TestSystemAppConfigDirectory_WithPublisher() {
	platform := suite.createTestPlatform("Company", "myapp")
	
	configDir := platform.SystemAppConfigDirectory()
	expected := filepath.Join(suite.tempDir, "ProgramFiles", "Company", "myapp")
	
	assert.Equal(suite.T(), expected, configDir)
	assert.True(suite.T(), strings.HasSuffix(configDir, "myapp"))
	assert.True(suite.T(), strings.Contains(configDir, "Company"))
	assert.True(suite.T(), strings.Contains(configDir, "ProgramFiles"))
}

// Test MDM functionality (should return empty/false since not supported)
func (suite *WindowsTestSuite) TestMDMConfigPath() {
	platform := suite.createTestPlatform("com.company", "myapp")
	
	mdmPath := platform.MDMConfigPath()
	assert.Empty(suite.T(), mdmPath)
}

func (suite *WindowsTestSuite) TestMDMConfigExists() {
	platform := suite.createTestPlatform("com.company", "myapp")
	
	exists := platform.MDMConfigExists()
	assert.False(suite.T(), exists)
}

func (suite *WindowsTestSuite) TestGetMDMData() {
	platform := suite.createTestPlatform("com.company", "myapp")
	
	data, err := platform.GetMDMData()
	assert.Nil(suite.T(), data)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "MDM is not supported on Windows")
}

func (suite *WindowsTestSuite) TestGetMDMDataAsJSON() {
	platform := suite.createTestPlatform("com.company", "myapp")
	
	data, err := platform.GetMDMDataAsJSON(false)
	assert.Nil(suite.T(), data)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "MDM is not supported on Windows")
}

// Test SystemAppDataDirectoryWithMDM
func (suite *WindowsTestSuite) TestSystemAppDataDirectoryWithMDM_NoPublisher() {
	platform := suite.createTestPlatform("", "myapp")
	
	systemDir, opts := platform.SystemAppDataDirectoryWithMDM()
	expected := filepath.Join(suite.tempDir, "ProgramData", "myapp")
	
	assert.Equal(suite.T(), expected, systemDir)
	assert.Len(suite.T(), opts, 1) // Should only have WithStoreDirectory option
}

func (suite *WindowsTestSuite) TestSystemAppDataDirectoryWithMDM_WithPublisher() {
	platform := suite.createTestPlatform("Company", "myapp")
	
	systemDir, opts := platform.SystemAppDataDirectoryWithMDM("com.company.myapp")
	expected := filepath.Join(suite.tempDir, "ProgramData", "Company", "myapp")
	
	assert.Equal(suite.T(), expected, systemDir)
	assert.Len(suite.T(), opts, 1) // Should only have WithStoreDirectory option, no MDM on Windows
}

// Test path construction edge cases
func (suite *WindowsTestSuite) TestEmptyPublisher() {
	platform := suite.createTestPlatform("", "myapp")
	
	// User directories
	userDataDir := platform.UserAppDataDirectory()
	userConfigDir := platform.UserAppConfigDirectory()
	
	// Check no double slashes or backslashes
	assert.False(suite.T(), strings.Contains(userDataDir, "//"))
	assert.False(suite.T(), strings.Contains(userDataDir, "\\\\"))
	assert.False(suite.T(), strings.Contains(userConfigDir, "//"))
	assert.False(suite.T(), strings.Contains(userConfigDir, "\\\\"))
	
	// System directories
	systemDataDir := platform.SystemAppDataDirectory()
	systemConfigDir := platform.SystemAppConfigDirectory()
	
	assert.False(suite.T(), strings.Contains(systemDataDir, "//"))
	assert.False(suite.T(), strings.Contains(systemDataDir, "\\\\"))
	assert.False(suite.T(), strings.Contains(systemConfigDir, "//"))
	assert.False(suite.T(), strings.Contains(systemConfigDir, "\\\\"))
}

func (suite *WindowsTestSuite) TestEmptyNamespace() {
	platform := suite.createTestPlatform("Company", "")
	
	// Should still work with empty namespace
	userDataDir := platform.UserAppDataDirectory()
	assert.True(suite.T(), strings.Contains(userDataDir, "Company"))
	assert.True(suite.T(), strings.HasSuffix(userDataDir, "Company")) // namespace is empty, so ends with publisher
}

// Test environment variable override behavior
func (suite *WindowsTestSuite) TestEnvironmentOverride() {
	platform := suite.createTestPlatform("Company", "myapp")
	
	// System directories should use the test environment override
	systemDataDir := platform.SystemAppDataDirectory()
	systemConfigDir := platform.SystemAppConfigDirectory()
	
	assert.True(suite.T(), strings.HasPrefix(systemDataDir, suite.tempDir))
	assert.True(suite.T(), strings.HasPrefix(systemConfigDir, suite.tempDir))
	
	// User directories should use the fake local app data (not affected by test env override)
	userDataDir := platform.UserAppDataDirectory()
	userConfigDir := platform.UserAppConfigDirectory()
	
	assert.True(suite.T(), strings.HasPrefix(userDataDir, suite.fakeLocalAppData))
	assert.True(suite.T(), strings.HasPrefix(userConfigDir, suite.fakeLocalAppData))
}

// Test NewPlatformWindows constructor
func (suite *WindowsTestSuite) TestNewPlatformWindows_Success() {
	platform, err := NewPlatformWindows("Company", "myapp")
	
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), platform)
	
	// Username will be the actual system user since NewPlatformWindows uses user.Current()
	assert.NotEmpty(suite.T(), platform.GetUsername())
	assert.Equal(suite.T(), "Company", platform.servicePublisher)
	assert.Equal(suite.T(), "myapp", platform.serviceNamespace)
	assert.Equal(suite.T(), suite.fakeProgramFiles, platform.programFiles)
	assert.Equal(suite.T(), suite.fakeProgramData, platform.programData)
	assert.Equal(suite.T(), suite.fakeLocalAppData, platform.localAppData)
}

func (suite *WindowsTestSuite) TestNewPlatformWindows_MissingEnvironmentVars() {
	// Save current values
	originalLocalAppData := os.Getenv(envKeyLocalAppData)
	originalProgramData := os.Getenv(envKeyProgramData)
	originalProgramFiles := os.Getenv(envKeyProgramFiles)
	
	// Test missing LOCALAPPDATA
	os.Unsetenv(envKeyLocalAppData)
	platform, err := NewPlatformWindows("Company", "myapp")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), platform)
	assert.Contains(suite.T(), err.Error(), "LOCALAPPDATA")
	
	// Restore LOCALAPPDATA, unset PROGRAMDATA
	os.Setenv(envKeyLocalAppData, originalLocalAppData)
	os.Unsetenv(envKeyProgramData)
	platform, err = NewPlatformWindows("Company", "myapp")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), platform)
	assert.Contains(suite.T(), err.Error(), "PROGRAMDATA")
	
	// Restore PROGRAMDATA, unset PROGRAMFILES
	os.Setenv(envKeyProgramData, originalProgramData)
	os.Unsetenv(envKeyProgramFiles)
	platform, err = NewPlatformWindows("Company", "myapp")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), platform)
	assert.Contains(suite.T(), err.Error(), "PROGRAMFILES")
	
	// Restore all
	os.Setenv(envKeyProgramFiles, originalProgramFiles)
}

// Test Windows-specific path handling
func (suite *WindowsTestSuite) TestWindowsPathSeparators() {
	platform := suite.createTestPlatform("Company", "myapp")
	
	// All paths should work with filepath.Join (cross-platform)
	userDataDir := platform.UserAppDataDirectory()
	systemDataDir := platform.SystemAppDataDirectory()
	
	// Verify paths are properly constructed
	assert.Contains(suite.T(), userDataDir, "Company")
	assert.Contains(suite.T(), userDataDir, "myapp")
	assert.Contains(suite.T(), systemDataDir, "Company")
	assert.Contains(suite.T(), systemDataDir, "myapp")
}