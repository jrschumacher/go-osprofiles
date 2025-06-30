package profiles

import (
	"encoding/json"

	"github.com/jrschumacher/go-osprofiles/internal/global"
	"github.com/jrschumacher/go-osprofiles/pkg/store"
)

type ProfileStore struct {
	// Store is the specific initialized driver that satisfies the StoreInterface.
	store store.StoreInterface
	// Profile is the struct that holds the profile data and satisfies the NamedProfile interface.
	// Exported to allow write/read access to the profile data being stored.
	Profile NamedProfile
	// temporaryOverrides holds temporary field overrides that don't get persisted
	temporaryOverrides map[string]interface{}
}

// NamedProfile is the holder of a profile containing a name and all stored profile data.
// It is marshaled on Get and unmarshaled on Set, so an interface is used to allow
// for any struct to be stored. The struct satisfying the interface must have JSON tags
// for each stored field.
//
// Example:
//
//	type MyProfile struct {
//		 Name string `json:"name"`
//		 Email string `json:"email"`
//	}
//
//	func (p *MyProfile) GetName() string {
//	 return p.Name
//	}
type NamedProfile interface {
	GetName() string
}

func NewProfileStore(serviceNamespace string, newStore store.NewStoreInterface, profile NamedProfile) (*ProfileStore, error) {
	profileName := profile.GetName()

	if err := validateProfileName(profileName); err != nil {
		return nil, err
	}

	store, err := newStore(serviceNamespace, getStoreKey(profileName))
	if err != nil {
		return nil, err
	}

	p := &ProfileStore{
		store:              store,
		Profile:            profile,
		temporaryOverrides: make(map[string]interface{}),
	}
	return p, nil
}

func LoadProfileStore[T NamedProfile](serviceNamespace string, newStore store.NewStoreInterface, profileName string) (*ProfileStore, error) {
	if err := validateProfileName(profileName); err != nil {
		return nil, err
	}

	store, err := newStore(serviceNamespace, getStoreKey(profileName))
	if err != nil {
		return nil, err
	}

	p := &ProfileStore{
		store:              store,
		temporaryOverrides: make(map[string]interface{}),
	}
	_, err = GetStoredProfile[T](p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// Generic wrapper for working with specific types
func GetStoredProfile[T NamedProfile](store *ProfileStore) (T, error) {
	var profile T
	data, err := store.store.Get()
	if err != nil {
		return profile, err
	}
	err = json.Unmarshal(data, &profile)
	store.Profile = profile
	return profile, err
}

// Save the current profile data to the store
func (p *ProfileStore) Save() error {
	return p.store.Set(p.Profile)
}

// Delete the current profile from the store
func (p *ProfileStore) Delete() error {
	return p.store.Delete()
}

// Profile Name
func (p *ProfileStore) GetProfileName() string {
	return p.Profile.GetName()
}

// utility functions

func getStoreKey(n string) string {
	return global.STORE_KEY_PROFILE + "-" + n
}

// SetTemporary sets a temporary field override that won't be persisted to storage
func (p *ProfileStore) SetTemporary(fieldName string, value interface{}) error {
	if p.temporaryOverrides == nil {
		p.temporaryOverrides = make(map[string]interface{})
	}
	
	// Validate that the field exists in the profile struct
	if err := store.SetFieldValue(p.Profile, fieldName, value); err != nil {
		return err
	}
	
	// Store the override without persisting
	p.temporaryOverrides[fieldName] = value
	return nil
}

// ClearTemporary removes all temporary field overrides
func (p *ProfileStore) ClearTemporary() {
	p.temporaryOverrides = make(map[string]interface{})
}

// ClearTemporaryField removes a specific temporary field override
func (p *ProfileStore) ClearTemporaryField(fieldName string) {
	if p.temporaryOverrides != nil {
		delete(p.temporaryOverrides, fieldName)
	}
}

// GetTemporaryFields returns a copy of all temporary field overrides
func (p *ProfileStore) GetTemporaryFields() map[string]interface{} {
	if p.temporaryOverrides == nil {
		return make(map[string]interface{})
	}
	
	// Return a copy to prevent external modification
	result := make(map[string]interface{})
	for k, v := range p.temporaryOverrides {
		result[k] = v
	}
	return result
}

// HasTemporaryOverrides returns true if there are any temporary overrides
func (p *ProfileStore) HasTemporaryOverrides() bool {
	return p.temporaryOverrides != nil && len(p.temporaryOverrides) > 0
}

// GetProfileWithOverrides returns the profile with temporary overrides applied
func (p *ProfileStore) GetProfileWithOverrides() (NamedProfile, error) {
	if !p.HasTemporaryOverrides() {
		return p.Profile, nil
	}
	
	// Create a copy of the profile structure
	profileCopy, err := store.CopyProfileStructure(p.Profile)
	if err != nil {
		return nil, err
	}
	
	// Set all original field values
	originalClassification, err := store.ClassifyProfile(p.Profile)
	if err != nil {
		return nil, err
	}
	
	allFields := append(originalClassification.SecureFields, originalClassification.PlaintextFields...)
	allFields = append(allFields, originalClassification.TemporaryFields...)
	
	for _, field := range allFields {
		if err := store.SetFieldValue(&profileCopy, field.Name, field.Value); err != nil {
			// Skip fields that can't be set
			continue
		}
	}
	
	// Apply temporary overrides
	for fieldName, value := range p.temporaryOverrides {
		if err := store.SetFieldValue(&profileCopy, fieldName, value); err != nil {
			// Skip fields that can't be set
			continue
		}
	}
	
	// Cast back to NamedProfile
	if namedProfile, ok := profileCopy.(NamedProfile); ok {
		return namedProfile, nil
	}
	
	return nil, ErrInvalidProfile
}
