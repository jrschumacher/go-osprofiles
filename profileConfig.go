package profiles

import (
	"github.com/jrschumacher/go-osprofiles/pkg/store"
)

type ProfileStore struct {
	store store.StoreInterface

	config ProfileConfig
}

type ProfileConfig struct {
	Name string `json:"profile"`
	// TODO: map[string]interface{}
	// TODO: interface{}?
	Endpoint        string          `json:"endpoint"`
	TlsNoVerify     bool            `json:"tlsNoVerify"`
	AuthCredentials AuthCredentials `json:"authCredentials"`
}

// TODO: do we need both of these (New and Load both)?

func NewProfileStore(configName string, newStore store.NewStoreInterface, profileName string, endpoint string, tlsNoVerify bool) (*ProfileStore, error) {
	if err := validateProfileName(profileName); err != nil {
		return nil, err
	}

	store, err := newStore(configName, getStoreKey(profileName))
	if err != nil {
		return nil, err
	}

	p := &ProfileStore{
		store: store,
		config: ProfileConfig{
			Name:        profileName,
			Endpoint:    endpoint,
			TlsNoVerify: tlsNoVerify,
		},
	}
	return p, nil
}

func LoadProfileStore(configName string, newStore store.NewStoreInterface, profileName string) (*ProfileStore, error) {
	if err := validateProfileName(profileName); err != nil {
		return nil, err
	}

	store, err := newStore(configName, getStoreKey(profileName))
	if err != nil {
		return nil, err
	}

	p := &ProfileStore{
		store: store,
	}
	return p, p.Get()
}

func (p *ProfileStore) Get() error {
	return p.store.Get(&p.config)
}

func (p *ProfileStore) Save() error {
	return p.store.Set(p.config)
}

func (p *ProfileStore) Delete() error {
	return p.store.Delete()
}

// Profile Name
func (p *ProfileStore) GetProfileName() string {
	return p.config.Name
}

// Endpoint
func (p *ProfileStore) GetEndpoint() string {
	return p.config.Endpoint
}

func (p *ProfileStore) SetEndpoint(endpoint string) error {
	p.config.Endpoint = endpoint
	return p.Save()
}

// TLS No Verify
func (p *ProfileStore) GetTLSNoVerify() bool {
	return p.config.TlsNoVerify
}

func (p *ProfileStore) SetTLSNoVerify(tlsNoVerify bool) error {
	p.config.TlsNoVerify = tlsNoVerify
	return p.Save()
}

// utility functions

func getStoreKey(n string) string {
	return STORE_KEY_PROFILE + "-" + n
}
