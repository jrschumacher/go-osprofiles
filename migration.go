package profiles

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jrschumacher/go-osprofiles/internal/global"
	"github.com/jrschumacher/go-osprofiles/pkg/store"
)

// MigrationReport contains information about the migration process
type MigrationReport struct {
	ProfilesMigrated    int                        `json:"profiles_migrated"`
	ProfilesSkipped     int                        `json:"profiles_skipped"`
	MigrationErrors     []string                   `json:"migration_errors,omitempty"`
	MigrationWarnings   []string                   `json:"migration_warnings,omitempty"`
	BackupCreated       bool                       `json:"backup_created"`
	BackupLocation      string                     `json:"backup_location,omitempty"`
	VersionsDetected    map[string]ProfileVersionInfo `json:"versions_detected,omitempty"`
}

// ProfileVersionInfo contains version information for a migrated profile
type ProfileVersionInfo struct {
	ProfileName          string `json:"profile_name"`
	OldFormatVersion     string `json:"old_format_version"`
	NewFormatVersion     string `json:"new_format_version"`
	GoOSProfilesVersion  string `json:"go_osprofiles_version"`
	AppVersion          string `json:"app_version,omitempty"`
}

// MigrationOptions configures the migration process
type MigrationOptions struct {
	// DryRun performs migration validation without making changes
	DryRun bool
	// CreateBackup creates a backup of existing data before migration
	CreateBackup bool
	// BackupDirectory specifies where to create backups (optional)
	BackupDirectory string
	// SkipSecurityValidation bypasses security classification validation
	SkipSecurityValidation bool
	// ForceMigration proceeds even if there are validation warnings
	ForceMigration bool
}

// MigrateProfiles migrates existing profiles from old storage backends to hybrid storage
func MigrateProfiles[T NamedProfile](
	configName string,
	oldDriver global.ProfileDriver,
	driverOpts ...store.DriverOpt,
) (*MigrationReport, error) {
	return MigrateProfilesWithOptions[T](configName, oldDriver, MigrationOptions{
		CreateBackup: true,
		DryRun:       false,
	}, driverOpts...)
}

// MigrateProfilesWithOptions migrates profiles with custom migration options
func MigrateProfilesWithOptions[T NamedProfile](
	configName string,
	oldDriver global.ProfileDriver,
	options MigrationOptions,
	driverOpts ...store.DriverOpt,
) (*MigrationReport, error) {
	report := &MigrationReport{
		MigrationErrors:   make([]string, 0),
		MigrationWarnings: make([]string, 0),
		VersionsDetected:  make(map[string]ProfileVersionInfo),
	}

	// Validate inputs
	if configName == "" {
		return report, fmt.Errorf("%w: configName cannot be empty", ErrInvalidProfile)
	}

	// Create old store factory
	oldStoreFactory := newStoreFactory(oldDriver)
	if oldStoreFactory == nil {
		return report, fmt.Errorf("%w: invalid old driver %v", ErrInvalidStoreDriver, oldDriver)
	}

	// Load existing global configuration
	globalStore, err := global.LoadGlobalConfig(configName, oldStoreFactory, driverOpts...)
	if err != nil {
		return report, fmt.Errorf("failed to load global config: %w", err)
	}

	profileNames := globalStore.ListProfiles()
	if len(profileNames) == 0 {
		report.MigrationWarnings = append(report.MigrationWarnings, "No profiles found to migrate")
		return report, nil
	}

	// Create backup if requested
	if options.CreateBackup && !options.DryRun {
		backupLocation, err := createMigrationBackup(configName, globalStore, oldStoreFactory, profileNames, options.BackupDirectory)
		if err != nil {
			report.MigrationErrors = append(report.MigrationErrors, fmt.Sprintf("Failed to create backup: %v", err))
			if !options.ForceMigration {
				return report, fmt.Errorf("backup creation failed: %w", err)
			}
		} else {
			report.BackupCreated = true
			report.BackupLocation = backupLocation
		}
	}

	// Migrate each profile
	for _, profileName := range profileNames {
		// Detect version information for this profile
		versionInfo, err := DetectProfileVersion(configName, profileName, oldDriver)
		if err != nil {
			report.MigrationWarnings = append(report.MigrationWarnings,
				fmt.Sprintf("Warning: could not detect version for profile '%s': %v", profileName, err))
		} else {
			report.VersionsDetected[profileName] = versionInfo
		}
		
		if err := migrateProfile[T](configName, profileName, oldStoreFactory, options, report); err != nil {
			errorMsg := fmt.Sprintf("Failed to migrate profile '%s': %v", profileName, err)
			report.MigrationErrors = append(report.MigrationErrors, errorMsg)
			
			if !options.ForceMigration {
				return report, fmt.Errorf("migration failed for profile '%s': %w", profileName, err)
			}
		} else {
			report.ProfilesMigrated++
		}
	}

	// Migrate global configuration to hybrid storage
	if !options.DryRun {
		newGlobalStore, err := global.LoadGlobalConfig(configName, store.NewHybridStore, driverOpts...)
		if err != nil {
			report.MigrationErrors = append(report.MigrationErrors, fmt.Sprintf("Failed to create new global store: %v", err))
			return report, fmt.Errorf("failed to create new global store: %w", err)
		}

		// Copy profile list and default profile
		for _, profileName := range profileNames {
			if err := newGlobalStore.AddProfile(profileName); err != nil {
				report.MigrationWarnings = append(report.MigrationWarnings, 
					fmt.Sprintf("Warning: failed to add profile '%s' to global config: %v", profileName, err))
			}
		}

		defaultProfile := globalStore.GetDefaultProfile()
		if defaultProfile != "" {
			if err := newGlobalStore.SetDefaultProfile(defaultProfile); err != nil {
				report.MigrationWarnings = append(report.MigrationWarnings,
					fmt.Sprintf("Warning: failed to set default profile '%s': %v", defaultProfile, err))
			}
		}
	}

	return report, nil
}

