package main

import (
	"log"

	"github.com/jrschumacher/go-osprofiles"
)

// SystemConfig represents read-only system configuration
type SystemConfig struct {
	Name           string   `json:"name"`
	ServerEndpoint string   `json:"server_endpoint"`
	AllowedHosts   []string `json:"allowed_hosts"`
	MaxConnections int      `json:"max_connections"`
	DebugMode      bool     `json:"debug_mode"`
}

func (s SystemConfig) GetName() string { return "system-config" }

func main() {
	// Create system config profiler - reads from system configuration directory
	// This is typically managed by administrators or installers
	profiler, err := osprofiles.New("com.mycompany.myapp", osprofiles.WithSystemStore())
	if err != nil {
		log.Fatal(err)
	}

	// Try to load system configuration
	if profiler.ProfileExists("system-config") {
		config, err := osprofiles.GetProfile[SystemConfig](profiler, "system-config")
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Loaded system configuration")
		log.Printf("Server: %s", config.ServerEndpoint)
		log.Printf("Max connections: %d", config.MaxConnections)
		log.Printf("Debug mode: %t", config.DebugMode)

		// Use the configuration in your application
		startServer(config)
	} else {
		log.Println("No system configuration found")
		log.Println("Application will use built-in defaults")
		
		// Fallback to defaults when no system config exists
		defaultConfig := SystemConfig{
			Name:           "system-config",
			ServerEndpoint: "https://api.example.com",
			AllowedHosts:   []string{"localhost", "*.example.com"},
			MaxConnections: 100,
			DebugMode:      false,
		}
		
		startServer(defaultConfig)
	}
}

// Example of using the configuration in your application
func startServer(config SystemConfig) {
	log.Printf("Starting server with endpoint: %s", config.ServerEndpoint)
	log.Printf("Allowed hosts: %v", config.AllowedHosts)
	
	if config.DebugMode {
		log.Println("Debug mode enabled - verbose logging active")
	}
	
	// Your server startup logic here...
	log.Println("Server started successfully")
}