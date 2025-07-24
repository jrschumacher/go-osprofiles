package main

import (
	"encoding/json"
	"log"
	"runtime"

	"github.com/jrschumacher/go-osprofiles/pkg/platform"
	"github.com/jrschumacher/go-osprofiles/pkg/store"
)

// EnterpriseConfig represents configuration that can come from MDM or fallback sources
type EnterpriseConfig struct {
	Name           string            `json:"name"`
	ServerEndpoint string            `json:"server_endpoint"`
	DatabaseURL    string            `json:"database_url"`
	FeatureFlags   map[string]bool   `json:"feature_flags"`
	SecurityPolicy map[string]string `json:"security_policy"`
	LogLevel       string            `json:"log_level"`
}

func (e EnterpriseConfig) GetName() string { return "enterprise-config" }

func main() {
	// Only works on macOS - check platform first
	if runtime.GOOS != "darwin" {
		log.Println("MDM example requires macOS")
		log.Println("On other platforms, use regular system store")
		return
	}

	log.Println("macOS MDM configuration example with fallbacks")

	// Create platform with MDM support
	// This checks for MDM plist files first, then falls back to regular system storage
	plat, err := platform.NewPlatform("com.mycompany", "myapp", runtime.GOOS)
	if err != nil {
		log.Fatal(err)
	}

	// Get system directory with MDM support
	// Priority: MDM plist > System files > Defaults
	_, storeOpts := plat.SystemAppDataDirectoryWithMDM()

	// Create store with MDM fallback support
	configStore, err := store.NewFileStore("myapp", "enterprise-config", storeOpts...)
	if err != nil {
		log.Fatal(err)
	}

	var config EnterpriseConfig

	if configStore.Exists() {
		// Configuration found - could be from MDM plist or system files
		configData, err := configStore.Get()
		if err != nil {
			log.Fatal(err)
		}

		// The store automatically handles JSON parsing from plist strings
		if err := json.Unmarshal(configData, &config); err != nil {
			log.Fatal(err)
		}

		log.Println("Enterprise configuration loaded")
		log.Printf("Source: %s", plat.MDMConfigPath()) // Shows where config came from
		log.Printf("Server: %s", config.ServerEndpoint)
		log.Printf("Features enabled: %v", config.FeatureFlags)

	} else {
		// No MDM or system config found - use application defaults
		log.Println("No MDM or system configuration found")
		log.Println("Using application defaults")

		config = EnterpriseConfig{
			Name:           "enterprise-config",
			ServerEndpoint: "https://api.company.com/prod",
			DatabaseURL:    "postgres://localhost:5432/myapp",
			FeatureFlags: map[string]bool{
				"analytics":     true,
				"beta_features": false,
				"audit_logging": true,
			},
			SecurityPolicy: map[string]string{
				"encryption": "AES-256",
				"auth_mode":  "SSO",
			},
			LogLevel: "INFO",
		}
	}

	// Configuration is read-only when from MDM
	// Attempting to modify will fail if source is MDM plist
	testReadOnlyBehavior(configStore, config)

	// Use the configuration
	startApplicationWithConfig(config)
}

func testReadOnlyBehavior(store *store.FileStore, config EnterpriseConfig) {
	// Try to modify the configuration
	config.LogLevel = "DEBUG" // User attempting to enable debug mode
	
	configJSON, _ := json.Marshal(config)
	err := store.Set(configJSON)
	
	if err != nil {
		log.Printf("Configuration is read-only (MDM managed): %v", err)
	} else {
		log.Println("Configuration modified (no MDM active)")
	}
}

func startApplicationWithConfig(config EnterpriseConfig) {
	log.Println("Starting application with enterprise configuration")
	log.Printf("Database: %s", config.DatabaseURL)
	log.Printf("Log level: %s", config.LogLevel)
	
	// Your application startup logic here...
	if config.FeatureFlags["audit_logging"] {
		log.Println("Audit logging enabled")
	}
	
	log.Println("Application started successfully")
}

// Example MDM plist that IT would deploy:
// /Library/Managed Preferences/com.mycompany.myapp.plist
//
// <?xml version="1.0" encoding="UTF-8"?>
// <!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
// <plist version="1.0">
// <string>{"name":"enterprise-config","server_endpoint":"https://prod.company.com","database_url":"postgres://prod-db:5432/myapp","feature_flags":{"analytics":true,"beta_features":false,"audit_logging":true},"security_policy":{"encryption":"AES-256","auth_mode":"SSO"},"log_level":"INFO"}</string>
// </plist>