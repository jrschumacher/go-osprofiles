package main

import (
	"log"
	"runtime"

	osprofiles "github.com/jrschumacher/go-osprofiles"
	"github.com/jrschumacher/go-osprofiles/pkg/platform"
)

// AppData represents application data stored system-wide
type AppData struct {
	Name         string            `json:"name"`
	CacheEnabled bool              `json:"cache_enabled"`
	Databases    map[string]string `json:"databases"`
	LogLevel     string            `json:"log_level"`
}

func (a AppData) GetName() string { return "app-data" }

func main() {
	// Create system profiler - stores in system application directory
	// Cross-platform: /Library/Application Support (macOS), %ProgramData% (Windows), /usr/local/share (Linux)
	plat, err := platform.NewPlatform("com.mycompany", "myapp", runtime.GOOS)
	if err != nil {
		log.Fatal(err)
	}
	systemDir := plat.SystemAppDataDirectory()
	profiler, err := osprofiles.New("com.mycompany.myapp", osprofiles.WithFileStore(systemDir))
	if err != nil {
		log.Fatal(err)
	}

	// Load existing app data or initialize
	var appData AppData
	globalStore := osprofiles.GetGlobalConfig(profiler)
	if globalStore.ProfileExists("app-data") {
		profileStore, err := osprofiles.GetProfile[AppData](profiler, "app-data")
		if err != nil {
			log.Fatal(err)
		}
		appData, err = osprofiles.GetStoredProfile[AppData](profileStore)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Loaded existing application data")
	} else {
		// Initialize system-wide application data
		appData = AppData{
			Name:         "app-data",
			CacheEnabled: true,
			Databases: map[string]string{
				"primary": "sqlite:///var/lib/myapp/primary.db",
				"cache":   "redis://localhost:6379",
			},
			LogLevel: "INFO",
		}
		log.Println("Created default application data")
	}

	// Update application configuration
	appData.LogLevel = "DEBUG"
	appData.Databases["analytics"] = "postgres://localhost:5432/analytics"

	// Save to system directory (requires appropriate permissions)
	err = profiler.AddProfile(appData, true)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Saved app data: log_level=%s, databases=%d", appData.LogLevel, len(appData.Databases))
	log.Println("Application data stored in system directory")
}