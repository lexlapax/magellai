// ABOUTME: Command discovery and automatic registration mechanisms
// ABOUTME: Provides ways to discover and register commands from packages

package command

import (
	"fmt"
	"reflect"
	"strings"
)

// Discoverer defines the interface for command discovery
type Discoverer interface {
	// Discover finds and returns commands
	Discover() ([]Interface, error)
}

// PackageDiscoverer discovers commands from a package
type PackageDiscoverer struct {
	packagePath string
	prefix      string
}

// NewPackageDiscoverer creates a new package discoverer
func NewPackageDiscoverer(packagePath, prefix string) *PackageDiscoverer {
	return &PackageDiscoverer{
		packagePath: packagePath,
		prefix:      prefix,
	}
}

// Discover finds commands in the package
func (d *PackageDiscoverer) Discover() ([]Interface, error) {
	// This is a placeholder - in a real implementation, this would use
	// reflection or code generation to find commands in a package
	// For now, return empty slice
	return []Interface{}, nil
}

// RegisterFunction is a function that registers commands
type RegisterFunction func(*Registry) error

// AutoRegister automatically registers commands using register functions
func AutoRegister(registry *Registry, registerFuncs ...RegisterFunction) error {
	for _, fn := range registerFuncs {
		if err := fn(registry); err != nil {
			return fmt.Errorf("auto-registration failed: %w", err)
		}
	}
	return nil
}

// DiscoverAndRegister discovers and registers commands
func DiscoverAndRegister(registry *Registry, discoverers ...Discoverer) error {
	for _, discoverer := range discoverers {
		commands, err := discoverer.Discover()
		if err != nil {
			return fmt.Errorf("discovery failed: %w", err)
		}

		for _, cmd := range commands {
			if err := registry.Register(cmd); err != nil {
				return fmt.Errorf("registration failed for command '%s': %w",
					cmd.Metadata().Name, err)
			}
		}
	}
	return nil
}

// BuilderDiscoverer discovers commands from a builder pattern
type BuilderDiscoverer struct {
	builders []CommandBuilder
}

// CommandBuilder is a function that builds a command
type CommandBuilder func() (Interface, error)

// NewBuilderDiscoverer creates a new builder discoverer
func NewBuilderDiscoverer(builders ...CommandBuilder) *BuilderDiscoverer {
	return &BuilderDiscoverer{
		builders: builders,
	}
}

// Discover builds and returns commands
func (d *BuilderDiscoverer) Discover() ([]Interface, error) {
	var commands []Interface

	for _, builder := range d.builders {
		cmd, err := builder()
		if err != nil {
			return nil, fmt.Errorf("builder failed: %w", err)
		}
		commands = append(commands, cmd)
	}

	return commands, nil
}

// ReflectionDiscoverer discovers commands using reflection
type ReflectionDiscoverer struct {
	target interface{}
	prefix string
}

// NewReflectionDiscoverer creates a new reflection-based discoverer
func NewReflectionDiscoverer(target interface{}, prefix string) *ReflectionDiscoverer {
	return &ReflectionDiscoverer{
		target: target,
		prefix: prefix,
	}
}

// Discover finds commands using reflection
func (d *ReflectionDiscoverer) Discover() ([]Interface, error) {
	var commands []Interface

	targetType := reflect.TypeOf(d.target)
	targetValue := reflect.ValueOf(d.target)

	// Look for methods that match command pattern
	for i := 0; i < targetType.NumMethod(); i++ {
		method := targetType.Method(i)

		// Check if method name starts with prefix
		if !strings.HasPrefix(method.Name, d.prefix) {
			continue
		}

		// Check if method returns (Interface, error)
		if method.Type.NumOut() != 2 {
			continue
		}

		if !method.Type.Out(0).Implements(reflect.TypeOf((*Interface)(nil)).Elem()) {
			continue
		}

		if method.Type.Out(1) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}

		// Call the method to get the command
		results := method.Func.Call([]reflect.Value{targetValue})

		if err, ok := results[1].Interface().(error); ok && err != nil {
			return nil, fmt.Errorf("method %s failed: %w", method.Name, err)
		}

		if cmd, ok := results[0].Interface().(Interface); ok {
			commands = append(commands, cmd)
		}
	}

	return commands, nil
}

// RegistryInitializer provides a way to initialize registries
type RegistryInitializer struct {
	registry *Registry
}

// NewRegistryInitializer creates a new registry initializer
func NewRegistryInitializer(registry *Registry) *RegistryInitializer {
	return &RegistryInitializer{
		registry: registry,
	}
}

// Initialize sets up the registry with default commands
func (i *RegistryInitializer) Initialize() error {
	// This would be called by packages to register their commands
	// For example:
	// - Core commands package would call this to register help, version, etc.
	// - Plugin system would call this to register plugin commands
	// - Feature packages would register their specific commands

	return nil
}

// MustDiscoverAndRegister discovers and registers commands, panicking on error
func MustDiscoverAndRegister(registry *Registry, discoverers ...Discoverer) {
	if err := DiscoverAndRegister(registry, discoverers...); err != nil {
		panic(err)
	}
}
