# go-osprofiles

A Go library to simplify the creation and management of application profiles native to the operating system. This library provides a unified interface for storing and retrieving application configuration profiles across different storage backends with full support for enterprise MDM (Mobile Device Management) systems.

## 🚀 Features

- **Cross-platform storage** - Works on macOS, Linux, and Windows
- **Multiple storage backends** - Keyring, filesystem, and in-memory
- **Enterprise MDM support** - Seamless integration with macOS managed preferences via platform layer
- **Hierarchical configuration** - MDM managed preferences take precedence over user settings (read-only)
- **Type-safe profiles** - Generic interface for strongly-typed profile management
- **Encryption by default** - File storage uses AES-256-GCM encryption

## 📦 Installation

```bash
go get github.com/jrschumacher/go-osprofiles
```

## 🔧 Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "runtime"
    "github.com/jrschumacher/go-osprofiles"
    "github.com/jrschumacher/go-osprofiles/pkg/platform"
)

type MyProfile struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

func (p *MyProfile) GetName() string {
    return p.Name
}

func main() {
    // Create platform-aware directory resolver
    plat, err := platform.NewPlatform("com.company", "myapp", runtime.GOOS)
    if err != nil {
        panic(err)
    }
    
    // Use appropriate directory (system for admin/dev, user for regular users)
    configDir := plat.UserAppDataDirectory() // or SystemAppDataDirectory()
    
    // Create profiler with platform-aware directory
    profiler, err := profiles.New("myapp", profiles.WithFileStore(configDir))
    if err != nil {
        panic(err)
    }

    // Add a profile
    profile := &MyProfile{Name: "john", Email: "john@example.com"}
    err = profiler.AddProfile(profile, true) // true = set as default
    if err != nil {
        panic(err)
    }

    // Retrieve profile
    stored, err := profiles.GetProfile[*MyProfile](profiler, "john")
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Profile: %+v\n", stored.Profile)
}
```

### Simplified Usage (No Publisher)

If you don't need publisher-based directory organization:

```go
// Simple approach - uses OS-native directories without publisher
profiler, err := profiles.New("com.company.myapp", profiles.WithKeyringStore())
```

## 🏗️ Architecture Overview

This library has a **two-layer architecture** that separates directory resolution from profile management:

### Layer 1: Platform Directory Resolution (`pkg/platform`)

Handles OS-specific directory conventions and publisher-based organization:

```go
// Create platform resolver with publisher and namespace
plat, err := platform.NewPlatform("com.company", "myapp", runtime.GOOS)

// Choose appropriate directory based on your use case:
userDir := plat.UserAppDataDirectory()     // ~/Library/Application Support/com.company/myapp/
systemDir := plat.SystemAppDataDirectory() // /Library/Application Support/com.company/myapp/
```

**Directory patterns per OS:**

| OS | User Data | System Data | Notes |
|---|---|---|---|
| **macOS** | `~/Library/Application Support/<publisher>/<namespace>` | `/Library/Application Support/<publisher>/<namespace>` | User-writable vs system-wide |
| **Linux** | `~/.config/<publisher>/<namespace>` | `/usr/local/<publisher>/<namespace>` | XDG Base Directory vs system |
| **Windows** | `%LOCALAPPDATA%\<publisher>\<namespace>` | `%PROGRAMDATA%\<publisher>\<namespace>` | User profile vs all users |

**Key Distinctions:**
- **User Directories**: Per-user configurations, user-modifiable, no MDM checking
- **System Directories**: System-wide configurations, may require admin privileges, **MDM-aware on macOS**
- **ProgramData vs ProgramFiles** (Windows): ProgramData is writable app data, ProgramFiles is read-only binaries

### Layer 2: Profile Management (`profiles`)

Handles profile storage, encryption, and MDM integration:

```go
// Use directory from platform layer
profiler, err := profiles.New("myapp", profiles.WithFileStore(platformDir))

