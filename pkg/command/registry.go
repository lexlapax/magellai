// ABOUTME: Central command registry for managing all commands
// ABOUTME: Provides registration, discovery, and lookup for commands across interfaces

package command

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	
	"github.com/lexlapax/magellai/internal/logging"
)

// Registry manages all registered commands
type Registry struct {
	mu       sync.RWMutex
	commands map[string]Interface
	aliases  map[string]string // alias -> primary name mapping
}

// GlobalRegistry is the default command registry
var GlobalRegistry = NewRegistry()

// NewRegistry creates a new command registry
func NewRegistry() *Registry {
	logging.LogDebug("Creating new command registry")
	return &Registry{
		commands: make(map[string]Interface),
		aliases:  make(map[string]string),
	}
}

// Register adds a command to the registry
func (r *Registry) Register(cmd Interface) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	meta := cmd.Metadata()
	if meta.Name == "" {
		logging.LogError(nil, "Cannot register command with empty name")
		return ErrInvalidCommand
	}

	logging.LogDebug("Registering command", "name", meta.Name, "category", meta.Category)

	// Validate the command
	if err := cmd.Validate(); err != nil {
		logging.LogError(err, "Command validation failed during registration", "name", meta.Name)
		return fmt.Errorf("command validation failed: %w", err)
	}

	// Check if command already exists
	if _, exists := r.commands[meta.Name]; exists {
		logging.LogError(nil, "Command already registered", "name", meta.Name)
		return ErrCommandAlreadyRegistered
	}

	// Register the primary command
	r.commands[meta.Name] = cmd
	logging.LogInfo("Command registered", "name", meta.Name, "category", meta.Category, "aliasCount", len(meta.Aliases))

	// Register aliases
	for _, alias := range meta.Aliases {
		if _, exists := r.aliases[alias]; exists {
			// Rollback registration if alias conflict
			delete(r.commands, meta.Name)
			logging.LogError(nil, "Alias conflict during registration", "alias", alias, "command", meta.Name)
			return fmt.Errorf("alias '%s' already registered", alias)
		}
		r.aliases[alias] = meta.Name
		logging.LogDebug("Alias registered", "alias", alias, "command", meta.Name)
	}

	return nil
}

// Get retrieves a command by name or alias
func (r *Registry) Get(name string) (Interface, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	logging.LogDebug("Looking up command", "name", name)

	// Check primary names first
	if cmd, exists := r.commands[name]; exists {
		logging.LogDebug("Command found by primary name", "name", name)
		return cmd, nil
	}

	// Check aliases
	if primaryName, exists := r.aliases[name]; exists {
		if cmd, exists := r.commands[primaryName]; exists {
			logging.LogDebug("Command found by alias", "alias", name, "primaryName", primaryName)
			return cmd, nil
		}
	}

	logging.LogError(nil, "Command not found", "name", name)
	return nil, ErrCommandNotFound
}

// List returns all commands filtered by category
func (r *Registry) List(category Category) []Interface {
	r.mu.RLock()
	defer r.mu.RUnlock()

	logging.LogDebug("Listing commands", "category", category)

	var commands []Interface
	for _, cmd := range r.commands {
		meta := cmd.Metadata()
		if category == CategoryShared || meta.Category == category || meta.Category == CategoryShared {
			commands = append(commands, cmd)
		}
	}

	logging.LogDebug("Commands found", "count", len(commands), "category", category)

	// Sort by name for consistent output
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Metadata().Name < commands[j].Metadata().Name
	})

	return commands
}

// Names returns all command names (primary and aliases)
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.commands)+len(r.aliases))

	// Add primary names
	for name := range r.commands {
		names = append(names, name)
	}

	// Add aliases
	for alias := range r.aliases {
		names = append(names, alias)
	}

	sort.Strings(names)
	return names
}

// Search finds commands by partial name match
func (r *Registry) Search(query string) []Interface {
	r.mu.RLock()
	defer r.mu.RUnlock()

	logging.LogDebug("Searching for commands", "query", query)

	query = strings.ToLower(query)
	var matches []Interface
	seen := make(map[string]bool)

	// Search in primary names
	for name, cmd := range r.commands {
		if strings.Contains(strings.ToLower(name), query) {
			matches = append(matches, cmd)
			seen[name] = true
		}
	}

	// Search in aliases
	for alias, primaryName := range r.aliases {
		if strings.Contains(strings.ToLower(alias), query) && !seen[primaryName] {
			if cmd, exists := r.commands[primaryName]; exists {
				matches = append(matches, cmd)
				seen[primaryName] = true
			}
		}
	}

	logging.LogDebug("Search completed", "query", query, "matchCount", len(matches))

	// Sort by name for consistent output
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Metadata().Name < matches[j].Metadata().Name
	})

	return matches
}

// Clear removes all commands from the registry
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	logging.LogDebug("Clearing command registry", "commandCount", len(r.commands), "aliasCount", len(r.aliases))

	r.commands = make(map[string]Interface)
	r.aliases = make(map[string]string)

	logging.LogInfo("Command registry cleared")
}

// GetExecutor returns a command executor for this registry
func (r *Registry) GetExecutor() *CommandExecutor {
	return NewExecutor(r)
}

// MustRegister registers a command and panics on error
func (r *Registry) MustRegister(cmd Interface) {
	if err := r.Register(cmd); err != nil {
		panic(fmt.Sprintf("failed to register command: %v", err))
	}
}

// Helper functions for global registry

// Register adds a command to the global registry
func Register(cmd Interface) error {
	return GlobalRegistry.Register(cmd)
}

// MustRegister registers a command to the global registry and panics on error
func MustRegister(cmd Interface) {
	GlobalRegistry.MustRegister(cmd)
}

// Get retrieves a command from the global registry
func Get(name string) (Interface, error) {
	return GlobalRegistry.Get(name)
}

// List returns commands from the global registry
func List(category Category) []Interface {
	return GlobalRegistry.List(category)
}

// Names returns all command names from the global registry
func Names() []string {
	return GlobalRegistry.Names()
}

// Search finds commands in the global registry
func Search(query string) []Interface {
	return GlobalRegistry.Search(query)
}
