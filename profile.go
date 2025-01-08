package profiles

import (
	"github.com/jrschumacher/go-osprofiles/internal/global"
	"github.com/jrschumacher/go-osprofiles/pkg/store"
)

type profileConfig struct {
	configName string
	driver     global.ProfileDriver

	driverOpts []store.DriverOpt
}

// Profiler is the main interface for managing profiles
type Profiler struct {
	config profileConfig

	globalStore         *global.GlobalStore
	currentProfileStore *ProfileStore
}

type (
	profileConfigVariadicFunc func(profileConfig) profileConfig
)

// Variadic functions to set different storage drivers

func WithInMemoryStore() profileConfigVariadicFunc {
	return func(c profileConfig) profileConfig {
		c.driver = global.PROFILE_DRIVER_IN_MEMORY
		return c
	}
}

func WithKeyringStore() profileConfigVariadicFunc {
	return func(c profileConfig) profileConfig {
		c.driver = global.PROFILE_DRIVER_KEYRING
		return c
	}
}

func WithFileStore(storeDir string) profileConfigVariadicFunc {
	return func(c profileConfig) profileConfig {
		c.driver = global.PROFILE_DRIVER_FILE
		c.driverOpts = append(c.driverOpts, store.WithStoreDirectory(storeDir))
		return c
	}
}

func WithCustomStore(newCustomStore store.NewStoreInterface) profileConfigVariadicFunc {
	return func(c profileConfig) profileConfig {
		c.driver = global.PROFILE_DRIVER_CUSTOM
		store.NewCustomStore = newCustomStore
		return c
	}
}

// newStoreFactory returns a storage interface based on the configured driver
func newStoreFactory(driver global.ProfileDriver) store.NewStoreInterface {
	switch driver {
	case global.PROFILE_DRIVER_KEYRING:
		return store.NewKeyringStore
	case global.PROFILE_DRIVER_IN_MEMORY:
		return store.NewMemoryStore
	case global.PROFILE_DRIVER_FILE:
		return store.NewFileStore
	case global.PROFILE_DRIVER_CUSTOM:
		return store.NewCustomStore
	default:
		return nil
	}
}

// New creates a new Profile with the specified configuration options.
// The configName is required and must be unique to the application.
func New(configName string, opts ...profileConfigVariadicFunc) (*Profiler, error) {
	var err error

	// Apply configuration options
	config := profileConfig{
		driver: global.PROFILE_DRIVER_DEFAULT,
	}
	for _, opt := range opts {
		config = opt(config)
	}

	// Validate and initialize the store
	newStore := newStoreFactory(config.driver)
	if newStore == nil {
		return nil, ErrInvalidStoreDriver
	}

	p := &Profiler{
		config: config,
	}

	// Load global configuration
	p.globalStore, err = global.LoadGlobalConfig(configName, newStore)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// GetGlobalConfig returns the global configuration
func (p *Profiler) GetGlobalConfig() *global.GlobalStore {
	return p.globalStore
}

// AddProfile adds a new profile to the current configuration
func (p *Profiler) AddProfile(profile NamedProfile, setDefault bool) error {
	var err error
	profileName := profile.GetName()

	if err := validateProfileName(profileName); err != nil {
		return err
	}

	// Check if the profile already exists
	if p.globalStore.ProfileExists(profileName) {
		return ErrProfileNameConflict
	}

	// Create profile store and save
	p.currentProfileStore, err = NewProfileStore(p.config.configName, newStoreFactory(p.config.driver), profile)
	if err != nil {
		return err
	}
	if err := p.currentProfileStore.Save(); err != nil {
		return err
	}

	// Add profile to global configuration
	if err := p.globalStore.AddProfile(profileName); err != nil {
		return err
	}

	if setDefault || p.globalStore.GetDefaultProfile() == "" {
		return p.globalStore.SetDefaultProfile(profileName)
	}

	return nil
}

// GetCurrentProfile returns the current stored profile
func (p *Profiler) GetCurrentProfile() (*ProfileStore, error) {
	if p.currentProfileStore == nil {
		return nil, ErrMissingCurrentProfile
	}
	return p.currentProfileStore, nil
}

// GetProfile returns the profile store for the specified profile name
func (p *Profiler) GetProfile(profileName string) (*ProfileStore, error) {
	if !p.globalStore.ProfileExists(profileName) {
		return nil, ErrMissingProfileName
	}
	return LoadProfileStore(p.config.configName, newStoreFactory(p.config.driver), profileName)
}

// ListProfiles returns a list of all profile names
func (p *Profiler) ListProfiles() []string {
	return p.globalStore.ListProfiles()
}

// UseProfile sets the current profile to the specified profile name
func (p *Profiler) UseProfile(profileName string) (*ProfileStore, error) {
	var err error

	// If current profile is already set to this, return it
	if p.currentProfileStore != nil && p.currentProfileStore.Profile.GetName() == profileName {
		return p.currentProfileStore, nil
	}

	// Set current profile
	p.currentProfileStore, err = p.GetProfile(profileName)
	return p.currentProfileStore, err
}

// UseDefaultProfile sets the current profile to the default profile
func (p *Profiler) UseDefaultProfile() (*ProfileStore, error) {
	defaultProfile := p.globalStore.GetDefaultProfile()
	if defaultProfile == "" {
		return nil, ErrMissingDefaultProfile
	}
	return p.UseProfile(defaultProfile)
}

// SetDefaultProfile sets the a specified profile to the default profile
func (p *Profiler) SetDefaultProfile(profileName string) error {
	if !p.globalStore.ProfileExists(profileName) {
		return ErrMissingProfileName
	}
	return p.globalStore.SetDefaultProfile(profileName)
}

// DeleteProfile removes a profile from storage
func (p *Profiler) DeleteProfile(profileName string) error {
	// Check if the profile exists
	if !p.globalStore.ProfileExists(profileName) {
		return ErrMissingProfileName
	}

	// Retrieve the profile
	profile, err := LoadProfileStore(p.config.configName, newStoreFactory(p.config.driver), profileName)
	if err != nil {
		return err
	}

	// Remove profile from global configuration
	if err := p.globalStore.RemoveProfile(profileName); err != nil {
		return err
	}

	return profile.Delete()
}
