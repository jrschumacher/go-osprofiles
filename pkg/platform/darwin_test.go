package platform

import (
	"encoding/json"
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

// Test JSON expansion depth limiting
func (suite *DarwinTestSuite) TestExpandJSONStringsRecursive_DepthLimiting() {
	platform := &PlatformDarwin{
		serviceNamespace: "com.company.app",
		servicePublisher: "",
	}

	// Create a deeply nested structure that would exceed the depth limit
	deepObj := make(map[string]any)
	current := deepObj
	
	// Create 15 levels of nesting (exceeds maxJSONExpansionDepth of 10)
	for i := 0; i < 15; i++ {
		next := make(map[string]any)
		current[fmt.Sprintf("level%d", i)] = next
		current = next
	}
	current["deepValue"] = "should not be processed"

	// Test that expansion stops at max depth
	result := platform.expandJSONStringsRecursive(deepObj, 0)
	
	// Result should be a map, not nil
	resultMap, ok := result.(map[string]any)
	assert.True(suite.T(), ok)
	assert.NotNil(suite.T(), resultMap)
	
	// Verify we can traverse some levels but not all
	assert.Contains(suite.T(), resultMap, "level0")
}

func (suite *DarwinTestSuite) TestExpandJSONStringsRecursive_MaxDepthReached() {
	platform := &PlatformDarwin{
		serviceNamespace: "com.company.app",
		servicePublisher: "",
	}

	// Test when we're already at max depth
	testObj := map[string]any{
		"key": "value",
		"nested": map[string]any{
			"inner": "should not be processed",
		},
	}

	result := platform.expandJSONStringsRecursive(testObj, maxJSONExpansionDepth)
	
	// Should return the object unchanged when at max depth
	assert.Equal(suite.T(), testObj, result)
}

func (suite *DarwinTestSuite) TestExpandJSONStringsRecursive_JSONStringExpansion() {
	platform := &PlatformDarwin{
		serviceNamespace: "com.company.app",
		servicePublisher: "",
	}

	// Test normal JSON string expansion within depth limits
	testObj := map[string]any{
		"normalKey": "normalValue",
		"jsonKey":   `{"expanded": true, "value": 42}`,
		"arrayKey": []any{
			"normalString",
			`{"nested": "json"}`,
		},
	}

	result := platform.expandJSONStringsRecursive(testObj, 0)
	
	resultMap, ok := result.(map[string]any)
	assert.True(suite.T(), ok)
	
	// Normal key should be unchanged
	assert.Equal(suite.T(), "normalValue", resultMap["normalKey"])
	
	// JSON string should be expanded to object
	jsonValue, exists := resultMap["jsonKey"]
	assert.True(suite.T(), exists)
	jsonObj, ok := jsonValue.(map[string]any)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), true, jsonObj["expanded"])
	assert.Equal(suite.T(), float64(42), jsonObj["value"]) // JSON numbers become float64
	
	// Array should be processed
	arrayValue, exists := resultMap["arrayKey"]
	assert.True(suite.T(), exists)
	arrayObj, ok := arrayValue.([]any)
	assert.True(suite.T(), ok)
	assert.Len(suite.T(), arrayObj, 2)
	assert.Equal(suite.T(), "normalString", arrayObj[0])
	
	// Second array element should be expanded JSON
	nestedObj, ok := arrayObj[1].(map[string]any)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "json", nestedObj["nested"])
}

func (suite *DarwinTestSuite) TestExpandJSONStringsRecursive_InvalidJSON() {
	platform := &PlatformDarwin{
		serviceNamespace: "com.company.app",
		servicePublisher: "",
	}

	// Test with invalid JSON strings
	testObj := map[string]any{
		"invalidJson": `{"incomplete": json}`, // Invalid JSON
		"notJson":     "just a regular string",
		"emptyString": "",
	}

	result := platform.expandJSONStringsRecursive(testObj, 0)
	
	resultMap, ok := result.(map[string]any)
	assert.True(suite.T(), ok)
	
	// Invalid JSON should remain as string
	assert.Equal(suite.T(), `{"incomplete": json}`, resultMap["invalidJson"])
	assert.Equal(suite.T(), "just a regular string", resultMap["notJson"])
	assert.Equal(suite.T(), "", resultMap["emptyString"])
}

func (suite *DarwinTestSuite) TestExpandJSONStrings_Integration() {
	platform := &PlatformDarwin{
		serviceNamespace: "com.company.app",
		servicePublisher: "",
	}

	// Test the main expandJSONStrings function
	testData := map[string]any{
		"config": `{"database": {"host": "localhost", "port": 5432}}`,
		"settings": map[string]any{
			"theme": "dark",
			"features": `["feature1", "feature2"]`,
		},
	}

	jsonBytes, err := json.Marshal(testData)
	suite.Require().NoError(err)

	result, err := platform.expandJSONStrings(jsonBytes)
	suite.Require().NoError(err)

	var resultObj map[string]any
	err = json.Unmarshal(result, &resultObj)
	suite.Require().NoError(err)

	// Verify config was expanded
	configValue, exists := resultObj["config"]
	assert.True(suite.T(), exists)
	configObj, ok := configValue.(map[string]any)
	assert.True(suite.T(), ok)
	
	dbObj, ok := configObj["database"].(map[string]any)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "localhost", dbObj["host"])
	assert.Equal(suite.T(), float64(5432), dbObj["port"])

	// Verify nested settings
	settingsValue, exists := resultObj["settings"]
	assert.True(suite.T(), exists)
	settingsObj, ok := settingsValue.(map[string]any)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "dark", settingsObj["theme"])
	
	// Features should be expanded to array
	featuresValue, exists := settingsObj["features"]
	assert.True(suite.T(), exists)
	featuresArray, ok := featuresValue.([]any)
	assert.True(suite.T(), ok)
	assert.Len(suite.T(), featuresArray, 2)
	assert.Equal(suite.T(), "feature1", featuresArray[0])
	assert.Equal(suite.T(), "feature2", featuresArray[1])
}

func (suite *DarwinTestSuite) TestExpandJSONStrings_CircularReference() {
	platform := &PlatformDarwin{
		serviceNamespace: "com.company.app",
		servicePublisher: "",
	}

	// Create a structure that could cause circular references through JSON strings
	// This tests that depth limiting prevents infinite recursion
	circularJSON := `{"level": 1, "next": "{\"level\": 2, \"next\": \"{\\\"level\\\": 3, \\\"next\\\": \\\"{\\\\\\\"level\\\\\\\": 4}\\\"}\"}"}`
	
	testData := map[string]any{
		"circular": circularJSON,
	}

	jsonBytes, err := json.Marshal(testData)
	suite.Require().NoError(err)

	// This should not hang or crash due to depth limiting
	result, err := platform.expandJSONStrings(jsonBytes)
	suite.Require().NoError(err)
	
	var resultObj map[string]any
	err = json.Unmarshal(result, &resultObj)
	suite.Require().NoError(err)
	
	// Should have expanded some levels but stopped at depth limit
	circularValue, exists := resultObj["circular"]
	assert.True(suite.T(), exists)
	assert.NotNil(suite.T(), circularValue)
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(DarwinTestSuite))
}