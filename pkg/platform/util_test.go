package platform

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UtilTestSuite struct {
	suite.Suite
	originalEnvVar      string
	originalSystemRoot  string
	originalDarwinPaths map[string]string
	originalLinuxPaths  map[string]string
	originalWindowsPaths map[string]string
}

func (suite *UtilTestSuite) SetupTest() {
	// Save original values
	suite.originalEnvVar = testBasePathEnvVar
	suite.originalSystemRoot = systemRootPath
	
	// Save Darwin paths
	suite.originalDarwinPaths = map[string]string{
		"library":      darwinLibraryPath,
		"managedPrefs": darwinManagedPrefsPath,
		"appSupport":   darwinAppSupportPath,
	}
	
	// Save Linux paths
	suite.originalLinuxPaths = map[string]string{
		"usrLocal": linuxUsrLocalPath,
		"etc":      linuxEtcPath,
	}
	
	// Save Windows paths
	suite.originalWindowsPaths = map[string]string{
		"programData":  windowsProgramDataPath,
		"programFiles": windowsProgramFilesPath,
	}
}

func (suite *UtilTestSuite) TearDownTest() {
	// Restore original values
	testBasePathEnvVar = suite.originalEnvVar
	systemRootPath = suite.originalSystemRoot
	
	// Restore Darwin paths
	darwinLibraryPath = suite.originalDarwinPaths["library"]
	darwinManagedPrefsPath = suite.originalDarwinPaths["managedPrefs"]
	darwinAppSupportPath = suite.originalDarwinPaths["appSupport"]
	
	// Restore Linux paths
	linuxUsrLocalPath = suite.originalLinuxPaths["usrLocal"]
	linuxEtcPath = suite.originalLinuxPaths["etc"]
	
	// Restore Windows paths
	windowsProgramDataPath = suite.originalWindowsPaths["programData"]
	windowsProgramFilesPath = suite.originalWindowsPaths["programFiles"]
	
	// Clean up environment
	os.Unsetenv("OSPROFILES_TEST_BASE_PATH")
	os.Unsetenv("CUSTOM_TEST_ENV")
}

func (suite *UtilTestSuite) TestOverrideEnvironmentVariable() {
	// Override the environment variable name for testing
	testBasePathEnvVar = "CUSTOM_TEST_ENV"
	
	// Set the custom environment variable
	testPath := "/tmp/custom-test"
	os.Setenv("CUSTOM_TEST_ENV", testPath)
	
	// Should use the custom environment variable
	result := getTestBasePath()
	assert.Equal(suite.T(), testPath, result)
	
	// Should not use the standard environment variable
	os.Setenv("OSPROFILES_TEST_BASE_PATH", "/should/not/be/used")
	result = getTestBasePath()
	assert.Equal(suite.T(), testPath, result) // Should still be custom path
}

func (suite *UtilTestSuite) TestOverrideSystemRootPath() {
	// Override system root path for testing
	systemRootPath = "/test/root"
	
	// Clear environment variable to use system root path
	os.Unsetenv("OSPROFILES_TEST_BASE_PATH")
	
	result := buildSystemPath("usr", "local", "test")
	expected := filepath.Join("/test/root", "usr", "local", "test")
	assert.Equal(suite.T(), expected, result)
}

func (suite *UtilTestSuite) TestOverrideDarwinPaths() {
	// Override Darwin-specific paths for testing
	darwinLibraryPath = "TestLibrary"
	darwinManagedPrefsPath = "TestManagedPrefs"
	darwinAppSupportPath = "TestAppSupport"
	
	// Clear environment variable to use overridden paths
	os.Unsetenv("OSPROFILES_TEST_BASE_PATH")
	
	result := buildDarwinSystemPath(darwinAppSupportPath, "com.test", "app")
	expected := filepath.Join("/", "TestLibrary", "TestAppSupport", "com.test", "app")
	assert.Equal(suite.T(), expected, result)
}

