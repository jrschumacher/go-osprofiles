package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/jrschumacher/go-osprofiles/pkg/platform"
	"github.com/jrschumacher/go-osprofiles/pkg/store"
)

// ComplexConfig demonstrates the power of JSON strings in plist files
type ComplexConfig struct {
	Name           string `json:"name"`
	DatabaseConfig struct {
		Primary struct {
			Host            string `json:"host"`
			Port            int    `json:"port"`
			Database        string `json:"database"`
			ConnectionPool  struct {
				MinConnections int `json:"min_connections"`
				MaxConnections int `json:"max_connections"`
				Timeout        int `json:"timeout_seconds"`
			} `json:"connection_pool"`
			SSL struct {
				Enabled  bool   `json:"enabled"`
				Mode     string `json:"mode"`
				CertPath string `json:"cert_path,omitempty"`
			} `json:"ssl"`
		} `json:"primary"`
		ReadReplicas []struct {
			Host   string `json:"host"`
			Port   int    `json:"port"`
			Weight int    `json:"weight"`
		} `json:"read_replicas"`
	} `json:"database_config"`
	ServiceEndpoints struct {
		API       string `json:"api"`
		WebSocket string `json:"websocket"`
		GraphQL   string `json:"graphql,omitempty"`
	} `json:"service_endpoints"`
	FeatureFlags map[string]bool `json:"feature_flags"`
	Logging      struct {
		Level    string   `json:"level"`
		Outputs  []string `json:"outputs"`
		Rotation struct {
			MaxSize string `json:"max_size"`
			MaxAge  string `json:"max_age"`
			MaxBackups int `json:"max_backups"`
		} `json:"rotation"`
	} `json:"logging"`
	Security struct {
		EncryptionKey string   `json:"encryption_key"`
		AllowedOrigins []string `json:"allowed_origins"`
		RateLimit     struct {
			RequestsPerMinute int `json:"requests_per_minute"`
			BurstLimit        int `json:"burst_limit"`
		} `json:"rate_limit"`
	} `json:"security"`
}

func (c ComplexConfig) GetName() string { return "complex-config" }

func main() {
	// Set up test environment with sample plist files
	testDir := filepath.Join(os.TempDir(), "osprofiles-mdm-example")
	if os.Getenv("OSPROFILES_TEST_BASE_PATH") == "" {
		os.Setenv("OSPROFILES_TEST_BASE_PATH", testDir)
		setupTestData(testDir) // Copy sample plist files to test directory
	}
	
	log.Println("Starting MDM JSON plist parsing example")

	// Demo configurations to test different plist setups
	demoConfigurations := []struct{
		name string
		publisher string
		namespace string
		description string // Used for documentation
	}{
		{"Complex Mixed Config", "com.company", "complex-app", "Mixed plist with JSON strings"},
		{"JSON Root Config", "com.company", "json-root", "Single JSON string as root"},
	}

	for _, demo := range demoConfigurations {
		
		// Create platform with demo-specific publisher/namespace
		demoPlatform, err := platform.NewPlatform(demo.publisher, demo.namespace, runtime.GOOS)
		if err != nil {
			log.Fatalf("Failed to create platform: %v", err)
		}
		
		// Get system directory with MDM support
		_, opts := demoPlatform.SystemAppDataDirectoryWithMDM()
		
		// Create store with MDM support  
		configStore, err := store.NewFileStore(demo.namespace, "app-config", opts...)
		if err != nil {
			log.Fatalf("Failed to create config store: %v", err)
		}
		
		if configStore.Exists() {
			configData, err := configStore.Get()
			if err != nil {
				log.Fatalf("Error reading config: %v", err)
			}
			
			// Parse and display the JSON structure from the plist
			var complexConfig ComplexConfig
			if err := json.Unmarshal(configData, &complexConfig); err == nil {
				log.Printf("Successfully parsed %s configuration", demo.name)
				// Example: Access parsed data
				log.Printf("Database host: %s", complexConfig.DatabaseConfig.Primary.Host)
				log.Printf("API endpoint: %s", complexConfig.ServiceEndpoints.API)
			} else {
				log.Printf("Config exists but failed to parse: %v", err)
			}
		} else {
			// No configuration found - this is normal in development
			log.Printf("No MDM configuration found for %s", demo.name)
		}
	}

	
	log.Println("MDM JSON plist parsing example completed")
}

// setupTestData copies the test plist files to the test directory
func setupTestData(testDir string) {
	// Create the managed preferences directory structure
	mdmDir := filepath.Join(testDir, "Library", "Managed Preferences")
	if err := os.MkdirAll(mdmDir, 0755); err != nil {
		return // Silently continue if directory creation fails
	}
	
	// Copy test plist files from testdata directory
	testDataDir := "testdata"
	entries, err := os.ReadDir(testDataDir)
	if err != nil {
		return // No testdata directory - that's ok
	}
	
	// Copy all .plist files to the test directory
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".plist" {
			srcPath := filepath.Join(testDataDir, entry.Name())
			dstPath := filepath.Join(mdmDir, entry.Name())
			
			data, err := os.ReadFile(srcPath)
			if err != nil {
				continue
			}
			
			os.WriteFile(dstPath, data, 0644)
		}
	}
}

// Example of how the MDM plist file looks:
// /Library/Managed Preferences/com.company.complex-app.plist
//
// <?xml version="1.0" encoding="UTF-8"?>
// <!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
// <plist version="1.0">
// <string>{"name":"Production Config","database_config":{...}}</string>
// </plist>
//
// The JSON string is automatically parsed into the ComplexConfig struct.