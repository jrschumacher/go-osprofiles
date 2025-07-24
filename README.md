# go-osprofiles

A Go library for creating and managing application profiles native to operating systems. Provides a unified interface for storing/retrieving application configuration profiles across different storage backends with cross-platform support and enterprise MDM integration.

## 🚀 Features

- **Cross-platform storage** - macOS, Linux, and Windows with native directory structures
- **Multiple storage backends** - Keyring (encrypted), filesystem, and in-memory
- **Enterprise MDM support** - macOS managed preferences with automatic fallbacks
- **Type-safe profiles** - Generic interface for strongly-typed profile management
- **Encryption by default** - AES-256-GCM encryption with system keyring integration

## 📦 Installation

```bash
go get github.com/jrschumacher/go-osprofiles
```

## 🔧 Quick Start

### Basic User Preferences

```go
package main

import (
    "log"
    osprofiles "github.com/jrschumacher/go-osprofiles"
)

type UserSettings struct {
    Name     string `json:"name"`
    Theme    string `json:"theme"`
    Language string `json:"language"`
}

func (u UserSettings) GetName() string { return "user-settings" }

func main() {
    // Create profiler for user preferences (encrypted by default)
    profiler, err := osprofiles.New("com.mycompany.myapp")
    if err != nil {
        log.Fatal(err)
    }

    // Save user settings
    settings := UserSettings{
        Name: "user-settings",
        Theme: "dark", 
        Language: "en",
    }
    
    err = profiler.AddProfile(settings, true)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Println("Settings saved securely with encryption")
}
```

## 📚 Examples & Use Cases

The [**examples/**](./examples/) directory contains practical, copy-paste ready examples for common scenarios:

### 🎯 **[Simple User Space](./examples/simple-user-space/)**
Personal user preferences and settings with automatic encryption
- User directory storage (`~/Library/Application Support/`, `%APPDATA%`, `~/.local/share/`)
- AES-256 encryption with system keyring
- Perfect for: User preferences, themes, personal API keys

### 🏢 **[Simple System App Data](./examples/simple-system-app-data/)**
System-wide application data shared across users
- System directory storage (`/Library/Application Support/`, `%ProgramData%`, `/usr/local/share/`)
- Requires appropriate permissions
- Perfect for: Application databases, shared configuration, system-wide settings

### ⚙️ **[Simple System App Config](./examples/simple-system-app-config/)**
Read-only system configuration with graceful defaults
- Loads from system directories with fallback to defaults
- Typically managed by installers or administrators
- Perfect for: Server endpoints, system policies, deployment configuration

### 🔄 **[User + System Config](./examples/user-plus-system-config/)**
Hybrid configuration combining system policies with user preferences
- **System policies**: Corporate/admin settings (read-only)  
- **User preferences**: Personal customization (read-write)
- **Merged configuration**: Intelligent precedence handling
- Perfect for: Enterprise applications with user customization

### 🍎 **[macOS MDM with Fallbacks](./examples/macos-mdm-with-fallbacks/)**
Enterprise macOS deployment with MDM managed preferences
- **MDM plist support**: Reads from `/Library/Managed Preferences/`
- **Automatic fallbacks**: MDM → System files → Application defaults
- **Read-only enforcement**: Prevents user modification of MDM settings
- Perfect for: Corporate macOS deployments, enterprise policy management

## 🏗️ Architecture

### Storage Backends

```go
// Default: Encrypted files with system keyring (recommended)
profiler, _ := osprofiles.New("com.mycompany.myapp")

// Unencrypted files (for development/testing)
profiler, _ := osprofiles.New("com.mycompany.myapp", osprofiles.WithFileStore("/custom/path"))

// In-memory only (for unit tests)
profiler, _ := osprofiles.New("com.mycompany.myapp", osprofiles.WithInMemoryStore())
```

### Cross-Platform Directory Support

| Platform | User Directory | System Directory |
|----------|----------------|------------------|
| **macOS** | `~/Library/Application Support/` | `/Library/Application Support/` |
| **Windows** | `%APPDATA%` | `%ProgramData%` |
| **Linux** | `~/.local/share/` | `/usr/local/share/` |

### Enterprise MDM Integration (macOS)

The library automatically detects and reads MDM managed preferences:

1. **MDM plist files** - `/Library/Managed Preferences/com.company.app.plist`
2. **JSON in plist** - Automatic parsing of JSON strings in plist files  
3. **Read-only enforcement** - Users cannot modify MDM-managed settings
4. **Graceful fallbacks** - MDM → System files → Application defaults

## 🧪 Testing

Set `OSPROFILES_TEST_BASE_PATH` environment variable to override system paths for testing:

```bash
export OSPROFILES_TEST_BASE_PATH=/tmp/osprofiles-test
go run main.go
```

## 🔐 Security

- **Encryption**: AES-256-GCM encryption for file storage
- **Key management**: Uses system keyring for encryption keys
- **MDM enforcement**: Corporate policies cannot be modified by users
- **Permission handling**: Graceful handling of permission issues

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Run tests (`go test ./...`)
4. Commit changes (`git commit -m 'Add amazing feature'`)
5. Push to branch (`git push origin feature/amazing-feature`)
6. Open a Pull Request

---

**💡 Tip**: Start with the [examples](./examples/) to see practical implementations for your use case!