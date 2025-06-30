# go-osprofiles

A Go library for managing application profiles with **field-level security classification** and **hybrid storage backends** native to the operating system.

## 🔐 Key Features

- **Field-Level Security**: Classify profile fields as `secure`, `plaintext`, or `temporary` using struct tags
- **Hybrid Storage**: Automatically splits data across encrypted files (secure) and plaintext files (non-sensitive)
- **Windows Keyring Solution**: Solves Windows keyring size limitations by storing only encryption keys in keyring
- **Temporary Overrides**: Runtime field overrides for flag-based configuration without persistence
- **Automatic Migration**: Seamless migration from legacy storage backends with versioning support
- **Cross-Platform**: Native OS integration for macOS, Linux, and Windows

## 🚀 Quick Start

### Basic Usage with Hybrid Security Storage

```go
package main

import (
    "fmt"
    "github.com/jrschumacher/go-osprofiles"
)

// Define your profile with security classification
type MyProfile struct {
    Name      string `json:"name" security:"plaintext"`      // Stored unencrypted
    APIKey    string `json:"api_key" security:"secure"`      // Encrypted + keyring key
    Endpoint  string `json:"endpoint" security:"plaintext"`  // Stored unencrypted
    TempToken string `json:"temp_token" security:"temporary"` // Memory only
}

func (p *MyProfile) GetName() string { return p.Name }

func main() {
    // Create profiler with hybrid security storage
    profiler, err := profiles.New("my-app", 
        profiles.WithHybridStore("./profiles", "1.0.0"), // directory, app version
        profiles.WithSecurityMode(store.SecurityModeKeyring),
    )
    if err != nil {
        panic(err)
    }

    // Create and add a profile
    profile := &MyProfile{
        Name:      "production",
        APIKey:    "secret-api-key-123",
        Endpoint:  "https://api.example.com",
        TempToken: "temp-session-token",
    }

    err = profiler.AddProfile(profile, true) // true = set as default
    if err != nil {
        panic(err)
    }

    fmt.Println("Profile saved with hybrid security!")
}
```

## 📁 Storage Backends

### 1. Hybrid Security Storage (Recommended)
**Default backend** that automatically classifies and splits profile data:

- **Secure fields**: AES-256-GCM encrypted files + OS keyring for key management
- **Plaintext fields**: Unencrypted JSON files for non-sensitive data
- **Temporary fields**: Memory-only storage, never persisted

```go
profiler, err := profiles.New("my-app", 
    profiles.WithHybridStore("./profiles", "1.0.0"),
    profiles.WithSecurityMode(store.SecurityModeKeyring), // or SecurityModeInsecure
)
```

### 2. Legacy Storage Backends
Maintained for backward compatibility and migration:

```go
// OS Keyring (legacy)
profiler, err := profiles.New("my-app", profiles.WithKeyringStore())

// Encrypted File System (legacy)
profiler, err := profiles.New("my-app", profiles.WithFileStore("./profiles"))

// In-Memory (testing)
profiler, err := profiles.New("my-app", profiles.WithInMemoryStore())
```

## 🔄 Migration from Legacy Storage

Automatically migrate existing profiles to hybrid storage:

```go
// 1. Validate migration (dry run)
report, err := profiles.ValidateMigration[MyProfile]("my-app", global.PROFILE_DRIVER_FILE)
if err != nil {
    fmt.Printf("Migration validation failed: %v\n", err)
}

// 2. Perform migration with automatic backup
report, err = profiles.MigrateProfiles[MyProfile]("my-app", global.PROFILE_DRIVER_FILE)
if err != nil {
    fmt.Printf("Migration failed: %v\n", err)
} else {
    fmt.Printf("Successfully migrated %d profiles\n", report.ProfilesMigrated)
    if report.BackupCreated {
        fmt.Printf("Backup created at: %s\n", report.BackupLocation)
    }
}
```

## 🏷️ Security Classification

Use struct tags to classify field security levels:

| Tag | Storage | Use Case |
|-----|---------|----------|
| `security:"secure"` | Encrypted file + keyring key | API keys, passwords, tokens |
| `security:"plaintext"` | Unencrypted JSON file | URLs, usernames, non-sensitive config |
| `security:"temporary"` | Memory only | Session data, runtime flags |

```go
type Profile struct {
    // Secure: encrypted storage
    APIKey    string `json:"api_key" security:"secure"`
    Password  string `json:"password" security:"secure"`
    
    // Plaintext: unencrypted storage  
    Username  string `json:"username" security:"plaintext"`
    Endpoint  string `json:"endpoint" security:"plaintext"`
    
    // Temporary: memory only
    SessionID string `json:"session_id" security:"temporary"`
    DebugMode bool   `json:"debug_mode" security:"temporary"`
}
```

## ⚡ Temporary Field Overrides

Override field values at runtime without persisting changes (perfect for CLI flags):

```go
// Load profile
profileStore, err := profiles.UseProfile[MyProfile](profiler, "production")

// Set temporary override (e.g., from CLI flag)
err = profileStore.SetTemporary("Endpoint", "https://staging.api.example.com")

// Get profile with overrides applied
profile, err := profileStore.GetProfileWithOverrides()
// profile.Endpoint is now "https://staging.api.example.com" but not saved

// Clear temporary overrides
profileStore.ClearTemporary()
```

## 🔧 Advanced Configuration

### Security Modes

```go
// Secure mode (default): Uses keyring for encryption keys
profiles.WithSecurityMode(store.SecurityModeKeyring)

// Insecure mode: Falls back to plaintext with warnings
profiles.WithSecurityMode(store.SecurityModeInsecure)

// Configure warning behavior
profiles.WithInsecureWarnings(true) // Show warnings for plaintext fallback
```

### Versioning Support

All profiles include automatic versioning:
- **go-osprofiles version**: Library version for compatibility
- **App version**: Your application version for custom migration logic  
- **Profile format version**: Evolution tracking (v0 → v1 → v2)

## 🛡️ Security Benefits

1. **Windows Keyring Solution**: Only encryption keys stored in keyring (not full profile data)
2. **Minimal Attack Surface**: Sensitive data encrypted, non-sensitive data accessible
3. **Graceful Degradation**: Automatic fallback when keyring unavailable
4. **Security Validation**: Built-in detection of sensitive data in plaintext fields
5. **Audit Trail**: Comprehensive metadata and version tracking

## 📋 Platform Support

- **macOS**: Keychain integration
- **Linux**: Secret Service API (GNOME Keyring, KDE Wallet)
- **Windows**: Windows Credential Manager

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Test specific packages
go test ./pkg/store    # Storage backends
go test ./pkg/platform # OS platform integration

# Build and validate
go build ./...
go vet ./...
```

## 📚 API Documentation

For detailed API documentation, see the [CLAUDE.md](./CLAUDE.md) file which contains comprehensive architectural guidance and usage patterns.

## 🤝 Contributing

This project was born out of [OpenTDF](https://github.com/opentdf/platform) and [otdfctl](https://github.com/opentdf/otdfctl) to provide a robust, security-first profile management solution.

## 📄 License

See [LICENSE](./LICENSE) file for details.