// Or use built-in storage (keyring, memory, etc.)
profiler, err := profiles.New("com.company.myapp", profiles.WithKeyringStore())
```

### When to Use Each Layer

| Use Case | Approach | Example |
|----------|----------|---------|
| **Publisher-based directories** | Platform + Profiles | Enterprise apps with multiple products |
| **Simple single-product apps** | Profiles only | Small utilities, personal projects |
| **Dev vs Prod environments** | Platform choice | `SystemAppDataDirectory()` for dev, `UserAppDataDirectory()` for prod |
| **MDM enterprise deployment** | RDNS naming | `profiles.New("com.company.app")` |

## 🏢 Enterprise MDM Support

### Overview

The library provides seamless integration with enterprise MDM systems, particularly for macOS environments. MDM (Mobile Device Management) allows IT administrators to deploy and manage application configurations across an organization's devices.

### How MDM Works

1. **IT Deployment**: Administrators deploy managed preferences via MDM systems (Intune, Jamf, etc.)
2. **File Location**: Preferences are stored in `/Library/Managed Preferences/{reverse-dns-id}.plist`
3. **Hierarchical Priority**: MDM settings take precedence over user-defined settings
4. **Read-Only Security**: Applications cannot modify managed preferences

### Key Concepts

- **Reverse DNS Identifier**: MDM requires identifiers like `com.company.app` for plist naming
- **System vs User Directories**: Only system directories check for MDM (user directories are purely user-controlled)
- **Automatic Detection**: When namespace is reverse DNS format, MDM is automatically enabled for system directories
- **Platform Responsibility**: The platform layer handles MDM detection and directory resolution
- **Profile Layer Transparency**: The profiles layer transparently checks MDM first, then falls back to regular storage

### MDM Integration Options

#### Option 1: Automatic MDM Support (Recommended)

For applications with reverse DNS namespace:

```go
// Platform with reverse DNS namespace automatically enables MDM
plat, err := platform.NewPlatform("com.company", "com.company.myapp", runtime.GOOS)
systemDir := plat.SystemAppDataDirectory() // MDM auto-enabled for system directory
profiler, err := profiles.New("myapp", profiles.WithFileStore(systemDir))

// Hierarchy: 
// 1. /Library/Managed Preferences/com.company.myapp.plist (if exists, read-only)
// 2. /Library/Application Support/com.company/com.company.myapp/ (user-writable)
```

#### Option 2: Explicit MDM Identifier

For applications with regular namespace but need MDM:

```go
// Regular namespace with explicit MDM identifier
plat, err := platform.NewPlatform("company", "myapp", runtime.GOOS)
systemDir, opts := plat.SystemAppDataDirectoryWithMDM("com.company.myapp")
profiler, err := profiles.New("myapp", profiles.WithFileStore(systemDir, opts...))

// Hierarchy:
// 1. /Library/Managed Preferences/com.company.myapp.plist (if exists, read-only) 
// 2. /Library/Application Support/company/myapp/ (user-writable)
```

#### Option 3: User Directory (No MDM)

For user-only configurations:

```go
// User directory never checks MDM
plat, err := platform.NewPlatform("com.company", "myapp", runtime.GOOS)
userDir := plat.UserAppDataDirectory() // No MDM checking
profiler, err := profiles.New("myapp", profiles.WithFileStore(userDir))

// Uses only: ~/Library/Application Support/com.company/myapp/
```

### MDM Behavior & Error Handling

| Scenario | MDM File Exists | Read Behavior | Write/Delete Behavior |
|----------|----------------|---------------|----------------------|
| **Key in MDM** | ✅ Yes | Returns MDM value | ❌ OS error → wrapped as "managed remotely (MDM)" |
| **Key not in MDM** | ✅ Yes | Falls back to file storage | ✅ Writes to file storage |
| **No MDM file** | ❌ No | Uses file storage only | ✅ Writes to file storage |
| **Non-macOS** | ❌ N/A | Uses file storage only | ✅ Writes to file storage (or "read-only" errors) |

**OS-Native Error Handling:**
- Let the OS enforce permissions naturally (no explicit blocking)
- Wrap OS write errors with meaningful context for downstream apps
- Same pattern works for Windows ProgramFiles, Linux system directories
- Apps can catch specific error types and show user-friendly messages

**Example Error Types:**
```go
// Downstream apps can handle these specifically
if errors.Is(err, profiles.ErrManagedByMDM) {
    showMessage("Settings are managed by your IT department")
} else if errors.Is(err, profiles.ErrReadOnlyLocation) {
    showMessage("Cannot save to this location - insufficient permissions")
}
```

### Enterprise Example

```go
// Enterprise application setup with automatic MDM
plat, err := platform.NewPlatform("com.acme", "com.acme.secure-app", runtime.GOOS)
systemDir := plat.SystemAppDataDirectory() // Auto-enables MDM
profiler, err := profiles.New("secure-app", profiles.WithFileStore(systemDir))

// IT Admin deploys this plist to all devices:
// /Library/Managed Preferences/com.acme.secure-app.plist
// {
//   "server-url": "https://prod.acme.com",
//   "compliance-mode": true,
//   "user-override-allowed": false
// }

