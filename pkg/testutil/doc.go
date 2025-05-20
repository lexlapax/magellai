// ABOUTME: Test utilities package for the Magellai project
// ABOUTME: Provides shared mocks, fixtures, and helper functions for testing

// Package testutil provides reusable test utilities for the Magellai project.
//
// The package is organized into three main subpackages:
//
// - mocks: Contains mock implementations of common interfaces (Provider, Backend, Command, etc.)
// - fixtures: Provides test data factories for creating consistent test objects
// - helpers: Offers utility functions for common testing tasks
//
// Example usage:
//
//	// Using mock backend
//	backend := mocks.NewMockBackend()
//	backend.SetError(errors.New("test error"))
//
//	// Using fixtures
//	session := fixtures.CreateTestSessionWithMessages("test-id", 5)
//	message := fixtures.CreateTestUserMessage("Hello")
//
//	// Using helpers
//	tempDir := helpers.CreateTempDir(t, "test")
//	helpers.AssertNoError(t, err)
//
// Test Organization Guidelines:
//
// - Use mocks for interface implementations in unit tests
// - Use fixtures for consistent test data across tests
// - Use helpers for common test operations (I/O, comparisons, context)
// - Keep test-specific utilities in the test file unless they're reusable
//
// Naming Conventions:
//
// - Mock types: Mock<Interface> (e.g., MockProvider, MockBackend)
// - Fixture functions: Create<Type> (e.g., CreateTestSession, CreateTestMessage)
// - Helper functions: <Action><Object> (e.g., CompareMessages, AssertNoError)
package testutil
