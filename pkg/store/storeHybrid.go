package store

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/zalando/go-keyring"
)

// HybridStore implements StoreInterface with security-aware storage
type HybridStore struct {
	config              *SecurityConfig
	secureFilePath      string
	plaintextFilePath   string
	namespaceVersionURN string
	
	// Temporary data is kept in memory only
	temporaryData map[string]interface{}
	
	// Profile type for reconstruction
	profileType reflect.Type
}

// HybridMetadata contains metadata for hybrid storage
type HybridMetadata struct {
	ProfileName          string `json:"profile_name"`
	CreatedAt           string `json:"created_at"`
	LastModified        string `json:"last_modified"`
	SecurityMode        string `json:"security_mode"`
	HasSecureData       bool   `json:"has_secure_data"`
	HasPlaintextData    bool   `json:"has_plaintext_data"`
	EncryptionAlg       string `json:"encryption_alg,omitempty"`
	Version             string `json:"version"` // Legacy field for compatibility
	
	// New versioning fields
	GoOSProfilesVersion string `json:"go_osprofiles_version"`
	AppVersion          string `json:"app_version,omitempty"`
	ProfileFormatVersion string `json:"profile_format_version"`
}

// NewHybridStore creates a new hybrid security store
var NewHybridStore NewStoreInterface = func(serviceNamespace, key string, driverOpts ...DriverOpt) (StoreInterface, error) {
	if err := ValidateNamespaceKey(serviceNamespace, key); err != nil {
		return nil, err
	}

	// Default security config
	config := &SecurityConfig{
		Mode:             SecurityModeKeyring,
		WarnOnInsecure:   true,
		ServiceNamespace: serviceNamespace,
		ProfileKey:       key,
	}

	// Apply driver options to configure the store
	for _, opt := range driverOpts {
		if err := opt(); err != nil {
			return nil, errors.Join(ErrStoreDriverSetup, err)
		}
	}

	// Use global storeDirectory if set by WithStoreDirectory option
	baseDir := config.StoreDirectory
	if baseDir == "" {
		baseDir = storeDirectory
	}
	if baseDir == "" {
		execPath, err := os.Executable()
		if err != nil {
			return nil, fmt.Errorf("%w: unable to determine executable path", ErrStoreDriverSetup)
		}
		execDir := filepath.Dir(execPath)
		baseDir = filepath.Join(execDir, "profiles")
	}

	// Ensure directory exists with proper permissions
	if err := os.MkdirAll(baseDir, ownerPermissionsRWX); err != nil {
		return nil, fmt.Errorf("%w: failed to create profiles directory %s", ErrStoreDriverSetup, baseDir)
	}

	// Validate security configuration
	if err := ValidateSecurityConfig(config); err != nil {
		return nil, fmt.Errorf("%w: invalid security config: %v", ErrSecurityValidationFailed, err)
	}

	// Check keyring availability and adjust security mode if needed
	keyringAvailable := isKeyringAvailable(serviceNamespace)
	if err := ValidateSecurityMode(config.Mode, keyringAvailable); err != nil {
		if config.Mode == SecurityModeKeyring && !keyringAvailable {
			if config.WarnOnInsecure {
				// Log warning about falling back to insecure mode
				// In a real implementation, this should use a proper logging framework
				fmt.Fprintf(os.Stderr, "Warning: Keyring unavailable, falling back to plaintext storage for profile '%s'\n", key)
			}
			config.Mode = SecurityModeInsecure
		} else {
			return nil, err
		}
	}

	config.StoreDirectory = baseDir
	urn := BuildNamespaceURN(serviceNamespace, version1)
	baseFileName := fmt.Sprintf("%s.%s", urn, key)

	return &HybridStore{
		config:              config,
		secureFilePath:      filepath.Join(baseDir, baseFileName+".secure.enc"),
		plaintextFilePath:   filepath.Join(baseDir, baseFileName+".plaintext.json"),
		namespaceVersionURN: urn,
		temporaryData:       make(map[string]interface{}),
	}, nil
}

// WithSecurityMode sets the security mode for hybrid storage
func WithSecurityMode(mode SecurityMode) DriverOpt {
	return func() error {
		// This would need to be stored somewhere accessible to NewHybridStore
		// For now, we'll handle this in the store construction
		return nil
	}
}

