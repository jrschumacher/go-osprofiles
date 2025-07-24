package main

import (
	"log"

	"github.com/jrschumacher/go-osprofiles"
)

// SystemPolicy represents read-only corporate/admin policies
type SystemPolicy struct {
	Name               string   `json:"name"`
	RequiredServerURL  string   `json:"required_server_url"`
	AllowedFeatures    []string `json:"allowed_features"`
	MaxUploadSizeMB    int      `json:"max_upload_size_mb"`
	EnforceSSL         bool     `json:"enforce_ssl"`
}

func (s SystemPolicy) GetName() string { return "system-policy" }

// UserPreferences represents user's personal settings
type UserPreferences struct {
	Name           string `json:"name"`
	Theme          string `json:"theme"`
	Language       string `json:"language"`
	AutoSave       bool   `json:"auto_save"`
	NotificationsEnabled bool `json:"notifications_enabled"`
}

func (u UserPreferences) GetName() string { return "user-preferences" }

// AppConfig represents the merged configuration used by the application
type AppConfig struct {
	ServerURL        string
	Theme            string
	Language         string
	AutoSave         bool
	Notifications    bool
	AllowedFeatures  []string
	MaxUploadSizeMB  int
	SSLRequired      bool
}

func main() {
	// Create two profilers: one for system policies, one for user preferences
	systemProfiler, err := osprofiles.New("com.mycompany.myapp", osprofiles.WithSystemStore())
	if err != nil {
		log.Fatal(err)
	}

	userProfiler, err := osprofiles.New("com.mycompany.myapp") // Default: user store with encryption
	if err != nil {
		log.Fatal(err)
	}

	// Load system policies (read-only, set by IT/admin)
	var systemPolicy SystemPolicy
	if systemProfiler.ProfileExists("system-policy") {
		systemPolicy, err = osprofiles.GetProfile[SystemPolicy](systemProfiler, "system-policy")
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Loaded system policies")
	} else {
		log.Println("No system policies found - using defaults")
		systemPolicy = SystemPolicy{
			Name:               "system-policy",
			RequiredServerURL:  "https://api.company.com",
			AllowedFeatures:    []string{"sync", "backup", "sharing"},
			MaxUploadSizeMB:    100,
			EnforceSSL:         true,
		}
	}

	// Load user preferences (read-write)
	var userPrefs UserPreferences
	if userProfiler.ProfileExists("user-preferences") {
		userPrefs, err = osprofiles.GetProfile[UserPreferences](userProfiler, "user-preferences")
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Loaded user preferences")
	} else {
		log.Println("Creating default user preferences")
		userPrefs = UserPreferences{
			Name:                 "user-preferences",
			Theme:                "light",
			Language:             "en",
			AutoSave:             true,
			NotificationsEnabled: true,
		}
	}

	// User makes changes to their preferences
	userPrefs.Theme = "dark"
	userPrefs.AutoSave = false

	// Save user changes (system policies remain untouched)
	err = userProfiler.AddProfile(userPrefs, true)
	if err != nil {
		log.Fatal(err)
	}

	// Merge system policies + user preferences into final app config
	appConfig := AppConfig{
		// System policies take precedence for security/compliance
		ServerURL:       systemPolicy.RequiredServerURL,
		AllowedFeatures: systemPolicy.AllowedFeatures,
		MaxUploadSizeMB: systemPolicy.MaxUploadSizeMB,
		SSLRequired:     systemPolicy.EnforceSSL,
		
		// User preferences for personalization
		Theme:         userPrefs.Theme,
		Language:      userPrefs.Language,
		AutoSave:      userPrefs.AutoSave,
		Notifications: userPrefs.NotificationsEnabled,
	}

	// Use the merged configuration
	log.Println("Final application configuration:")
	log.Printf("  Server: %s (from system policy)", appConfig.ServerURL)
	log.Printf("  Theme: %s (from user preferences)", appConfig.Theme)
	log.Printf("  Language: %s (from user preferences)", appConfig.Language)
	log.Printf("  Max upload: %dMB (from system policy)", appConfig.MaxUploadSizeMB)
	log.Printf("  SSL required: %t (from system policy)", appConfig.SSLRequired)

	startApplication(appConfig)
}

func startApplication(config AppConfig) {
	log.Println("Application started with merged user + system configuration")
	// Your application logic here using the merged config...
}