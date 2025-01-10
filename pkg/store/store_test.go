package store

import (
	"encoding/json"
	"errors"
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
		Name:      "test",
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
