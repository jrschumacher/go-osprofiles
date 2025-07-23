package platform

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type DarwinTestSuite struct {
	suite.Suite
	tempDir     string
	originalEnv string
}

func (suite *DarwinTestSuite) SetupSuite() {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "darwin-platform-test-*")
	suite.Require().NoError(err)
	suite.tempDir = tempDir

	// Store original environment variable
	suite.originalEnv = os.Getenv("OSPROFILES_TEST_BASE_PATH")

	// Set environment variable to use temp directory
	os.Setenv("OSPROFILES_TEST_BASE_PATH", tempDir)

	// Create test directory structure
	err = os.MkdirAll(filepath.Join(tempDir, "Library", "Application Support"), 0755)
	suite.Require().NoError(err)
	err = os.MkdirAll(filepath.Join(tempDir, "Library", "Managed Preferences"), 0755)
	suite.Require().NoError(err)
}

func (suite *DarwinTestSuite) TearDownSuite() {
	// Restore original environment variable
	if suite.originalEnv != "" {
		os.Setenv("OSPROFILES_TEST_BASE_PATH", suite.originalEnv)
	} else {
		os.Unsetenv("OSPROFILES_TEST_BASE_PATH")
	}

	// Clean up temp directory
	os.RemoveAll(suite.tempDir)
}

func (suite *DarwinTestSuite) copyTestPlist(filename, destination string) {
	testdataPath := filepath.Join("testdata", filename)
	content, err := os.ReadFile(testdataPath)
	suite.Require().NoError(err)
	
	err = os.WriteFile(destination, content, 0644)
	suite.Require().NoError(err)
}

func (suite *DarwinTestSuite) TestMDMConfigPath_RDNS_Namespace() {
	platform := &PlatformDarwin{
		serviceNamespace: "com.company.app",
		servicePublisher: "",
	}

	expectedPath := filepath.Join(suite.tempDir, "Library", "Managed Preferences", "com.company.app.plist")
	actualPath := platform.MDMConfigPath()
	
	assert.Equal(suite.T(), expectedPath, actualPath)
}

func (suite *DarwinTestSuite) TestMDMConfigPath_RDNS_Publisher_With_Namespace() {
	platform := &PlatformDarwin{
		serviceNamespace: "myapp",
		servicePublisher: "com.company",
	}

	expectedPath := filepath.Join(suite.tempDir, "Library", "Managed Preferences", "com.company.myapp.plist")
	actualPath := platform.MDMConfigPath()
	
	assert.Equal(suite.T(), expectedPath, actualPath)
}

func (suite *DarwinTestSuite) TestMDMConfigPath_NonRDNS_Publisher_With_Namespace() {
	platform := &PlatformDarwin{
		serviceNamespace: "testapp",
		servicePublisher: "testpublisher",
	}

	expectedPath := filepath.Join(suite.tempDir, "Library", "Managed Preferences", "testpublisher.testapp.plist")
	actualPath := platform.MDMConfigPath()
	
	assert.Equal(suite.T(), expectedPath, actualPath)
}

func (suite *DarwinTestSuite) TestMDMConfigPath_Namespace_Only() {
	platform := &PlatformDarwin{
		serviceNamespace: "simple-app",
		servicePublisher: "",
	}

	expectedPath := filepath.Join(suite.tempDir, "Library", "Managed Preferences", "simple-app.plist")
	actualPath := platform.MDMConfigPath()
	
	assert.Equal(suite.T(), expectedPath, actualPath)
}

func (suite *DarwinTestSuite) TestMDMConfigExists_FileExists_Readable() {
	platform := &PlatformDarwin{
		serviceNamespace: "com.company.app",
		servicePublisher: "",
	}

	// Copy test plist to expected location
	plistPath := filepath.Join(suite.tempDir, "Library", "Managed Preferences", "com.company.app.plist")
	suite.copyTestPlist("com.company.app.plist", plistPath)

	exists := platform.MDMConfigExists()
	assert.True(suite.T(), exists)
}

func (suite *DarwinTestSuite) TestMDMConfigExists_FileNotExists() {
	platform := &PlatformDarwin{
		serviceNamespace: "non.existent.app",
		servicePublisher: "",
	}

	exists := platform.MDMConfigExists()
	assert.False(suite.T(), exists)
}

