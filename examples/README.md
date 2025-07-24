# Go-OSProfiles Examples

This directory contains practical, copy-paste ready examples for common use cases of the go-osprofiles library.

## Examples Overview

### 1. [Simple User Space](./simple-user-space/) 
**Use case**: Personal user preferences and settings
- Stores data in user's personal directory
- Automatic encryption with system keyring
- Cross-platform user directory support
- Shows storage backend options (keyring, filesystem, memory)

### 2. [Simple System App Data](./simple-system-app-data/)
**Use case**: System-wide application data
- Stores data in system application directory  
- Shared across all users on the system
- Requires appropriate permissions to write
- Cross-platform system directory support

### 3. [Simple System App Config](./simple-system-app-config/)
**Use case**: Read-only system configuration
- Loads configuration from system directory
- Typically managed by installers or administrators
- Graceful fallback to defaults when no config exists
- Shows error handling patterns

### 4. [User + System Config](./user-plus-system-config/)
**Use case**: Hybrid configuration (very common pattern)
- **System policies**: Read-only corporate/admin settings
- **User preferences**: Personal customization options  
- **Merged configuration**: Combines both sources intelligently
- Shows dual profiler pattern with proper precedence

### 5. [macOS MDM with Fallbacks](./macos-mdm-with-fallbacks/)
**Use case**: Enterprise macOS deployment with MDM
- **MDM plist support**: Reads from `/Library/Managed Preferences/`
- **Automatic fallbacks**: MDM → System files → Application defaults
- **Read-only enforcement**: Prevents user modification of MDM settings
- **JSON in plist**: Automatic parsing of JSON strings in plist files

## Running Examples

Each example is self-contained with its own `go.mod`:

```bash
cd simple-user-space
go run main.go
```

## Cross-Platform Notes

All examples work across platforms with appropriate directory mappings:

- **macOS**: `~/Library/Application Support/`, `/Library/Application Support/`
- **Windows**: `%APPDATA%`, `%ProgramData%`  
- **Linux**: `~/.local/share/`, `/usr/local/share/`

## Storage Options

Examples demonstrate different storage backends:

- **Default**: Encrypted files with system keyring (recommended)
- **File Store**: Unencrypted files (for testing/development)
- **Memory Store**: In-memory only (for unit tests)

## Error Handling

Each example includes:
- Proper error checking
- Graceful fallbacks when configurations don't exist
- Platform-specific considerations
- Basic logging for debugging

## Best Practices Shown

- **Configuration layering**: System policies + user preferences
- **Security**: Encryption for sensitive data, read-only MDM configs
- **Cross-platform**: Same code works on all platforms
- **Fallback strategies**: Graceful degradation when configs missing
- **Enterprise deployment**: MDM support for corporate environments