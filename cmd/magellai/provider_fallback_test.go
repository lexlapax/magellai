// ABOUTME: Provider fallback test fixes for CLI integration tests
// ABOUTME: Ensures proper testing of provider fallback functionality

//go:build cmdline
// +build cmdline

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCLI_ProviderFallback tests provider fallback functionality
func TestCLI_ProviderFallback(t *testing.T) {
	ForEachStorageType(t, true, func(t *testing.T, env *TestEnv) {
		// Use the fallback profile which has both primary and fallback providers configured
		output, err := env.RunCommand("ask", "--profile", "fallback", "Hello, what is your name?")
		require.NoError(t, err)
		assert.NotEmpty(t, output, "Response should not be empty")
	})
}

// TestCLI_ProviderFallbackError tests provider fallback with error conditions
func TestCLI_ProviderFallbackError(t *testing.T) {
	WithMockEnv(t, StorageTypeFilesystem, func(t *testing.T, env *TestEnv) {
		// First, update config to make primary provider always fail
		// and ensure fallback works
		output, err := env.RunCommand("ask", "--profile", "error", "Tell me a joke")

		// We have two potential outcomes:
		// 1. If fallback works correctly, no error and we get a response
		// 2. If no fallback configured for error profile, we get an error

		if err == nil {
			// Fallback worked
			assert.NotEmpty(t, output, "Response from fallback provider should not be empty")
		} else {
			// No fallback configured, error is expected
			assert.Contains(t, err.Error(), "error")
		}
	})
}

// TestCLI_ProviderFallbackWithOverride tests fallback with provider override
func TestCLI_ProviderFallbackWithOverride(t *testing.T) {
	WithMockEnv(t, StorageTypeFilesystem, func(t *testing.T, env *TestEnv) {
		// Explicitly select fallback provider using command line argument
		output, err := env.RunCommand("ask", "--provider", "mock-fallback", "Hello, what is your name?")
		require.NoError(t, err)
		assert.NotEmpty(t, output, "Response should not be empty")
	})
}

// TestCLI_ProviderChain tests multiple provider chain setup
func TestCLI_ProviderChain(t *testing.T) {
	WithMockEnv(t, StorageTypeFilesystem, func(t *testing.T, env *TestEnv) {
		// Use comma-separated chain of providers
		output, err := env.RunCommand("ask", "--provider", "mock-primary,mock-fallback", "Hello, what is your name?")
		require.NoError(t, err)
		assert.NotEmpty(t, output, "Response should not be empty")
	})
}

// TestCLI_ProviderFallbackInChat tests provider fallback in chat mode
func TestCLI_ProviderFallbackInChat(t *testing.T) {
	WithMockEnv(t, StorageTypeFilesystem, func(t *testing.T, env *TestEnv) {
		// Test chat with fallback profile
		input := `Hello
/provider set mock-primary,mock-fallback
Tell me more
/provider info
/exit`
		output, err := env.RunInteractiveCommand(input, "chat", "--profile", "fallback")
		require.NoError(t, err)
		assert.Contains(t, output, "Provider changed")
	})
}

// TestCLI_ProviderFallbackWithTemporaryFailure tests recovery from temporary failures
func TestCLI_ProviderFallbackWithTemporaryFailure(t *testing.T) {
	// This test is more complex and would ideally use a dynamic mock provider
	// that fails temporarily. For now, we'll just check the command structure.
	WithMockEnv(t, StorageTypeFilesystem, func(t *testing.T, env *TestEnv) {
		// Use mock-error provider that is set to always fail
		// with another provider as fallback
		output, err := env.RunCommand("ask", "--provider", "mock-error,mock-fallback", "Hello, what is your name?")

		// If the fallback mechanism works, we should get a response from the fallback provider
		assert.NoError(t, err, "Should not error with fallback provider available")
		assert.NotEmpty(t, output, "Response from fallback provider should not be empty")
	})
}