func (suite *DarwinTestSuite) TestMDMConfigExists_FileExists_NotReadable() {
	platform := &PlatformDarwin{
		serviceNamespace: "unreadable.app",
		servicePublisher: "",
	}

	// Create file but make it unreadable
	plistPath := filepath.Join(suite.tempDir, "Library", "Managed Preferences", "unreadable.app.plist")
	suite.copyTestPlist("simple-app.plist", plistPath)
	
	// Make file unreadable
	err := os.Chmod(plistPath, 0000)
	suite.Require().NoError(err)
	defer os.Chmod(plistPath, 0644) // cleanup

	exists := platform.MDMConfigExists()
	assert.False(suite.T(), exists)
}

func (suite *DarwinTestSuite) TestSystemAppDataDirectoryWithMDM_ExplicitRDNS() {
	platform := &PlatformDarwin{
		serviceNamespace: "myapp",
		servicePublisher: "testpublisher",
		userHomeDir:      "/Users/testuser",
	}

	systemDir, opts := platform.SystemAppDataDirectoryWithMDM("com.explicit.rdns")
	
	expectedDir := filepath.Join(suite.tempDir, "Library", "Application Support", "testpublisher", "myapp")
	assert.Equal(suite.T(), expectedDir, systemDir)
	
	// Check that we have the correct number of options (store directory + MDM)
	require.Len(suite.T(), opts, 2)
}

func (suite *DarwinTestSuite) TestSystemAppDataDirectoryWithMDM_RDNS_Namespace() {
	platform := &PlatformDarwin{
		serviceNamespace: "com.company.app",
		servicePublisher: "",
		userHomeDir:      "/Users/testuser",
	}

	systemDir, opts := platform.SystemAppDataDirectoryWithMDM()
	
	expectedDir := filepath.Join(suite.tempDir, "Library", "Application Support", "com.company.app")
	assert.Equal(suite.T(), expectedDir, systemDir)
	
	// Check that we have the correct number of options (store directory + MDM)
	require.Len(suite.T(), opts, 2)
}

func (suite *DarwinTestSuite) TestSystemAppDataDirectoryWithMDM_RDNS_Publisher() {
	platform := &PlatformDarwin{
		serviceNamespace: "myapp",
		servicePublisher: "com.company",
		userHomeDir:      "/Users/testuser",
	}

	systemDir, opts := platform.SystemAppDataDirectoryWithMDM()
	
	expectedDir := filepath.Join(suite.tempDir, "Library", "Application Support", "com.company", "myapp")
	assert.Equal(suite.T(), expectedDir, systemDir)
	
	// Check that we have the correct number of options (store directory + MDM)
	require.Len(suite.T(), opts, 2)
}

func (suite *DarwinTestSuite) TestSystemAppDataDirectoryWithMDM_NonRDNS() {
	platform := &PlatformDarwin{
		serviceNamespace: "testapp",
		servicePublisher: "testpublisher",
		userHomeDir:      "/Users/testuser",
	}

	systemDir, opts := platform.SystemAppDataDirectoryWithMDM()
	
	expectedDir := filepath.Join(suite.tempDir, "Library", "Application Support", "testpublisher", "testapp")
	assert.Equal(suite.T(), expectedDir, systemDir)
	
	// Check that we have the correct number of options (store directory + MDM)
	require.Len(suite.T(), opts, 2)
}

func (suite *DarwinTestSuite) TestSystemAppDataDirectoryWithAutoMDM_RDNS_EnablesMDM() {
	platform := &PlatformDarwin{
		serviceNamespace: "com.company.app",
		servicePublisher: "",
		userHomeDir:      "/Users/testuser",
	}

	systemDir, opts := platform.SystemAppDataDirectoryWithAutoMDM()
	
	expectedDir := filepath.Join(suite.tempDir, "Library", "Application Support", "com.company.app")
	assert.Equal(suite.T(), expectedDir, systemDir)
	
	// Should have MDM support enabled (store directory + MDM)
	require.Len(suite.T(), opts, 2)
}

func (suite *DarwinTestSuite) TestSystemAppDataDirectoryWithAutoMDM_NonRDNS_NoMDM() {
	platform := &PlatformDarwin{
		serviceNamespace: "simpleapp",
		servicePublisher: "",
		userHomeDir:      "/Users/testuser",
	}

	systemDir, opts := platform.SystemAppDataDirectoryWithAutoMDM()
	
	expectedDir := filepath.Join(suite.tempDir, "Library", "Application Support", "simpleapp")
	assert.Equal(suite.T(), expectedDir, systemDir)
	
	// Should only have store directory option, no MDM
	require.Len(suite.T(), opts, 1)
}