// migrateProfile migrates a single profile from old storage to hybrid storage
func migrateProfile[T NamedProfile](
	configName string,
	profileName string,
	oldStoreFactory store.NewStoreInterface,
	options MigrationOptions,
	report *MigrationReport,
) error {
	// Load profile from old storage
	oldProfileStore, err := LoadProfileStore[T](configName, oldStoreFactory, profileName)
	if err != nil {
		return fmt.Errorf("failed to load profile from old storage: %w", err)
	}

	profile := oldProfileStore.Profile

	// Validate security classification if not skipped
	if !options.SkipSecurityValidation {
		if err := store.ValidateProfileSecurity(profile); err != nil {
			warningMsg := fmt.Sprintf("Security validation warning for profile '%s': %v", profileName, err)
			report.MigrationWarnings = append(report.MigrationWarnings, warningMsg)
			
			if !options.ForceMigration {
				return fmt.Errorf("security validation failed: %w", err)
			}
		}

		// Generate security report
		securityReport, err := store.GenerateSecurityReport(profile)
		if err != nil {
			report.MigrationWarnings = append(report.MigrationWarnings,
				fmt.Sprintf("Failed to generate security report for profile '%s': %v", profileName, err))
		} else if len(securityReport.SecurityWarnings) > 0 {
			for _, warning := range securityReport.SecurityWarnings {
				report.MigrationWarnings = append(report.MigrationWarnings,
					fmt.Sprintf("Profile '%s': %s", profileName, warning))
			}
		}
	}

	if options.DryRun {
		// In dry run mode, just validate that we can create the new profile store
		_, err := NewProfileStore(configName, store.NewHybridStore, profile)
		if err != nil {
			return fmt.Errorf("dry run validation failed: %w", err)
		}
		return nil
	}

	// Create new profile store with hybrid storage
	newProfileStore, err := NewProfileStore(configName, store.NewHybridStore, profile)
	if err != nil {
		return fmt.Errorf("failed to create new profile store: %w", err)
	}

	// Save profile to new storage
	if err := newProfileStore.Save(); err != nil {
		return fmt.Errorf("failed to save profile to new storage: %w", err)
	}

	return nil
}

// createMigrationBackup creates a backup of existing profile data
func createMigrationBackup(
	configName string,
	globalStore *global.GlobalStore,
	oldStoreFactory store.NewStoreInterface,
	profileNames []string,
	backupDir string,
) (string, error) {
	if backupDir == "" {
		execPath, err := os.Executable()
		if err != nil {
			return "", fmt.Errorf("unable to determine executable path: %w", err)
		}
		execDir := filepath.Dir(execPath)
		backupDir = filepath.Join(execDir, "profile-backups")
	}

	// Create timestamped backup directory
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	backupPath := filepath.Join(backupDir, fmt.Sprintf("backup-%s-%s", configName, timestamp))
	
	if err := os.MkdirAll(backupPath, 0o700); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Backup global configuration
	globalData, err := globalStore.GetRawData()
	if err != nil {
		return "", fmt.Errorf("failed to get global data: %w", err)
	}

	globalBackupPath := filepath.Join(backupPath, "global.json")
	if err := os.WriteFile(globalBackupPath, globalData, 0o600); err != nil {
		return "", fmt.Errorf("failed to backup global config: %w", err)
	}

	// Backup each profile
	for _, profileName := range profileNames {
		profileStore, err := oldStoreFactory(configName, getStoreKey(profileName))
		if err != nil {
			continue // Skip profiles that can't be loaded
		}

		profileData, err := profileStore.Get()
		if err != nil {
			continue // Skip profiles that can't be read
		}

		profileBackupPath := filepath.Join(backupPath, fmt.Sprintf("profile-%s.json", profileName))
		if err := os.WriteFile(profileBackupPath, profileData, 0o600); err != nil {
			return "", fmt.Errorf("failed to backup profile '%s': %w", profileName, err)
		}
	}

	return backupPath, nil
}

