package main

import (
	"log"

	osprofiles "github.com/jrschumacher/go-osprofiles"
)

// UserPreferences represents user settings stored in their personal space
type UserPreferences struct {
	Name     string `json:"name"`
	Theme    string `json:"theme"`
	Language string `json:"language"`
	APIKey   string `json:"api_key"` // Encrypted automatically
}

func (u UserPreferences) GetName() string { return "user-preferences" }

func main() {
	// Create user profiler - stores in user's personal directory
	// Cross-platform: ~/Library/Application Support (macOS), %APPDATA% (Windows), ~/.local/share (Linux)
	profiler, err := osprofiles.New("com.mycompany.myapp")
	if err != nil {
		log.Fatal(err)
	}

	// Storage options:
	// Default: Encrypted files with keys in system keyring (recommended)
	// Alternative: profiler, err := osprofiles.New("com.mycompany.myapp", osprofiles.WithInMemoryStore()) // For testing
	// Alternative: profiler, err := osprofiles.New("com.mycompany.myapp", osprofiles.WithFileStore())    // Unencrypted files

	// Load existing preferences or create defaults
	var prefs UserPreferences
	globalStore := osprofiles.GetGlobalConfig(profiler)
	if globalStore.ProfileExists("user-preferences") {
		profileStore, err := osprofiles.GetProfile[UserPreferences](profiler, "user-preferences")
		if err != nil {
			log.Fatal(err)
		}
		prefs, err = osprofiles.GetStoredProfile[UserPreferences](profileStore)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Loaded existing user preferences")
	} else {
		// Set up defaults for new user
		prefs = UserPreferences{
			Name:     "user-preferences",
			Theme:    "light",
			Language: "en",
			APIKey:   "your-api-key-here",
		}
		log.Println("Created default user preferences")
	}

	// User makes changes
	prefs.Theme = "dark"
	prefs.Language = "es"

	// Save changes (encrypted automatically)
	err = profiler.AddProfile(prefs, true) // true = set as default
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Saved preferences: theme=%s, language=%s", prefs.Theme, prefs.Language)
	log.Println("User preferences stored securely in user directory")
}