func (suite *UtilTestSuite) TestOverrideLinuxPaths() {
	// Override Linux-specific paths for testing
	linuxUsrLocalPath = "test/usr/local"
	linuxEtcPath = "test/etc"
	
	// Clear environment variable to use overridden paths
	os.Unsetenv("OSPROFILES_TEST_BASE_PATH")
	
	result1 := buildLinuxSystemPath(linuxUsrLocalPath, "com.test", "app")
	expected1 := filepath.Join("/", "test/usr/local", "com.test", "app")
	assert.Equal(suite.T(), expected1, result1)
	
	result2 := buildLinuxSystemPath(linuxEtcPath, "com.test", "app")
	expected2 := filepath.Join("/", "test/etc", "com.test", "app")
	assert.Equal(suite.T(), expected2, result2)
}

func (suite *UtilTestSuite) TestOverrideWindowsPaths() {
	// Override Windows-specific paths for testing
	windowsProgramDataPath = "TestProgramData"
	windowsProgramFilesPath = "TestProgramFiles"
	
	// Clear environment variable to use overridden paths
	os.Unsetenv("OSPROFILES_TEST_BASE_PATH")
	
	result1 := buildWindowsSystemPath("C:\\RealProgramData", windowsProgramDataPath, "com.test", "app")
	expected1 := filepath.Join("C:\\RealProgramData", "com.test", "app")
	assert.Equal(suite.T(), expected1, result1)
	
	result2 := buildWindowsSystemPath("C:\\RealProgramFiles", windowsProgramFilesPath, "com.test", "app")
	expected2 := filepath.Join("C:\\RealProgramFiles", "com.test", "app")
	assert.Equal(suite.T(), expected2, result2)
}

func (suite *UtilTestSuite) TestEnvironmentVariableOverridesPathOverrides() {
	// Set both environment variable and path overrides
	testPath := "/tmp/env-override"
	os.Setenv("OSPROFILES_TEST_BASE_PATH", testPath)
	
	// Override paths (these should still be used even when env var is set)
	systemRootPath = "/should/be/ignored"
	darwinLibraryPath = "ShouldBeIgnored"
	
	result := buildDarwinSystemPath("AppSupport", "test")
	expected := filepath.Join(testPath, "ShouldBeIgnored", "AppSupport", "test") // Should use overridden Library path
	assert.Equal(suite.T(), expected, result)
}

func (suite *UtilTestSuite) TestPathOverridesWorkWithoutEnvironmentVariable() {
	// Clear environment variable
	os.Unsetenv("OSPROFILES_TEST_BASE_PATH")
	
	// Override multiple paths
	systemRootPath = "/custom/root"
	darwinLibraryPath = "CustomLib"
	darwinAppSupportPath = "CustomAppSupport"
	
	result := buildDarwinSystemPath(darwinAppSupportPath, "test", "app")
	expected := filepath.Join("/custom/root", "CustomLib", "CustomAppSupport", "test", "app")
	assert.Equal(suite.T(), expected, result)
}

func (suite *UtilTestSuite) TestMultiplePathOverridesInSingleTest() {
	// Clear environment variable
	os.Unsetenv("OSPROFILES_TEST_BASE_PATH")
	
	// Test scenario 1: Custom library path
	darwinLibraryPath = "Scenario1Lib"
	result1 := buildDarwinSystemPath("AppSupport", "test1")
	expected1 := filepath.Join("/", "Scenario1Lib", "AppSupport", "test1")
	assert.Equal(suite.T(), expected1, result1)
	
	// Test scenario 2: Different library path
	darwinLibraryPath = "Scenario2Lib"
	result2 := buildDarwinSystemPath("AppSupport", "test2")
	expected2 := filepath.Join("/", "Scenario2Lib", "AppSupport", "test2")
	assert.Equal(suite.T(), expected2, result2)
	
	// Verify they are different
	assert.NotEqual(suite.T(), result1, result2)
}

func TestUtilTestSuite(t *testing.T) {
	suite.Run(t, new(UtilTestSuite))
}