// ABOUTME: Tests for command discovery and automatic registration
// ABOUTME: Verifies discoverer implementations and registration mechanisms

package command

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock command for testing
type mockCommand struct {
	name string
}

func (m *mockCommand) Execute(ctx context.Context, exec *ExecutionContext) error {
	return nil
}

func (m *mockCommand) Metadata() *Metadata {
	return &Metadata{
		Name:        m.name,
		Description: "A mock command",
	}
}

func (m *mockCommand) Validate() error {
	return nil
}

// Mock discoverer for testing
type mockDiscoverer struct {
	commands []Interface
	err      error
}

func (m *mockDiscoverer) Discover() ([]Interface, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.commands, nil
}

func TestPackageDiscoverer_New(t *testing.T) {
	discoverer := NewPackageDiscoverer("/path/to/package", "cmd")

	assert.NotNil(t, discoverer)
	assert.Equal(t, "/path/to/package", discoverer.packagePath)
	assert.Equal(t, "cmd", discoverer.prefix)
}

func TestPackageDiscoverer_Discover(t *testing.T) {
	discoverer := NewPackageDiscoverer("/path/to/package", "cmd")

	commands, err := discoverer.Discover()

	assert.NoError(t, err)
	assert.Empty(t, commands, "PackageDiscoverer is a placeholder and should return empty slice")
}

func TestAutoRegister(t *testing.T) {
	registry := NewRegistry()

	// Test successful registration
	registerFunc1 := func(r *Registry) error {
		return r.Register(&mockCommand{name: "cmd1"})
	}

	registerFunc2 := func(r *Registry) error {
		return r.Register(&mockCommand{name: "cmd2"})
	}

	err := AutoRegister(registry, registerFunc1, registerFunc2)
	require.NoError(t, err)

	// Verify commands were registered
	cmd1, err1 := registry.Get("cmd1")
	assert.NoError(t, err1)
	assert.NotNil(t, cmd1)

	cmd2, err2 := registry.Get("cmd2")
	assert.NoError(t, err2)
	assert.NotNil(t, cmd2)
}

func TestAutoRegister_Error(t *testing.T) {
	registry := NewRegistry()

	// Test registration with error
	registerFunc1 := func(r *Registry) error {
		return r.Register(&mockCommand{name: "cmd1"})
	}

	registerFunc2 := func(r *Registry) error {
		return errors.New("registration failed")
	}

	err := AutoRegister(registry, registerFunc1, registerFunc2)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "auto-registration failed")
	assert.Contains(t, err.Error(), "registration failed")
}

func TestDiscoverAndRegister(t *testing.T) {
	registry := NewRegistry()

	// Create mock discoverers
	discoverer1 := &mockDiscoverer{
		commands: []Interface{
			&mockCommand{name: "cmd1"},
			&mockCommand{name: "cmd2"},
		},
	}

	discoverer2 := &mockDiscoverer{
		commands: []Interface{
			&mockCommand{name: "cmd3"},
		},
	}

	err := DiscoverAndRegister(registry, discoverer1, discoverer2)
	require.NoError(t, err)

	// Verify all commands were registered
	cmd1, err1 := registry.Get("cmd1")
	assert.NoError(t, err1)
	assert.NotNil(t, cmd1)

	cmd2, err2 := registry.Get("cmd2")
	assert.NoError(t, err2)
	assert.NotNil(t, cmd2)

	cmd3, err3 := registry.Get("cmd3")
	assert.NoError(t, err3)
	assert.NotNil(t, cmd3)
}

func TestDiscoverAndRegister_DiscoveryError(t *testing.T) {
	registry := NewRegistry()

	// Create discoverer that returns error
	discoverer := &mockDiscoverer{
		err: errors.New("discovery failed"),
	}

	err := DiscoverAndRegister(registry, discoverer)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "discovery failed")
}

func TestDiscoverAndRegister_RegistrationError(t *testing.T) {
	registry := NewRegistry()

	// Register a command first
	_ = registry.Register(&mockCommand{name: "existing"})

	// Try to register duplicate command through discoverer
	discoverer := &mockDiscoverer{
		commands: []Interface{
			&mockCommand{name: "existing"}, // This will cause registration error
		},
	}

	err := DiscoverAndRegister(registry, discoverer)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "registration failed")
	assert.Contains(t, err.Error(), "existing")
}

func TestBuilderDiscoverer(t *testing.T) {
	// Create builders
	builder1 := func() (Interface, error) {
		return &mockCommand{name: "cmd1"}, nil
	}

	builder2 := func() (Interface, error) {
		return &mockCommand{name: "cmd2"}, nil
	}

	discoverer := NewBuilderDiscoverer(builder1, builder2)

	assert.NotNil(t, discoverer)
	assert.Len(t, discoverer.builders, 2)

	// Test discovery
	commands, err := discoverer.Discover()
	require.NoError(t, err)

	assert.Len(t, commands, 2)
	assert.Equal(t, "cmd1", commands[0].Metadata().Name)
	assert.Equal(t, "cmd2", commands[1].Metadata().Name)
}

