# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

- **Build**: `go build ./...` - Compiles all packages in the module
- **Test**: `go test ./...` - Runs all tests across the codebase
- **Test specific package**: `go test ./pkg/platform` or `go test ./pkg/store`
- **Format code**: `go fmt ./...` - Formats Go source files
- **Vet code**: `go vet ./...` - Examines Go source code for suspicious constructs

## Architecture Overview

This is a Go library for managing application profiles native to different operating systems. The library provides a unified interface for storing and retrieving application configuration data with field-level security classification and hybrid storage backends.

### Core Components

**Profile Management (`profile.go`, `profileConfig.go`)**
- `Profiler` struct is the main entry point for managing profiles
- Default storage backend is now **Hybrid Security Storage** (with fallback support for legacy backends)
- Provides profile CRUD operations with validation and temporary field overrides
- Support for temporary data overlays that don't persist to storage

**Hybrid Security Storage System (`pkg/store/storeHybrid.go`, `pkg/store/security.go`)**
- **Field-level security classification** using struct tags: `security:"secure|plaintext|temporary"`
- **Secure fields**: Encrypted and stored in files with keys managed in OS keyring
- **Plaintext fields**: Stored unencrypted in JSON files for non-sensitive data
- **Temporary fields**: Kept in memory only, never persisted
- Automatic keyring availability detection with graceful fallback to plaintext + warnings
- Security validation and compliance checking

**Versioning System (`versions.go`)**
- **go-osprofiles version**: Library version tracking for compatibility
- **App version**: Application-defined version for custom migration logic
- **Profile format version**: Format evolution tracking (v0 legacy → v1 file → v2 hybrid)
- Automatic migration path detection and backward compatibility

**Migration Utilities (`migration.go`)**
- `MigrateProfiles()`: Migrates existing profiles from legacy storage to hybrid storage
- Version detection and compatibility checking
- Backup creation before migration
- Dry-run support for validation
- Comprehensive migration reporting

**Platform Abstraction (`pkg/platform/`)**
- Cross-platform interface for OS-specific directory paths
- Implementations for Darwin (`darwin.go`), Linux (`linux.go`), and Windows (`windows.go`)
- Handles user home directories, app data directories, and config directories

**Legacy Storage Backends (`pkg/store/`)**
- Keyring storage (`storeKeyring.go`): OS keyring integration
- File storage (`storeFileSystem.go`): Encrypted file storage with metadata
- Memory storage (`storeMemory.go`): In-memory storage for testing
- All legacy backends maintained for backward compatibility

**Global Configuration (`internal/global/`)**
- Manages application-wide settings and default profiles
- Version tracking for global configuration migration

### Security Classification

Use struct tags to classify field security levels:

```go
type MyProfile struct {
    Name     string `json:"name" security:"plaintext"`      // Stored unencrypted
    APIKey   string `json:"api_key" security:"secure"`      // Encrypted + keyring key
    Endpoint string `json:"endpoint" security:"plaintext"`  // Stored unencrypted
    TempToken string `json:"temp_token" security:"temporary"` // Memory only
}
```

### Configuration Examples

**Hybrid Storage (Recommended)**:
```go
profiler, err := New("my-app", 
    WithHybridStore("./profiles", "1.0.0"), // directory, app version
    WithSecurityMode(SecurityModeKeyring),   // or SecurityModeInsecure
)
```

**Legacy Storage (For Migration)**:
```go
profiler, err := New("my-app", WithFileStore("./profiles"))
profiler, err := New("my-app", WithKeyringStore())
```

### Migration Workflow

```go
// 1. Validate migration
report, err := ValidateMigration[MyProfile]("my-app", global.PROFILE_DRIVER_FILE)

// 2. Perform migration with backup
report, err := MigrateProfiles[MyProfile]("my-app", global.PROFILE_DRIVER_FILE)

// 3. Check migration results
fmt.Printf("Migrated %d profiles\n", report.ProfilesMigrated)
```

### Key Patterns

- **Security-first design**: Field-level encryption with automatic keyring fallback
- **Factory pattern**: Platform and store creation based on OS/driver type
- **Functional options**: Configuration via variadic functions
- **Generic constraints**: Type-safe profile operations
- **Layered data**: Persistent storage + temporary overlays
- **Version-aware**: Automatic migration and compatibility checking

### Module Dependencies

- Uses `github.com/zalando/go-keyring` for OS keyring integration
- Test framework: `github.com/stretchr/testify`
- Supports Go 1.23.4+

### Important Security Notes

- **Windows keyring limitations**: Hybrid storage solves Windows keyring size limits by storing only encryption keys in keyring
- **Graceful degradation**: Automatic fallback to plaintext with warnings when keyring unavailable
- **Field validation**: Security compliance checking detects sensitive data in plaintext fields
- **Temporary overrides**: Flag-based field overrides for runtime configuration without persistence