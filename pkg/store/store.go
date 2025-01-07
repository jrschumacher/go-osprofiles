package store

import (
	"errors"
	"fmt"
	"regexp"

)

// DriverOpt is a variadic function to set driver options.
type DriverOpt func() error

// StoreInterface is an interface for a store of a single key and value under a namespace.
type NewStoreInterface func(namespace, key string, driverOpt ...DriverOpt) (StoreInterface, error)

// TODO: should we reconfigure this abstraction so we have a more traditional key-value store?

// StoreInterface is a reusable interface that varied drivers can share to implement a store.
type StoreInterface interface {
	// Exists returns true if the value exists in the store.
	Exists() bool
	// Get retrieves the entry from the store and unmarshals it into the provided value.
	Get(value interface{}) error
	// Set marshals the provided value into JSON and stores it.
	Set(value interface{}) error
	// Delete removes the entry from the store.
	Delete() error
	
}

const maxFileNameLength = 255

// ValidateNamespaceKey ensures the namespace and key are valid and within length bounds.
func ValidateNamespaceKey(namespace, key string) error {
	// Regular expression for allowed characters (alphanumerics, underscore, hyphen)
	validName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

	if len(namespace) == 0 {
		return errors.Join(ErrNamespaceInvalid, ErrValueEmpty)
	}

	if len(key) == 0 {
		return errors.Join(ErrKeyInvalid, ErrValueEmpty)
	}

	if !validName.MatchString(namespace) {
		return fmt.Errorf("%w, %w, namespace: %s", ErrNamespaceInvalid, ErrValueBadCharacters, namespace)
	}
	if !validName.MatchString(key) {
		return fmt.Errorf("%w, %w, key: %s", ErrKeyInvalid, ErrValueBadCharacters, key)
	}

	// Ensure the filename is within length bounds when including a file extension
	filename := fmt.Sprintf("%s_%s.ext", namespace, key)
	if len(filename) > maxFileNameLength {
		return fmt.Errorf("%w, <namespace_key>.ext exceeds maximum length (%d): %s", ErrLengthExceeded, maxFileNameLength, filename)
	}

	return nil
}