// ValidateMigration validates that a migration can be performed safely
func ValidateMigration[T NamedProfile](
	configName string,
	oldDriver global.ProfileDriver,
	driverOpts ...store.DriverOpt,
) (*MigrationReport, error) {
	options := MigrationOptions{
		DryRun:                 true,
		CreateBackup:           false,
		SkipSecurityValidation: false,
		ForceMigration:         false,
	}

	return MigrateProfilesWithOptions[T](configName, oldDriver, options, driverOpts...)
}

// RestoreFromBackup restores profiles from a backup created during migration
func RestoreFromBackup(backupPath string, configName string) error {
	// Verify backup exists
	if _, err := os.Stat(backupPath); err != nil {
		return fmt.Errorf("backup path does not exist: %w", err)
	}

	// This would implement the restore logic
	// For now, return a helpful error message
	return fmt.Errorf("restore functionality not yet implemented - manual restoration required from %s", backupPath)
}

// DetectProfileVersion attempts to detect the version of a profile
func DetectProfileVersion(
	configName string,
	profileName string,
	oldDriver global.ProfileDriver,
) (ProfileVersionInfo, error) {
	versionInfo := ProfileVersionInfo{
		ProfileName:          profileName,
		OldFormatVersion:     ProfileFormatVersionLegacy, // Default to v0 (legacy)
		NewFormatVersion:     ProfileFormatVersionCurrent,
		GoOSProfilesVersion:  GoOSProfilesVersionCurrent,
	}

	// For now, we'll use the driver type to infer version capability
	// In practice, you'd detect this by trying to read metadata files
	switch oldDriver {
	case global.PROFILE_DRIVER_FILE:
		// File store with metadata might have version info
		// For now, assume legacy format unless we can detect otherwise
		versionInfo.OldFormatVersion = ProfileFormatVersionV1
	case global.PROFILE_DRIVER_KEYRING, global.PROFILE_DRIVER_IN_MEMORY:
		// These don't have metadata files, so assume legacy
		versionInfo.OldFormatVersion = ProfileFormatVersionLegacy
	case global.PROFILE_DRIVER_CUSTOM:
		// Custom drivers might have version info, but we can't know for sure
		versionInfo.OldFormatVersion = ProfileFormatVersionLegacy
	}
	// Note: Hybrid store would be detected as needing no migration
	// since it's already the current format

	return versionInfo, nil
}

// loadFileStoreMetadata loads metadata from a file store .nfo file
// This is a placeholder for future implementation
func loadFileStoreMetadata(configName, profileName string) (map[string]interface{}, error) {
	// This is a simplified version - in practice we'd need to construct the full path
	// using the same logic as the file store and parse the .nfo file
	return nil, fmt.Errorf("metadata loading not yet implemented")
}

// loadHybridStoreMetadata loads metadata from a hybrid store metadata.json file
// This is a placeholder for future implementation
func loadHybridStoreMetadata(configName, profileName string) (map[string]interface{}, error) {
	// This is a simplified version - in practice we'd need to construct the full path
	// using the same logic as the hybrid store and parse the metadata.json file
	return nil, fmt.Errorf("metadata loading not yet implemented")
}

// GetVersionCompatibilityInfo returns information about version compatibility
func GetVersionCompatibilityInfo(profileFormatVersion string) VersionCompatibilityInfo {
	return VersionCompatibilityInfo{
		IsCompatible:     IsVersionCompatible(profileFormatVersion),
		RequiresMigration: RequiresMigration(profileFormatVersion),
		MigrationPath:    GetMigrationPath(profileFormatVersion),
	}
}

// VersionCompatibilityInfo contains information about version compatibility
type VersionCompatibilityInfo struct {
	IsCompatible     bool     `json:"is_compatible"`
	RequiresMigration bool     `json:"requires_migration"`
	MigrationPath    []string `json:"migration_path,omitempty"`
}