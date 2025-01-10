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
		ns          string
		key         string
		expectedErr error
	}{
		{
			"n ame",
			"key",
			ErrValueBadCharacters,
		},
		{
			"name",
			"k ey",
			ErrValueBadCharacters,
		},
		{
			"name!",
			"key",
			ErrValueBadCharacters,
		},
		{
			"name",
			"key&",
			ErrValueBadCharacters,
		},
		{
			"name%",
			"key",
			ErrValueBadCharacters,
		},
		{
			"name",
			"key@",
			ErrValueBadCharacters,
		},
		{
			"",
			"key",
			ErrNamespaceInvalid,
		},
		{
			"name",
			"",
			ErrKeyInvalid,
		},
		{
			strings.Repeat("a", maxFileNameLength+1),
			"key",
			ErrLengthExceeded,
		},
		{
			"name",
			strings.Repeat("b", maxFileNameLength+1),
			ErrLengthExceeded,
		},
	}

	for _, test := range tests {
		err := ValidateNamespaceKey(test.ns, test.key)
		if !errors.Is(err, test.expectedErr) {
			t.Errorf("expected error %v, got %v", test.expectedErr, err)
		}
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
	foundSingleProfileFile := false
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
		split := strings.Split(file.Name(), ":")
		assert.Equal(t, testNS, split[2])

		last := split[len(split)-1]
		if strings.Contains(last, testKey+".nfo") {
			foundGlobalConfigFile = true
		}
		if strings.Contains(last, testKey+".enc") {
			foundSingleProfileFile = true
		}
	}

	assert.True(t, foundGlobalConfigFile)
	assert.True(t, foundSingleProfileFile)

	data, err := store.Get()
	require.NoError(t, err)
	require.NotNil(t, data)

	var storedValue *mockStoredValue
	err = json.Unmarshal(data, &storedValue)
	require.NoError(t, err)

	assert.Equal(t, value.Name, storedValue.Name)
	assert.Equal(t, value.TestValue, storedValue.TestValue)
}
