# MDM JSON Plist Parsing

This example demonstrates the **powerful JSON parsing capabilities** for MDM plist files, including complex nested structures and automatic JSON string expansion.

## What This Shows

- 🧩 **Complex nested JSON** in plist strings
- 🔄 **Automatic JSON expansion** - no manual parsing needed
- 🎯 **Type-safe structs** - direct parsing to Go types
- 🏢 **Enterprise configuration** - database, services, security settings
- ⚙️ **Mixed content** - plist properties + JSON values

## Key Innovation

Traditional plist files are limited to basic types (strings, numbers, booleans, arrays, dictionaries). This library allows you to **embed entire JSON configurations as strings** that are automatically parsed into native Go data structures.

## Example Configuration

### Input: MDM Plist File
```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<string>{"database_config":{"primary":{"host":"prod-db.company.com","connection_pool":{"min_connections":5,"max_connections":25}}},"feature_flags":{"monitoring":true,"analytics":false}}</string>
</plist>
```

### Output: Parsed Go Struct
```go
config := ComplexConfig{
    DatabaseConfig: {
        Primary: {
            Host: "prod-db.company.com",
            ConnectionPool: {
                MinConnections: 5,
                MaxConnections: 25,
            },
        },
    },
    FeatureFlags: {
        "monitoring": true,
        "analytics": false,
    },
}
```

## Running the Example

```bash
cd examples/mdm-json-plist-parsing  
go run main.go
```

## Sample Output

```
🧩 MDM JSON Plist Parsing Example
=================================

📂 System Directory: /Library/Application Support/com.company/myapp
🔧 MDM Enabled: JSON strings will be automatically parsed
📍 Plist Path: /Library/Managed Preferences/com.company.myapp.plist

🔍 Reading complex JSON configuration from MDM plist...
📄 Parsed JSON structure:
```json
{
  "database_config": {
    "primary": {
      "host": "prod-db.company.com",
      "port": 5432,
      "connection_pool": {
        "min_connections": 5,
        "max_connections": 25,
        "timeout_seconds": 30
      },
      "ssl": {
        "enabled": true,
        "mode": "require"
      }
    },
    "read_replicas": [
      {"host": "replica1-db.company.com", "port": 5432, "weight": 100}
    ]
  },
  "feature_flags": {
    "authentication": true,
    "monitoring": true,
    "analytics": false
  }
}
```

✅ Complex configuration successfully parsed!

🗄️ Database Configuration:
   📍 Primary: prod-db.company.com:5432 (myapp_prod)
   🔗 Connection Pool: 5-25 connections (timeout: 30s)
   🔒 SSL: true (require)
   📚 Read Replicas: 2 replicas

🎯 Feature Flags:
   ✅ authentication
   ✅ monitoring  
   ❌ analytics

🛡️ Security Configuration:
   🔑 Encryption Key: base64-e...
   ⏱️ Rate Limit: 1000 req/min (burst: 50)
```

## Complex Data Structures Supported

- ✅ **Nested objects** - Deep object hierarchies
- ✅ **Arrays of objects** - Complex list structures  
- ✅ **Mixed types** - Strings, numbers, booleans, objects
- ✅ **Maps/dictionaries** - Dynamic key-value pairs
- ✅ **Optional fields** - Graceful handling of missing data
- ✅ **Type validation** - Automatic Go struct mapping

## Enterprise Benefits

1. **Rich Configuration**: Store complex database, service, and security settings
2. **Version Control**: JSON configurations can be easily diffed and versioned
3. **Validation**: Go struct tags provide automatic validation
4. **Maintainability**: Clear structure with strong typing
5. **IT Management**: Deploy via any MDM solution (Jamf, Intune, etc.)
6. **Fallback Safe**: Graceful handling if plist is missing or malformed

## Perfect For

- 🏢 **Complex enterprise applications** with rich configuration needs
- 🌐 **Microservice configurations** - databases, APIs, service discovery
- 🛡️ **Security settings** - encryption keys, rate limits, access controls
- 🎯 **Feature flag management** - enable/disable functionality enterprise-wide
- 📊 **Monitoring configuration** - logging levels, outputs, rotation policies
- 🔄 **Environment-specific settings** - dev/staging/production configurations

This approach gives you the **flexibility of JSON** with the **enterprise deployment capabilities of MDM plists**!