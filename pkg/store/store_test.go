package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ValidateNamespaceKey(t *testing.T) {
	tests := []struct {
		name        string
		ns          string
		key         string
		expectedErr error
	}{
		// Valid cases
		{
			name: "valid simple namespace and key",
			ns:   "myapp",
			key:  "profile1",
		},
		{
			name: "valid namespace and key with underscores and hyphens",
			ns:   "my-app_v2",
			key:  "user_profile-1",
		},
		{
			name: "valid RDNS namespace",
			ns:   "com.example.app",
			key:  "profile",
		},
		{
			name: "valid RDNS key",
			ns:   "myapp",
			key:  "com.example.profile",
		},
		{
			name: "valid both RDNS namespace and key",
			ns:   "com.company.app",
			key:  "com.user.profile",
		},
		{
			name: "valid complex RDNS",
			ns:   "org.example.sub-domain.app-name",
			key:  "user.profile.config-v2",
		},
		{
			name: "valid single character",
			ns:   "a",
			key:  "b",
		},

		// Invalid characters in namespace
		{
			name:        "namespace with space",
			ns:          "n ame",
			key:         "key",
			expectedErr: ErrValueBadCharacters,
		},
		{
			name:        "namespace with exclamation",
			ns:          "name!",
			key:         "key",
			expectedErr: ErrValueBadCharacters,
		},
		{
			name:        "namespace with percent",
			ns:          "name%",
			key:         "key",
			expectedErr: ErrValueBadCharacters,
		},
		{
			name:        "namespace with colon",
			ns:          "name:",
			key:         "key",
			expectedErr: ErrValueBadCharacters,
		},
		{
			name:        "namespace with pipe",
			ns:          "name|",
			key:         "key",
			expectedErr: ErrValueBadCharacters,
		},
		{
			name:        "namespace with forward slash",
			ns:          "name/",
			key:         "key",
			expectedErr: ErrValueBadCharacters,
		},
		{
			name:        "namespace with backslash",
			ns:          "name\\",
			key:         "key",
			expectedErr: ErrValueBadCharacters,
		},
		{
			name:        "namespace starting with period",
			ns:          ".hidden",
			key:         "key",
			expectedErr: ErrValueBadCharacters,
		},
		{
			name:        "namespace ending with period",
			ns:          "name.",
			key:         "key",
			expectedErr: ErrValueBadCharacters,
		},

		// Invalid characters in key
		{
			name:        "key with space",
			ns:          "name",
			key:         "k ey",
			expectedErr: ErrValueBadCharacters,
		},
		{
			name:        "key with ampersand",
			ns:          "name",
			key:         "key&",
			expectedErr: ErrValueBadCharacters,
		},
		{
			name:        "key with at symbol",
			ns:          "name",
			key:         "key@",
			expectedErr: ErrValueBadCharacters,
		},
		{
			name:        "key starting with period",
			ns:          "name",
			key:         ".hidden",
			expectedErr: ErrValueBadCharacters,
		},
		{
			name:        "key ending with period",
			ns:          "name",
			key:         "key.",
			expectedErr: ErrValueBadCharacters,
		},

		// Empty values
		{
			name:        "empty namespace",
			ns:          "",
			key:         "key",
			expectedErr: ErrNamespaceInvalid,
		},
		{
			name:        "empty key",
			ns:          "name",
			key:         "",
			expectedErr: ErrKeyInvalid,
		},

		// Length exceeded
		{
			name:        "namespace too long",
			ns:          strings.Repeat("a", maxFileNameLength+1),
			key:         "key",
			expectedErr: ErrLengthExceeded,
		},
		{
			name:        "key too long",
			ns:          "name",
			key:         strings.Repeat("b", maxFileNameLength+1),
			expectedErr: ErrLengthExceeded,
		},
		{
			name:        "combined namespace and key too long",
			ns:          strings.Repeat("a", 200),
			key:         strings.Repeat("b", 200),
			expectedErr: ErrLengthExceeded,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateNamespaceKey(test.ns, test.key)
			if test.expectedErr == nil {
				assert.NoError(t, err, "expected no error for valid input")
			} else {
				assert.Error(t, err, "expected error for invalid input")
				assert.True(t, errors.Is(err, test.expectedErr), 
					"expected error %v, got %v", test.expectedErr, err)
			}
		})
	}
}

