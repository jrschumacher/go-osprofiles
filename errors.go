package profiles

import "errors"

var (
	ErrProfileNameConflict    = errors.New("error: profile name already exists in storage")
	ErrMissingCurrentProfile  = errors.New("error: current profile not set")
	ErrMissingDefaultProfile  = errors.New("error: default profile not set")
	ErrMissingProfileName     = errors.New("error: profile name not found")
	ErrInvalidStoreDriver     = errors.New("error: invalid store driver")
	// Managed configuration errors for downstream apps to handle
	ErrManagedByMDM          = errors.New("configuration is managed remotely (MDM) and cannot be modified")
	ErrManagedBySystem       = errors.New("configuration is managed by system and cannot be modified")
	ErrReadOnlyLocation      = errors.New("configuration location is read-only")
)
