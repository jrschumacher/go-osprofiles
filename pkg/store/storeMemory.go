package store

import (
	"encoding/json"
	"time"
)

type memoryStore struct {
	namespace string
	key       string

	memory   *map[string]interface{}
	metadata *map[string]MemoryMetadata
}

// MemoryMetadata contains metadata for memory store entries (for consistency)
type MemoryMetadata struct {
	ProfileName          string `json:"profile_name"`
	CreatedAt           string `json:"created_at"`
	LastModified        string `json:"last_modified"`
	
	// Versioning fields
	GoOSProfilesVersion  string `json:"go_osprofiles_version"`
	AppVersion          string `json:"app_version,omitempty"`
	ProfileFormatVersion string `json:"profile_format_version"`
}

// NewMemoryStore creates a new in-memory store
// JSON is used to serialize the data to ensure the interface is consistent with other store implementations
var NewMemoryStore NewStoreInterface = func(serviceNamespace, key string, _ ...DriverOpt) (StoreInterface, error) {
	if err := ValidateNamespaceKey(serviceNamespace, key); err != nil {
		return nil, err
	}

	memory := make(map[string]interface{})
	metadata := make(map[string]MemoryMetadata)
	return &memoryStore{
		namespace: serviceNamespace,
		key:       key,
		memory:    &memory,
		metadata:  &metadata,
	}, nil
}

func (k *memoryStore) Exists() bool {
	m := *k.memory
	_, ok := m[k.key]
	return ok
}

func (k *memoryStore) Get() ([]byte, error) {
	m := *k.memory
	v, ok := m[k.key]
	if !ok {
		return nil, nil
	}

	return json.Marshal(v)
}

func (k *memoryStore) Set(value interface{}) error {
	m := *k.memory
	m[k.key] = value
	
	// Update metadata with versioning information
	metadata := *k.metadata
	now := time.Now().Format(time.RFC3339)
	
	goOSProfilesVersion := "1.0.0"  // This should match GoOSProfilesVersionCurrent
	profileFormatVersion := "0.0"   // Memory store is legacy format
	
	if existing, exists := metadata[k.key]; exists {
		// Update existing metadata
		existing.LastModified = now
		metadata[k.key] = existing
	} else {
		// Create new metadata
		metadata[k.key] = MemoryMetadata{
			ProfileName:          k.key,
			CreatedAt:           now,
			LastModified:        now,
			GoOSProfilesVersion:  goOSProfilesVersion,
			AppVersion:          appVersion, // Use global app version
			ProfileFormatVersion: profileFormatVersion,
		}
	}
	
	return nil
}

func (k *memoryStore) Delete() error {
	m := *k.memory
	delete(m, k.key)
	
	// Also delete metadata
	metadata := *k.metadata
	delete(metadata, k.key)
	
	return nil
}

// GetMetadata returns metadata for the stored entry (for consistency with other stores)
func (k *memoryStore) GetMetadata() (*MemoryMetadata, error) {
	metadata := *k.metadata
	if meta, exists := metadata[k.key]; exists {
		return &meta, nil
	}
	return nil, ErrStoredValueInvalid
}