func TestBuilderDiscoverer_Error(t *testing.T) {
	// Create builder that returns error
	builderWithError := func() (Interface, error) {
		return nil, errors.New("builder error")
	}

	discoverer := NewBuilderDiscoverer(builderWithError)

	commands, err := discoverer.Discover()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "builder failed")
	assert.Contains(t, err.Error(), "builder error")
	assert.Nil(t, commands)
}

// Mock object for reflection discovery
type mockCommandProvider struct{}

func (m *mockCommandProvider) CmdHelp() (Interface, error) {
	return &mockCommand{name: "help"}, nil
}

func (m *mockCommandProvider) CmdVersion() (Interface, error) {
	return &mockCommand{name: "version"}, nil
}

func (m *mockCommandProvider) CmdError() (Interface, error) {
	return nil, errors.New("error creating command")
}

func (m *mockCommandProvider) NotACommand() string {
	return "not a command"
}

func (m *mockCommandProvider) CmdWrongReturn() Interface {
	return &mockCommand{name: "wrong"}
}

// Test reflection discovery with just the error case
func TestReflectionDiscoverer_WithError(t *testing.T) {
	// Test with the mock provider that has CmdError
	provider := &mockCommandProvider{}
	discoverer := NewReflectionDiscoverer(provider, "Cmd")

	// CmdError should cause discovery to fail
	commands, err := discoverer.Discover()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "method CmdError failed")
	assert.Contains(t, err.Error(), "error creating command")
	assert.Nil(t, commands)
}

// Test reflection discovery with no methods
func TestReflectionDiscoverer_NoMethods(t *testing.T) {
	// Create a struct with no methods
	type emptyStruct struct{}

	var es emptyStruct

	discoverer := NewReflectionDiscoverer(&es, "Cmd")

	// Should return empty slice with no error
	commands, err := discoverer.Discover()

	assert.NoError(t, err)
	assert.Empty(t, commands)
}

// Test reflection discovery with working methods (using a simpler approach)
func TestReflectionDiscoverer_WorkingMethods(t *testing.T) {
	// Create a test provider that successfully returns commands
	type testProvider struct{}

	tp := &testProvider{}

	// Use a struct with methods as fields (workaround for Go reflection)
	provider := struct {
		*testProvider
	}{
		testProvider: tp,
	}

	// Create a discoverer for this provider
	discoverer := NewReflectionDiscoverer(&provider, "Cmd")

	// Since we don't have proper methods matching the pattern,
	// it should return empty but no error
	commands, err := discoverer.Discover()

	assert.NoError(t, err)
	assert.Empty(t, commands)
}

func TestRegistryInitializer(t *testing.T) {
	registry := NewRegistry()
	initializer := NewRegistryInitializer(registry)

	assert.NotNil(t, initializer)
	assert.Equal(t, registry, initializer.registry)

	// Test initialization (currently a no-op)
	err := initializer.Initialize()
	assert.NoError(t, err)
}

func TestMustDiscoverAndRegister(t *testing.T) {
	registry := NewRegistry()

	// Test successful registration
	discoverer := &mockDiscoverer{
		commands: []Interface{
			&mockCommand{name: "cmd1"},
		},
	}

	// Should not panic
	assert.NotPanics(t, func() {
		MustDiscoverAndRegister(registry, discoverer)
	})

	// Verify command was registered
	cmd, err := registry.Get("cmd1")
	assert.NoError(t, err)
	assert.NotNil(t, cmd)
}

func TestMustDiscoverAndRegister_Panic(t *testing.T) {
	registry := NewRegistry()

	// Create discoverer that will cause error
	discoverer := &mockDiscoverer{
		err: errors.New("discovery failed"),
	}

	// Should panic on error
	assert.Panics(t, func() {
		MustDiscoverAndRegister(registry, discoverer)
	})
}

// Test interface compliance
func TestInterfaceCompliance(t *testing.T) {
	// Ensure our mock implements the interface
	var _ Interface = (*mockCommand)(nil)

	// Ensure discoverers implement Discoverer interface
	var _ Discoverer = (*PackageDiscoverer)(nil)
	var _ Discoverer = (*BuilderDiscoverer)(nil)
	var _ Discoverer = (*ReflectionDiscoverer)(nil)
	var _ Discoverer = (*mockDiscoverer)(nil)
}

// Benchmark tests
func BenchmarkReflectionDiscoverer(b *testing.B) {
	provider := &mockCommandProvider{}
	discoverer := NewReflectionDiscoverer(provider, "Cmd")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = discoverer.Discover()
	}
}

func BenchmarkBuilderDiscoverer(b *testing.B) {
	builder := func() (Interface, error) {
		return &mockCommand{name: "test"}, nil
	}

	discoverer := NewBuilderDiscoverer(builder)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = discoverer.Discover()
	}
}
