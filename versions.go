package profiles

// Library version constants
const (
	// GoOSProfilesVersionV1 represents the first stable version
	GoOSProfilesVersionV1 = "1.0.0"
	// GoOSProfilesVersionCurrent is the current library version
	GoOSProfilesVersionCurrent = GoOSProfilesVersionV1
	
	// ProfileFormatVersionV1 represents the first profile format version
	ProfileFormatVersionV1 = "1.0"
	// ProfileFormatVersionV2 represents the hybrid security format
	ProfileFormatVersionV2 = "2.0"
	// ProfileFormatVersionCurrent is the current profile format version
	ProfileFormatVersionCurrent = ProfileFormatVersionV2
	
	// ProfileFormatVersionLegacy represents profiles without version info (assumes v0)
	ProfileFormatVersionLegacy = "0.0"
)

// VersionInfo contains version information for a profile
type VersionInfo struct {
	GoOSProfilesVersion  string `json:"go_osprofiles_version"`
	AppVersion          string `json:"app_version,omitempty"`
	ProfileFormatVersion string `json:"profile_format_version"`
}

// GetCurrentVersionInfo returns version info for new profiles
func GetCurrentVersionInfo(appVersion string) VersionInfo {
	return VersionInfo{
		GoOSProfilesVersion:  GoOSProfilesVersionCurrent,
		AppVersion:          appVersion,
		ProfileFormatVersion: ProfileFormatVersionCurrent,
	}
}

// IsVersionCompatible checks if a profile version is compatible with current library
func IsVersionCompatible(profileFormatVersion string) bool {
	switch profileFormatVersion {
	case ProfileFormatVersionLegacy, ProfileFormatVersionV1, ProfileFormatVersionV2:
		return true
	default:
		return false
	}
}

// RequiresMigration determines if a profile version requires migration
func RequiresMigration(profileFormatVersion string) bool {
	return profileFormatVersion != ProfileFormatVersionCurrent
}

// GetMigrationPath returns the migration steps needed for a profile version
func GetMigrationPath(fromVersion string) []string {
	switch fromVersion {
	case ProfileFormatVersionLegacy:
		return []string{ProfileFormatVersionV1, ProfileFormatVersionV2}
	case ProfileFormatVersionV1:
		return []string{ProfileFormatVersionV2}
	case ProfileFormatVersionV2:
		return []string{} // No migration needed
	default:
		return nil // Unknown version, cannot migrate
	}
}