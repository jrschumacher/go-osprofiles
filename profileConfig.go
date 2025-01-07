package profiles

import (
	"github.com/jrschumacher/go-osprofiles/internal/global"
	"github.com/jrschumacher/go-osprofiles/pkg/store"
)

type ProfileStore struct {
	// Store is the specific initialized driver that satisfies the StoreInterface.
	store store.StoreInterface
	// Profile is the struct that holds the profile data and satisfies the NamedProfile interface.
	profile NamedProfile
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

// TODO: do we need both of these (New and Load both)?

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
		store:   store,
		profile: profile,
	}
	return p, nil
}

func LoadProfileStore(serviceNamespace string, newStore store.NewStoreInterface, profileName string) (*ProfileStore, error) {
	if err := validateProfileName(profileName); err != nil {
		return nil, err
	}

	store, err := newStore(serviceNamespace, getStoreKey(profileName))
	if err != nil {
		return nil, err
	}

	p := &ProfileStore{
		store: store,
	}
	return p, p.Get()
}

// Get the current profile data from the store
func (p *ProfileStore) Get() error {
	return p.store.Get(&p.profile)
}

// Save the current profile data to the store
func (p *ProfileStore) Save() error {
	return p.store.Set(p.profile)
}

// Delete the current profile from the store
func (p *ProfileStore) Delete() error {
	return p.store.Delete()
}

// Profile Name
func (p *ProfileStore) GetProfileName() string {
	return p.profile.GetName()
}

// utility functions

func getStoreKey(n string) string {
	return global.STORE_KEY_PROFILE + "-" + n
}
