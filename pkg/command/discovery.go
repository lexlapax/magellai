// ABOUTME: Command discovery and automatic registration mechanisms
// ABOUTME: Provides ways to discover and register commands from packages

package command

import (
	"fmt"
	"reflect"
	"strings"
)

// discoverer defines the interface for command discovery
type discoverer interface {
	// discover finds and returns commands
	discover() ([]Interface, error)
}

// packageDiscoverer discovers commands from a package
type packageDiscoverer struct {
	packagePath string
	prefix      string
}

// newPackageDiscoverer creates a new package discoverer
func newPackageDiscoverer(packagePath, prefix string) *packageDiscoverer {
	return &packageDiscoverer{
		packagePath: packagePath,
		prefix:      prefix,
	}
}

// discover finds commands in the package
func (d *packageDiscoverer) discover() ([]Interface, error) {
	// This is a placeholder - in a real implementation, this would use
	// reflection or code generation to find commands in a package
	// For now, return empty slice
	return []Interface{}, nil
}

// registerFunction is a function that registers commands
type registerFunction func(*Registry) error

// autoRegister automatically registers commands using register functions
func autoRegister(registry *Registry, registerFuncs ...registerFunction) error {
	for _, fn := range registerFuncs {
		if err := fn(registry); err != nil {
			return fmt.Errorf("auto-registration failed: %w", err)
		}
	}
	return nil
}

// discoverAndRegister discovers and registers commands
func discoverAndRegister(registry *Registry, discoverers ...discoverer) error {
	for _, discoverer := range discoverers {
		commands, err := discoverer.discover()
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

// builderDiscoverer discovers commands from a builder pattern
type builderDiscoverer struct {
	builders []commandBuilder
}

// commandBuilder is a function that builds a command
type commandBuilder func() (Interface, error)

// newBuilderDiscoverer creates a new builder discoverer
func newBuilderDiscoverer(builders ...commandBuilder) *builderDiscoverer {
	return &builderDiscoverer{
		builders: builders,
	}
}

// discover builds and returns commands
func (d *builderDiscoverer) discover() ([]Interface, error) {
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

// reflectionDiscoverer discovers commands using reflection
type reflectionDiscoverer struct {
	target interface{}
	prefix string
}

// newReflectionDiscoverer creates a new reflection-based discoverer
func newReflectionDiscoverer(target interface{}, prefix string) *reflectionDiscoverer {
	return &reflectionDiscoverer{
		target: target,
		prefix: prefix,
	}
}

// discover finds commands using reflection
func (d *reflectionDiscoverer) discover() ([]Interface, error) {
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

// registryInitializer provides a way to initialize registries
type registryInitializer struct {
	registry *Registry
}

// newRegistryInitializer creates a new registry initializer
func newRegistryInitializer(registry *Registry) *registryInitializer {
	return &registryInitializer{
		registry: registry,
	}
}

// initialize sets up the registry with default commands
func (i *registryInitializer) initialize() error {
	// This would be called by packages to register their commands
	// For example:
	// - Core commands package would call this to register help, version, etc.
	// - Plugin system would call this to register plugin commands
	// - Feature packages would register their specific commands

	return nil
}

// mustDiscoverAndRegister discovers and registers commands, panicking on error
func mustDiscoverAndRegister(registry *Registry, discoverers ...discoverer) {
	if err := discoverAndRegister(registry, discoverers...); err != nil {
		panic(err)
	}
}