func (suite *DarwinTestSuite) TestNewPlatformDarwin_Success() {
	platform, err := NewPlatformDarwin("testpublisher", "testapp")
	
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), platform)
	
	assert.Equal(suite.T(), "testpublisher", platform.servicePublisher)
	assert.Equal(suite.T(), "testapp", platform.serviceNamespace)
	assert.NotEmpty(suite.T(), platform.username)
	assert.NotEmpty(suite.T(), platform.userHomeDir)
}

func (suite *DarwinTestSuite) TestGetUsername() {
	platform := &PlatformDarwin{
		username: "testuser",
	}
	
	assert.Equal(suite.T(), "testuser", platform.GetUsername())
}

func (suite *DarwinTestSuite) TestUserHomeDir() {
	platform := &PlatformDarwin{
		userHomeDir: "/Users/testuser",
	}
	
	assert.Equal(suite.T(), "/Users/testuser", platform.UserHomeDir())
}

func (suite *DarwinTestSuite) TestUserAppDataDirectory_NoPublisher() {
	platform := &PlatformDarwin{
		serviceNamespace: "testapp",
		servicePublisher: "",
		userHomeDir:      "/Users/testuser",
	}

	actual := platform.UserAppDataDirectory()
	
	// For user directories, we still use the real "Library" path, not the mocked one
	assert.True(suite.T(), strings.HasSuffix(actual, "Library/Application Support/testapp"))
	assert.True(suite.T(), strings.HasPrefix(actual, "/Users/testuser"))
}

func (suite *DarwinTestSuite) TestUserAppDataDirectory_WithPublisher() {
	platform := &PlatformDarwin{
		serviceNamespace: "testapp",
		servicePublisher: "testpublisher",
		userHomeDir:      "/Users/testuser",
	}

	actual := platform.UserAppDataDirectory()
	
	// For user directories, we still use the real "Library" path, not the mocked one
	assert.True(suite.T(), strings.HasSuffix(actual, "Library/Application Support/testpublisher/testapp"))
	assert.True(suite.T(), strings.HasPrefix(actual, "/Users/testuser"))
}

func (suite *DarwinTestSuite) TestUserAppConfigDirectory_NoPublisher() {
	platform := &PlatformDarwin{
		serviceNamespace: "testapp",
		servicePublisher: "",
		userHomeDir:      "/Users/testuser",
	}

	actual := platform.UserAppConfigDirectory()
	
	// For user directories, we still use the real "Library" path, not the mocked one
	assert.True(suite.T(), strings.HasSuffix(actual, "Library/Application Support/testapp"))
	assert.True(suite.T(), strings.HasPrefix(actual, "/Users/testuser"))
}

func (suite *DarwinTestSuite) TestUserAppConfigDirectory_WithPublisher() {
	platform := &PlatformDarwin{
		serviceNamespace: "testapp",
		servicePublisher: "testpublisher",
		userHomeDir:      "/Users/testuser",
	}

	actual := platform.UserAppConfigDirectory()
	
	// For user directories, we still use the real "Library" path, not the mocked one
	assert.True(suite.T(), strings.HasSuffix(actual, "Library/Application Support/testpublisher/testapp"))
	assert.True(suite.T(), strings.HasPrefix(actual, "/Users/testuser"))
}

func (suite *DarwinTestSuite) TestSystemAppDataDirectory_NoPublisher() {
	platform := &PlatformDarwin{
		serviceNamespace: "testapp",
		servicePublisher: "",
	}

	expected := filepath.Join(suite.tempDir, "Library", "Application Support", "testapp")
	actual := platform.SystemAppDataDirectory()
	
	assert.Equal(suite.T(), expected, actual)
}

func (suite *DarwinTestSuite) TestSystemAppDataDirectory_WithPublisher() {
	platform := &PlatformDarwin{
		serviceNamespace: "testapp",
		servicePublisher: "testpublisher",
	}

	expected := filepath.Join(suite.tempDir, "Library", "Application Support", "testpublisher", "testapp")
	actual := platform.SystemAppDataDirectory()
	
	assert.Equal(suite.T(), expected, actual)
}

func (suite *DarwinTestSuite) TestSystemAppConfigDirectory_NoPublisher() {
	platform := &PlatformDarwin{
		serviceNamespace: "testapp",
		servicePublisher: "",
	}

	expected := filepath.Join(suite.tempDir, "Library", "Application Support", "testapp")
	actual := platform.SystemAppConfigDirectory()
	
	assert.Equal(suite.T(), expected, actual)
}