// Application behavior:
// - server-url: Always uses MDM value (managed by IT, read-only)
// - compliance-mode: Always uses MDM value (managed by IT, read-only)  
// - user-preferences: Falls back to encrypted file storage (user configurable)
```

## 🗄️ Storage Backends

### Supported Storage Types

1. **File System** (`WithFileStore()`)
   - Encrypted JSON files using AES-256-GCM
   - Cross-platform directory support
   - **Enterprise**: Supports hierarchical MDM integration

2. **OS Keyring** (`WithKeyringStore()`)
   - macOS: Keychain
   - Linux: Secret Service (GNOME/KDE)
   - Windows: Credential Manager

3. **In-Memory** (`WithInMemoryStore()`)
   - Temporary storage for testing
   - Data lost when application exits

**Note**: MDM support is integrated directly into the file store when using system directories with platform-aware configuration.

### Storage Configuration Examples

```go
// Pattern 1: Platform-aware file storage (Recommended for multi-product companies)
plat, _ := platform.NewPlatform("com.company", "myapp", runtime.GOOS)
configDir := plat.UserAppDataDirectory() // OS-native directory with publisher
profiler, err := profiles.New("myapp", profiles.WithFileStore(configDir))

// Pattern 2: Simple keyring storage (Good for single-product apps)
profiler, err := profiles.New("com.company.myapp", profiles.WithKeyringStore())

// Pattern 3: Custom directory override
profiler, err := profiles.New("myapp", profiles.WithFileStore("/custom/path"))

// Pattern 4: In-memory storage (testing only)
profiler, err := profiles.New("myapp", profiles.WithInMemoryStore())

// Pattern 5: Enterprise with MDM + Platform directories
plat, _ := platform.NewPlatform("com.company", "myapp", runtime.GOOS)
configDir := plat.SystemAppDataDirectory() // System-wide for enterprise
profiler, err := profiles.New("com.company.myapp", 
    profiles.WithFileStore(configDir),
    profiles.WithMDMConfig(profiles.MDMConfig{})) // Auto-populated from name
```

## 🏗️ Architecture

### Profile Interface

All profiles must implement the `NamedProfile` interface:

```go
type NamedProfile interface {
    GetName() string
}
```

### Generic Type Safety

The library uses Go generics for type-safe operations:

```go
// Type-safe profile retrieval
profile, err := profiles.GetProfile[*MyProfile](profiler, "profile-name")

// Type-safe profile loading
store, err := profiles.LoadProfileStore[*MyProfile]("namespace", "profile-name")
```

### Cross-Platform Directories

The library automatically uses OS-appropriate directories:

| OS | User Data | User Config | System Data | System Config |
|----|-----------|-------------|-------------|---------------|
| **macOS** | `~/Library/Application Support/` | `~/Library/Application Support/` | `/Library/Application Support/` | `/Library/Application Support/` |
| **Linux** | `~/.local/share/` | `~/.config/` | `/usr/local/` | `/etc/` |
| **Windows** | `%LOCALAPPDATA%` | `%LOCALAPPDATA%` | `%PROGRAMDATA%` | `%PROGRAMFILES%` |

## 📋 API Reference

### Core Functions

```go
// Create new profiler
func New(configName string, opts ...ProfileConfigVariadicFunc) (*Profiler, error)

// Add profile 
func (p *Profiler) AddProfile(profile NamedProfile, setDefault bool) error

// Get profile
func GetProfile[T NamedProfile](p *Profiler, name string) (*ProfileStore, error)

// List profiles
func ListProfiles(p *Profiler) []string

// Use default profile
func UseDefaultProfile[T NamedProfile](p *Profiler) (*ProfileStore, error)

// Delete profile
func DeleteProfile[T NamedProfile](p *Profiler, name string) error
```

### Configuration Options

```go
// Storage backends
func WithFileStore(storeDir string) ProfileConfigVariadicFunc
func WithKeyringStore() ProfileConfigVariadicFunc  
func WithInMemoryStore() ProfileConfigVariadicFunc

// MDM configuration
func WithMDMConfig(config MDMConfig) ProfileConfigVariadicFunc

// MDM configuration struct
type MDMConfig struct {
    ReverseDNS string `json:"reverseDNS,omitempty"`
    // Future extensible fields
}
```

## 🧪 Testing

Run the test suite:

```bash
# Run all tests
go test ./...

# Run tests without integration tests
go test ./... -short

# Run with verbose output
go test ./... -v
```

## 🤝 Contributing

This project was born out of [OpenTDF](https://github.com/opentdf/platform) and [otdfctl](https://github.com/opentdf/otdfctl).

### Development Setup

1. Clone the repository
2. Install dependencies: `go mod tidy`
3. Run tests: `go test ./...`

### Future Roadmap

- [ ] Enhanced key/value store abstraction
- [ ] Additional storage backends
- [ ] Windows Group Policy integration
- [ ] Configuration validation schemas

## 📄 License

[Add your license information here]

## 🆘 Support

[Add support/contact information here]