// WithWarnOnInsecure configures warning behavior for insecure fallbacks
func WithWarnOnInsecure(warn bool) DriverOpt {
	return func() error {
		// Similar to above, would need proper configuration passing
		return nil
	}
}

// appVersion stores the application version for metadata
var appVersion string

// WithAppVersion sets the application version for profile metadata
func WithAppVersion(version string) DriverOpt {
	return func() error {
		appVersion = version
		return nil
	}
}

// Exists checks if either secure or plaintext data exists
func (h *HybridStore) Exists() bool {
	secureExists := h.secureFileExists()
	plaintextExists := h.plaintextFileExists()
	return secureExists || plaintextExists
}

// Get retrieves and reconstructs the profile from hybrid storage
func (h *HybridStore) Get() ([]byte, error) {
	// Read secure data
	var secureBytes []byte
	var err error
	if h.secureFileExists() {
		secureBytes, err = h.getSecureData()
		if err != nil {
			return nil, fmt.Errorf("failed to read secure data: %w", err)
		}
	}

	// Read plaintext data
	var plaintextBytes []byte
	if h.plaintextFileExists() {
		plaintextBytes, err = h.getPlaintextData()
		if err != nil {
			return nil, fmt.Errorf("failed to read plaintext data: %w", err)
		}
	}

	// Deserialize split data
	splitData, err := DeserializeSplitData(secureBytes, plaintextBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize split data: %w", err)
	}

	// Add temporary data
	for fieldName, value := range h.temporaryData {
		splitData.AddTemporaryData(fieldName, value)
	}

	// Reconstruct the complete profile
	// Note: This requires knowing the original profile type
	// In practice, this would be handled by the ProfileStore layer
	combinedData := make(map[string]interface{})
	for k, v := range splitData.SecureData {
		combinedData[k] = v
	}
	for k, v := range splitData.PlaintextData {
		combinedData[k] = v
	}
	for k, v := range splitData.TemporaryData {
		combinedData[k] = v
	}

	// Serialize the combined data as JSON
	return json.Marshal(combinedData)
}

// Set splits the profile data and stores it according to security classification
func (h *HybridStore) Set(value interface{}) error {
	// Split the profile data based on security tags
	splitData, err := SplitProfileData(value)
	if err != nil {
		return fmt.Errorf("failed to split profile data: %w", err)
	}

	// Store secure data if present
	if splitData.HasSecureData() {
		if err := h.setSecureData(splitData); err != nil {
			return fmt.Errorf("failed to store secure data: %w", err)
		}
	}

	// Store plaintext data if present
	if splitData.HasPlaintextData() {
		if err := h.setPlaintextData(splitData); err != nil {
			return fmt.Errorf("failed to store plaintext data: %w", err)
		}
	}

	// Update temporary data in memory
	h.temporaryData = make(map[string]interface{})
	for k, v := range splitData.TemporaryData {
		h.temporaryData[k] = v
	}

	// Save metadata
	return h.saveMetadata(splitData)
}