func (suite *DarwinTestSuite) TestSystemAppConfigDirectory_WithPublisher() {
	platform := &PlatformDarwin{
		serviceNamespace: "testapp",
		servicePublisher: "testpublisher",
	}

	expected := filepath.Join(suite.tempDir, "Library", "Application Support", "testpublisher", "testapp")
	actual := platform.SystemAppConfigDirectory()
	
	assert.Equal(suite.T(), expected, actual)
}

// Integration tests that use real file system paths (existing test patterns)
func TestPlatformDarwin_Integration_NoPublisher(t *testing.T) {
	fakeAppName := "test-darwin-mdm-app"
	var publisher string
	platform, err := NewPlatformDarwin(publisher, fakeAppName)

	require.NoError(t, err)
	require.NotNil(t, platform)

	// Test basic functionality
	assert.NotEmpty(t, platform.GetUsername())
	assert.NotEmpty(t, platform.UserHomeDir())

	// Test user directories
	userConfigDir := platform.UserAppConfigDirectory()
	assert.True(t, strings.HasSuffix(userConfigDir, fakeAppName))
	assert.Contains(t, userConfigDir, "Library/Application Support")
	assert.False(t, strings.HasPrefix(userConfigDir, "/Library/Application Support"))

	userDataDir := platform.UserAppDataDirectory()
	assert.True(t, strings.HasSuffix(userDataDir, fakeAppName))
	assert.Contains(t, userDataDir, "Library/Application Support")

	// Test system directories
	sysConfigDir := platform.SystemAppConfigDirectory()
	assert.True(t, strings.HasSuffix(sysConfigDir, fakeAppName))
	assert.Equal(t, fmt.Sprintf("/Library/Application Support/%s", fakeAppName), sysConfigDir)

	sysDataDir := platform.SystemAppDataDirectory()
	assert.True(t, strings.HasSuffix(sysDataDir, fakeAppName))
	assert.Equal(t, fmt.Sprintf("/Library/Application Support/%s", fakeAppName), sysDataDir)

	// Test MDM path generation
	mdmPath := platform.MDMConfigPath()
	assert.True(t, strings.HasSuffix(mdmPath, fmt.Sprintf("%s.plist", fakeAppName)))
	assert.Contains(t, mdmPath, "Managed Preferences")
}

func TestPlatformDarwin_Integration_WithPublisher(t *testing.T) {
	fakeAppName := "test-darwin-mdm-publisher-app"
	fakePublisher := "test-publisher"
	platform, err := NewPlatformDarwin(fakePublisher, fakeAppName)

	require.NoError(t, err)
	require.NotNil(t, platform)

	// Test system directories with publisher
	sysConfigDir := platform.SystemAppConfigDirectory()
	assert.True(t, strings.HasSuffix(sysConfigDir, fakeAppName))
	assert.Equal(t, fmt.Sprintf("/Library/Application Support/%s/%s", fakePublisher, fakeAppName), sysConfigDir)

	// Test MDM functionality
	mdmPath := platform.MDMConfigPath()
	expectedMDMName := fmt.Sprintf("%s.%s.plist", fakePublisher, fakeAppName)
	assert.True(t, strings.HasSuffix(mdmPath, expectedMDMName))

	// Test auto MDM detection
	systemDir, opts := platform.SystemAppDataDirectoryWithAutoMDM()
	assert.Equal(t, sysConfigDir, systemDir)
	assert.Len(t, opts, 1) // No MDM for non-RDNS namespace
}

func TestPlatformDarwin_Integration_RDNS_Namespace(t *testing.T) {
	rdnsNamespace := "com.test.rdns.app"
	platform, err := NewPlatformDarwin("", rdnsNamespace)

	require.NoError(t, err)
	require.NotNil(t, platform)

	// Test MDM path with RDNS namespace
	mdmPath := platform.MDMConfigPath()
	assert.True(t, strings.HasSuffix(mdmPath, fmt.Sprintf("%s.plist", rdnsNamespace)))

	// Test auto MDM detection should enable MDM for RDNS namespace
	systemDir, opts := platform.SystemAppDataDirectoryWithAutoMDM()
	expectedDir := fmt.Sprintf("/Library/Application Support/%s", rdnsNamespace)
	assert.Equal(t, expectedDir, systemDir)
	assert.Len(t, opts, 2) // Should have both store directory and MDM support
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(DarwinTestSuite))
}