// ABOUTME: Error definitions for the config package
// ABOUTME: Provides standard errors for configuration operations

package config

import "errors"

// Configuration-specific errors
var (
	// ErrConfigNotFound indicates the configuration was not found
	ErrConfigNotFound = errors.New("configuration not found")

	// ErrInvalidConfig indicates invalid configuration
	ErrInvalidConfig = errors.New("invalid configuration")

	// ErrConfigLoadFailed indicates configuration loading failed
	ErrConfigLoadFailed = errors.New("configuration load failed")

	// ErrConfigSaveFailed indicates configuration saving failed
	ErrConfigSaveFailed = errors.New("configuration save failed")

	// ErrProfileNotFound indicates the profile was not found
	ErrProfileNotFound = errors.New("profile not found")

	// ErrInvalidProfile indicates an invalid profile
	ErrInvalidProfile = errors.New("invalid profile")

	// ErrAliasNotFound indicates the alias was not found
	ErrAliasNotFound = errors.New("alias not found")

	// ErrInvalidAlias indicates an invalid alias
	ErrInvalidAlias = errors.New("invalid alias")

	// ErrSettingNotFound indicates the setting was not found
	ErrSettingNotFound = errors.New("setting not found")

	// ErrInvalidSettingValue indicates an invalid setting value
	ErrInvalidSettingValue = errors.New("invalid setting value")

	// ErrMergeConflict indicates a configuration merge conflict
	ErrMergeConflict = errors.New("configuration merge conflict")

	// ErrValidationFailed indicates configuration validation failed
	ErrValidationFailed = errors.New("configuration validation failed")

	// ErrPermission indicates a permission error accessing configuration
	ErrPermission = errors.New("configuration permission denied")
)