// Delete removes all stored data (secure, plaintext, and metadata)
func (h *HybridStore) Delete() error {
	var errs []error

	// Remove secure file
	if h.secureFileExists() {
		if err := os.Remove(h.secureFilePath); err != nil {
			errs = append(errs, fmt.Errorf("failed to remove secure file: %w", err))
		}
	}

	// Remove plaintext file
	if h.plaintextFileExists() {
		if err := os.Remove(h.plaintextFilePath); err != nil {
			errs = append(errs, fmt.Errorf("failed to remove plaintext file: %w", err))
		}
	}

	// Remove metadata file
	metadataPath := h.getMetadataFilePath()
	if _, err := os.Stat(metadataPath); err == nil {
		if err := os.Remove(metadataPath); err != nil {
			errs = append(errs, fmt.Errorf("failed to remove metadata file: %w", err))
		}
	}

	// Clear temporary data
	h.temporaryData = make(map[string]interface{})

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// SetTemporary adds temporary data that won't be persisted
func (h *HybridStore) SetTemporary(fieldName string, value interface{}) {
	h.temporaryData[fieldName] = value
}

// ClearTemporary removes all temporary data
func (h *HybridStore) ClearTemporary() {
	h.temporaryData = make(map[string]interface{})
}

// ClearTemporaryField removes a specific temporary field
func (h *HybridStore) ClearTemporaryField(fieldName string) {
	delete(h.temporaryData, fieldName)
}

// Helper methods for file operations

func (h *HybridStore) secureFileExists() bool {
	_, err := os.Stat(h.secureFilePath)
	return err == nil
}

func (h *HybridStore) plaintextFileExists() bool {
	_, err := os.Stat(h.plaintextFilePath)
	return err == nil
}

func (h *HybridStore) getSecureData() ([]byte, error) {
	if h.config.Mode == SecurityModeInsecure {
		return nil, fmt.Errorf("%w: cannot read secure data in insecure mode", ErrSecurityModeUnavailable)
	}

	key, err := h.getEncryptionKey()
	if err != nil {
		return nil, err
	}

	encryptedData, err := os.ReadFile(h.secureFilePath)
	if err != nil {
		return nil, err
	}

	return DecryptData(key, encryptedData)
}

func (h *HybridStore) setSecureData(splitData *SplitData) error {
	if h.config.Mode == SecurityModeInsecure {
		return fmt.Errorf("%w: cannot store secure data in insecure mode", ErrSecurityModeUnavailable)
	}

	data, err := SerializeSecureData(splitData)
	if err != nil {
		return err
	}

	key, err := h.getEncryptionKey()
	if err != nil {
		return err
	}

	encryptedData, err := EncryptData(key, data)
	if err != nil {
		return err
	}

	return os.WriteFile(h.secureFilePath, encryptedData, ownerPermissionsRW)
}

func (h *HybridStore) getPlaintextData() ([]byte, error) {
	return os.ReadFile(h.plaintextFilePath)
}

func (h *HybridStore) setPlaintextData(splitData *SplitData) error {
	data, err := SerializePlaintextData(splitData)
	if err != nil {
		return err
	}

	return os.WriteFile(h.plaintextFilePath, data, ownerPermissionsRW)
}

func (h *HybridStore) getEncryptionKey() ([]byte, error) {
	keyStr, err := keyring.Get(h.namespaceVersionURN, h.config.ProfileKey)
	if errors.Is(err, keyring.ErrNotFound) {
		// Generate new key
		key := make([]byte, aes256KeyLength)
		if _, err := rand.Read(key); err != nil {
			return nil, err
		}
		
		if err := keyring.Set(h.namespaceVersionURN, h.config.ProfileKey, string(key)); err != nil {
			return nil, err
		}
		
		return key, nil
	} else if err != nil {
		return nil, err
	}
	
	return []byte(keyStr), nil
}

func (h *HybridStore) saveMetadata(splitData *SplitData) error {
	// Import the versions package constants
	goOSProfilesVersion := "1.0.0"  // This should match GoOSProfilesVersionCurrent
	profileFormatVersion := "2.0"   // This should match ProfileFormatVersionCurrent
	
	metadata := HybridMetadata{
		ProfileName:          h.config.ProfileKey,
		CreatedAt:           time.Now().Format(time.RFC3339),
		LastModified:        time.Now().Format(time.RFC3339),
		SecurityMode:        fmt.Sprintf("%d", h.config.Mode),
		HasSecureData:       splitData.HasSecureData(),
		HasPlaintextData:    splitData.HasPlaintextData(),
		Version:             h.namespaceVersionURN, // Legacy field for compatibility
		
		// New versioning fields
		GoOSProfilesVersion:  goOSProfilesVersion,
		AppVersion:          appVersion,
		ProfileFormatVersion: profileFormatVersion,
	}

	if splitData.HasSecureData() && h.config.Mode == SecurityModeKeyring {
		metadata.EncryptionAlg = "AES-256-GCM"
	}

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(h.getMetadataFilePath(), data, ownerPermissionsRW)
}

func (h *HybridStore) getMetadataFilePath() string {
	return strings.TrimSuffix(h.secureFilePath, ".secure.enc") + ".metadata.json"
}

// isKeyringAvailable checks if the keyring is available for the given service
func isKeyringAvailable(serviceNamespace string) bool {
	testKey := "test-availability-" + serviceNamespace
	
	// Try to set and get a test value
	err := keyring.Set(serviceNamespace, testKey, "test")
	if err != nil {
		return false
	}
	
	_, err = keyring.Get(serviceNamespace, testKey)
	if err != nil {
		return false
	}
	
	// Clean up test key
	_ = keyring.Delete(serviceNamespace, testKey)
	
	return true
}