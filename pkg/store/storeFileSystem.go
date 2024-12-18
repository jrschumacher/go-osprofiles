package store

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zalando/go-keyring"
)

type FileStore struct {
	namespaceVersionURN string
	namespace           string
	key                 string
	filePath            string
}

// Metadata structure for unencrypted metadata about the encrypted file
type FileMetadata struct {
	ProfileName   string `json:"profile_name"`
	CreatedAt     string `json:"created_at"`
	EncryptionAlg string `json:"encryption_alg"`
	Version       string `json:"version"`
}

const (
	version1            = "v1"
	aes256KeyLength     = 32
	ownerPermissionsRW  = 0o600
	ownerPermissionsRWX = 0o700
)

// TODO: should we use this throughout all stores and add it to the interface?
// URN-based namespace template without UUID, using only profile name for uniqueness
func BuildNamespaceURN(configName, version string) string {
	return fmt.Sprintf("urn:goosprofiles:%s:profile:%s", configName, version) // e.g., urn:goosprofiles:otdfctl:profile:v1:<profileName>
}

// NewFileStore is the constructor function for FileStore, setting the file path based on executable directory or environment variable and hashed filename
var NewFileStore NewStoreInterface = func(namespace string, key string) (StoreInterface, error) {
	if err := ValidateNamespaceKey(namespace, key); err != nil {
		return nil, err
	}

	// Check for OTDFCTL_PROFILE_PATH environment variable
	baseDir := os.Getenv("OTDFCTL_PROFILE_PATH")
	if baseDir == "" {
		// If environment variable is not set, use the "profiles" directory relative to the executable
		execPath, err := os.Executable()
		if err != nil {
			panic("unable to determine the executable path for profile storage")
		}
		execDir := filepath.Dir(execPath)
		baseDir = filepath.Join(execDir, "profiles")
	}
	// Ensure the base directory exists with owner-only access including execute
	if err := os.MkdirAll(baseDir, ownerPermissionsRWX); err != nil {
		panic(fmt.Sprintf("failed to create profiles directory %s: please check directory permissions", baseDir))
	}
	// Check for read/write permissions by creating and removing a temp file
	testFilePath := filepath.Join(baseDir, ".tmp_profile_rw_test")
	testFile, err := os.Create(testFilePath)
	if err != nil {
		panic(fmt.Sprintf("unable to write to profiles directory %s: please ensure write permissions are granted", baseDir))
	}
	testFile.Close()
	if err := os.Remove(testFilePath); err != nil {
		panic(fmt.Sprintf("unable to delete temp file in profiles directory %s: please ensure delete permissions are granted", baseDir))
	}
	urn := BuildNamespaceURN(namespace, version1)
	// Generate the filename hashed for uniqueness
	// Note: other stores use the config.AppName, but want to rely on something more resilient like the namespace
	fileName := fmt.Sprintf("%s_%s", urn, key)
	filePath := filepath.Join(baseDir, fileName+".enc")
	return &FileStore{
		namespaceVersionURN: urn,
		namespace:           namespace,
		key:                 key,
		filePath:            filePath,
	}, nil
}

// Exists checks if the encrypted file exists
func (f *FileStore) Exists() bool {
	_, err := os.Stat(f.filePath)
	return err == nil
}

// Get retrieves and decrypts data from the file
func (f *FileStore) Get(value interface{}) error {
	key, err := f.getEncryptionKey()
	if err != nil {
		return err
	}
	encryptedData, err := os.ReadFile(f.filePath)
	if err != nil {
		return err
	}
	data, err := decryptData(key, encryptedData)
	if err != nil {
		return err
	}
	return json.NewDecoder(bytes.NewReader(data)).Decode(value)
}

// Set encrypts and saves data to the file, also saving metadata
func (f *FileStore) Set(value interface{}) error {
	key, err := f.getEncryptionKey()
	if err != nil {
		return err
	}
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(value); err != nil {
		return err
	}
	encryptedData, err := encryptData(key, b.Bytes())
	if err != nil {
		return err
	}
	// Write the encrypted profile file with proper permissions
	if err := os.WriteFile(f.filePath, encryptedData, ownerPermissionsRW); err != nil {
		return fmt.Errorf("failed to write encrypted profile to %s: %w", f.filePath, err)
	}
	// Save metadata as well
	profileName := f.key // or extract from value if it's part of a ProfileConfig struct
	return f.SaveMetadata(profileName)
}

// Delete removes the encrypted file and metadata file from disk
func (f *FileStore) Delete() error {
	if err := os.Remove(f.filePath); err != nil {
		return err
	}
	// Remove the extension from filePath and add .nfo for the metadata file
	metadataFilePath := strings.TrimSuffix(f.filePath, filepath.Ext(f.filePath)) + ".nfo"
	return os.Remove(metadataFilePath)
}

// getEncryptionKey retrieves the encryption key from the keyring or generates it if absent
func (f *FileStore) getEncryptionKey() ([]byte, error) {
	// Try retrieving the key as a string from the keyring
	keyStr, err := keyring.Get(f.namespaceVersionURN, f.key)
	if errors.Is(err, keyring.ErrNotFound) {
		// Generate a new key if not found
		key := make([]byte, aes256KeyLength)
		if _, err := rand.Read(key); err != nil {
			return nil, err
		}
		// Convert key to string for storage in the keyring
		if err := keyring.Set(f.namespaceVersionURN, f.key, string(key)); err != nil {
			return nil, err
		}
		return key, nil
	} else if err != nil {
		return nil, err
	}
	// Convert the stored string key back to []byte for use
	return []byte(keyStr), nil
}

// encryptData encrypts data using AES-GCM
func encryptData(key, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	// Encrypt the data with a separate destination buffer
	ciphertext := aesGCM.Seal(nil, nonce, data, nil)
	// Prepend the nonce to the ciphertext
	result := make([]byte, len(nonce)+len(ciphertext))
	copy(result, nonce)
	copy(result[len(nonce):], ciphertext)
	return result, nil
}

// decryptData decrypts data using AES-GCM
func decryptData(key, encryptedData []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := aesGCM.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, errors.Join(ErrStoredValueInvalid, ErrEncryptedDataInvalid)
	}
	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]
	return aesGCM.Open(nil, nonce, ciphertext, nil)
}

// SaveMetadata writes unencrypted metadata to a .nfo file
func (f *FileStore) SaveMetadata(profileName string) error {
	metadata := FileMetadata{
		ProfileName:   profileName,
		CreatedAt:     time.Now().Format(time.RFC3339),
		EncryptionAlg: "AES-256-GCM",
		Version:       f.namespaceVersionURN,
	}
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}
	metadataFilePath := strings.TrimSuffix(f.filePath, filepath.Ext(f.filePath)) + ".nfo"
	return os.WriteFile(metadataFilePath, data, ownerPermissionsRW)
}

// LoadMetadata loads and parses metadata from a .nfo file
func (f *FileStore) LoadMetadata() (*FileMetadata, error) {
	metadataFilePath := strings.TrimSuffix(f.filePath, filepath.Ext(f.filePath)) + ".nfo"
	data, err := os.ReadFile(metadataFilePath)
	if err != nil {
		return nil, err
	}
	var metadata FileMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}
	return &metadata, nil
}