type mockStoredValue struct {
	Name      string `json:"name"`
	TestValue string `json:"test_value"`
}

func Test_NewMemoryStore(t *testing.T) {
	testNS := "test_namespace"
	testKey := "profile"

	store, err := NewMemoryStore(testNS, testKey)
	require.NoError(t, err)
	require.NotNil(t, store)

	require.False(t, store.Exists())

	value := mockStoredValue{
		Name:      "test_memory",
		TestValue: "special_test_value",
	}
	err = store.Set(value)
	require.NoError(t, err)
	require.True(t, store.Exists())

	data, err := store.Get()
	require.NoError(t, err)
	require.NotNil(t, data)

	var storedValue *mockStoredValue
	err = json.Unmarshal(data, &storedValue)
	require.NoError(t, err)

	assert.Equal(t, value.Name, storedValue.Name)
	assert.Equal(t, value.TestValue, storedValue.TestValue)
}

func Test_NewFileSystemStore_DirectoryProvided(t *testing.T) {
	testNS := "test_namespace"
	testKey := "profile"

	dir := t.TempDir()
	store, err := NewFileStore(testNS, testKey, WithStoreDirectory(dir))
	require.NoError(t, err)
	require.NotNil(t, store)

	require.False(t, store.Exists())

	value := mockStoredValue{
		Name:      "fs_store_test",
		TestValue: "file_system_stored",
	}
	err = store.Set(value)
	require.NoError(t, err)
	require.True(t, store.Exists())

	// ensure two files were written to the temp dir
	files, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Len(t, files, 2)

	// check the written files
	foundGlobalConfigFile := false
	foundGlobalConfigEncFile := false
	for _, file := range files {
		path := filepath.Join(dir, file.Name())
		data, err := os.ReadFile(path)
		require.NoError(t, err)
		require.NotNil(t, data)

		fileContents := string(data)

		// ensure data is not human readable
		isEncrypted := !strings.Contains(fileContents, value.Name)
		assert.True(t, isEncrypted)
		isEncrypted = !strings.Contains(fileContents, value.TestValue)
		assert.True(t, isEncrypted)

		// assert file name matches URN
		split := strings.Split(file.Name(), ".")
		assert.Equal(t, testNS, split[2])

		last := file.Name()[len(testKey):]
		if strings.Contains(last, testKey+".nfo") {
			foundGlobalConfigFile = true
		}
		if strings.Contains(last, testKey+".enc") {
			foundGlobalConfigEncFile = true
		}
	}

	assert.True(t, foundGlobalConfigFile)
	assert.True(t, foundGlobalConfigEncFile)

	data, err := store.Get()
	require.NoError(t, err)
	require.NotNil(t, data)

	var storedValue *mockStoredValue
	err = json.Unmarshal(data, &storedValue)
	require.NoError(t, err)

	assert.Equal(t, value.Name, storedValue.Name)
	assert.Equal(t, value.TestValue, storedValue.TestValue)
}

func Test_NewKeyringStore(t *testing.T) {
	testNS := "test_namespace"
	testKey := "profile"

	// dir that should be ignored by the keyring store
	dir := t.TempDir()

	store, err := NewKeyringStore(testNS, testKey, WithStoreDirectory(dir))
	require.NoError(t, err)
	require.NotNil(t, store)

	// dir should still be empty after store init
	files, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Zero(t, len(files))

	value := mockStoredValue{
		Name:      "test_keyring",
		TestValue: "keyring_value",
	}
	err = store.Set(value)
	require.NoError(t, err)
	require.True(t, store.Exists())

	// ensure exactly zero files were written to the store driver directory
	files, err = os.ReadDir(dir)
	require.NoError(t, err)
	require.Zero(t, len(files))

	data, err := store.Get()
	require.NoError(t, err)
	require.NotNil(t, data)

	var storedValue *mockStoredValue
	err = json.Unmarshal(data, &storedValue)
	require.NoError(t, err)

	assert.Equal(t, value.Name, storedValue.Name)
	assert.Equal(t, value.TestValue, storedValue.TestValue)
}
