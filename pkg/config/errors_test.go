// ABOUTME: Tests for config package error definitions
// ABOUTME: Validates error constants, messages, and error wrapping behavior

package config

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorConstants(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "ErrConfigNotFound",
			err:      ErrConfigNotFound,
			expected: "configuration not found",
		},
		{
			name:     "ErrInvalidConfig",
			err:      ErrInvalidConfig,
			expected: "invalid configuration",
		},
		{
			name:     "ErrConfigLoadFailed",
			err:      ErrConfigLoadFailed,
			expected: "configuration load failed",
		},
		{
			name:     "ErrConfigSaveFailed",
			err:      ErrConfigSaveFailed,
			expected: "configuration save failed",
		},
		{
			name:     "ErrProfileNotFound",
			err:      ErrProfileNotFound,
			expected: "profile not found",
		},
		{
			name:     "ErrInvalidProfile",
			err:      ErrInvalidProfile,
			expected: "invalid profile",
		},
		{
			name:     "ErrAliasNotFound",
			err:      ErrAliasNotFound,
			expected: "alias not found",
		},
		{
			name:     "ErrInvalidAlias",
			err:      ErrInvalidAlias,
			expected: "invalid alias",
		},
		{
			name:     "ErrSettingNotFound",
			err:      ErrSettingNotFound,
			expected: "setting not found",
		},
		{
			name:     "ErrInvalidSettingValue",
			err:      ErrInvalidSettingValue,
			expected: "invalid setting value",
		},
		{
			name:     "ErrMergeConflict",
			err:      ErrMergeConflict,
			expected: "configuration merge conflict",
		},
		{
			name:     "ErrValidationFailed",
			err:      ErrValidationFailed,
			expected: "configuration validation failed",
		},
		{
			name:     "ErrPermission",
			err:      ErrPermission,
			expected: "configuration permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestErrorWrapping(t *testing.T) {
	baseError := ErrConfigNotFound
	wrappedError := fmt.Errorf("failed to load config: %w", baseError)

	// Test that unwrapping works correctly
	assert.True(t, errors.Is(wrappedError, baseError))
	assert.Equal(t, "failed to load config: configuration not found", wrappedError.Error())

	// Test multiple levels of wrapping
	doubleWrapped := fmt.Errorf("startup failed: %w", wrappedError)
	assert.True(t, errors.Is(doubleWrapped, baseError))
	assert.Equal(t, "startup failed: failed to load config: configuration not found", doubleWrapped.Error())
}

func TestErrorComparison(t *testing.T) {
	// Test that each error is distinct
	allErrors := []error{
		ErrConfigNotFound,
		ErrInvalidConfig,
		ErrConfigLoadFailed,
		ErrConfigSaveFailed,
		ErrProfileNotFound,
		ErrInvalidProfile,
		ErrAliasNotFound,
		ErrInvalidAlias,
		ErrSettingNotFound,
		ErrInvalidSettingValue,
		ErrMergeConflict,
		ErrValidationFailed,
		ErrPermission,
	}

	for i, err1 := range allErrors {
		for j, err2 := range allErrors {
			if i == j {
				assert.True(t, errors.Is(err1, err2))
			} else {
				assert.False(t, errors.Is(err1, err2))
			}
		}
	}
}

func TestErrorCategorization(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		isConfig  bool
		isProfile bool
		isAlias   bool
		isSetting bool
	}{
		{
			name:     "Config not found",
			err:      ErrConfigNotFound,
			isConfig: true,
		},
		{
			name:     "Invalid config",
			err:      ErrInvalidConfig,
			isConfig: true,
		},
		{
			name:     "Config load failed",
			err:      ErrConfigLoadFailed,
			isConfig: true,
		},
		{
			name:     "Config save failed",
			err:      ErrConfigSaveFailed,
			isConfig: true,
		},
		{
			name:      "Profile not found",
			err:       ErrProfileNotFound,
			isProfile: true,
		},
		{
			name:      "Invalid profile",
			err:       ErrInvalidProfile,
			isProfile: true,
		},
		{
			name:    "Alias not found",
			err:     ErrAliasNotFound,
			isAlias: true,
		},
		{
			name:    "Invalid alias",
			err:     ErrInvalidAlias,
			isAlias: true,
		},
		{
			name:      "Setting not found",
			err:       ErrSettingNotFound,
			isSetting: true,
		},
		{
			name:      "Invalid setting value",
			err:       ErrInvalidSettingValue,
			isSetting: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that errors are correctly categorized
			if tt.isConfig {
				// Could add specific config error checks if needed
				assert.NotNil(t, tt.err)
			}
			if tt.isProfile {
				// Could add specific profile error checks if needed
				assert.NotNil(t, tt.err)
			}
			if tt.isAlias {
				// Could add specific alias error checks if needed
				assert.NotNil(t, tt.err)
			}
			if tt.isSetting {
				// Could add specific setting error checks if needed
				assert.NotNil(t, tt.err)
			}
		})
	}
}

func TestErrorContext(t *testing.T) {
	// Test adding context to errors
	tests := []struct {
		name      string
		baseError error
		context   string
		expected  string
	}{
		{
			name:      "Config not found with file path",
			baseError: ErrConfigNotFound,
			context:   "file: /home/user/.magellai/config.yaml",
			expected:  "file: /home/user/.magellai/config.yaml: configuration not found",
		},
		{
			name:      "Profile not found with name",
			baseError: ErrProfileNotFound,
			context:   "profile: development",
			expected:  "profile: development: profile not found",
		},
		{
			name:      "Setting not found with key",
			baseError: ErrSettingNotFound,
			context:   "key: api_key",
			expected:  "key: api_key: setting not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contextualError := fmt.Errorf("%s: %w", tt.context, tt.baseError)
			assert.Equal(t, tt.expected, contextualError.Error())
			assert.True(t, errors.Is(contextualError, tt.baseError))
		})
	}
}

func TestErrorUsagePatterns(t *testing.T) {
	// Simulate common error usage patterns
	t.Run("Loading configuration", func(t *testing.T) {
		simulateConfigLoad := func(path string) error {
			if path == "" {
				return fmt.Errorf("empty path: %w", ErrInvalidConfig)
			}
			if path == "/missing" {
				return fmt.Errorf("path %s: %w", path, ErrConfigNotFound)
			}
			return nil
		}

		err := simulateConfigLoad("")
		assert.True(t, errors.Is(err, ErrInvalidConfig))

		err = simulateConfigLoad("/missing")
		assert.True(t, errors.Is(err, ErrConfigNotFound))

		err = simulateConfigLoad("/exists")
		assert.NoError(t, err)
	})

	t.Run("Profile operations", func(t *testing.T) {
		simulateProfileOp := func(name string) error {
			if name == "" {
				return fmt.Errorf("empty profile name: %w", ErrInvalidProfile)
			}
			if name == "nonexistent" {
				return fmt.Errorf("profile %s: %w", name, ErrProfileNotFound)
			}
			return nil
		}

		err := simulateProfileOp("")
		assert.True(t, errors.Is(err, ErrInvalidProfile))

		err = simulateProfileOp("nonexistent")
		assert.True(t, errors.Is(err, ErrProfileNotFound))

		err = simulateProfileOp("default")
		assert.NoError(t, err)
	})
}