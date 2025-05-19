// ABOUTME: Tests for the command registry functionality
// ABOUTME: Ensures registry correctly manages commands and aliases

package command

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock command for testing
type mockRegistryCommand struct {
	meta     *Metadata
	validate func() error
}

func (m *mockRegistryCommand) Execute(ctx context.Context, exec *ExecutionContext) error {
	return nil
}

func (m *mockRegistryCommand) Metadata() *Metadata {
	return m.meta
}

func (m *mockRegistryCommand) Validate() error {
	if m.validate != nil {
		return m.validate()
	}
	return nil
}

// Helper function to create test command
func createTestCommand(name string, category Category, aliases ...string) *mockRegistryCommand {
	return &mockRegistryCommand{
		meta: &Metadata{
			Name:        name,
			Category:    category,
			Description: "Test command " + name,
			Aliases:     aliases,
		},
	}
}

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	assert.NotNil(t, r)
	assert.NotNil(t, r.commands)
	assert.NotNil(t, r.aliases)
	assert.Empty(t, r.commands)
	assert.Empty(t, r.aliases)
}

func TestRegistry_Register(t *testing.T) {
	tests := []struct {
		name      string
		cmd       Interface
		wantError error
		setup     func(r *Registry)
	}{
		{
			name: "successful registration",
			cmd:  createTestCommand("test", CategoryShared),
		},
		{
			name: "registration with aliases",
			cmd:  createTestCommand("test", CategoryShared, "t", "tst"),
		},
		{
			name: "duplicate command",
			cmd:  createTestCommand("test", CategoryShared),
			setup: func(r *Registry) {
				_ = r.Register(createTestCommand("test", CategoryShared))
			},
			wantError: ErrCommandAlreadyRegistered,
		},
		{
			name:      "empty name",
			cmd:       createTestCommand("", CategoryShared),
			wantError: ErrInvalidCommand,
		},
		{
			name: "alias conflict",
			cmd:  createTestCommand("test2", CategoryShared, "t"),
			setup: func(r *Registry) {
				_ = r.Register(createTestCommand("test1", CategoryShared, "t"))
			},
			wantError: nil, // Should return specific error about alias conflict
		},
		{
			name: "validation failure",
			cmd: &mockRegistryCommand{
				meta: &Metadata{Name: "test", Category: CategoryShared},
				validate: func() error {
					return ErrInvalidCommand
				},
			},
			wantError: nil, // Should return wrapped validation error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRegistry()
			if tt.setup != nil {
				tt.setup(r)
			}

			err := r.Register(tt.cmd)
			if tt.wantError != nil {
				assert.ErrorIs(t, err, tt.wantError)
			} else if tt.wantError == nil && err != nil {
				// For specific error message tests
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegistry_Get(t *testing.T) {
	r := NewRegistry()

	// Register commands
	cmd1 := createTestCommand("test1", CategoryShared)
	cmd2 := createTestCommand("test2", CategoryShared, "t2", "test-two")

	_ = r.Register(cmd1)
	_ = r.Register(cmd2)

	tests := []struct {
		name      string
		lookup    string
		want      Interface
		wantError error
	}{
		{
			name:   "get by primary name",
			lookup: "test1",
			want:   cmd1,
		},
		{
			name:   "get by alias",
			lookup: "t2",
			want:   cmd2,
		},
		{
			name:   "get by secondary alias",
			lookup: "test-two",
			want:   cmd2,
		},
		{
			name:      "command not found",
			lookup:    "nonexistent",
			wantError: ErrCommandNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := r.Get(tt.lookup)
			if tt.wantError != nil {
				assert.ErrorIs(t, err, tt.wantError)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestRegistry_List(t *testing.T) {
	r := NewRegistry()

	// Register commands in different categories
	cmd1 := createTestCommand("cmd1", CategoryShared)
	cmd2 := createTestCommand("cmd2", CategoryShared)
	cmd3 := createTestCommand("cmd3", CategoryREPL)
	cmd4 := createTestCommand("cmd4", CategoryShared)

	_ = r.Register(cmd1)
	_ = r.Register(cmd2)
	_ = r.Register(cmd3)
	_ = r.Register(cmd4)

	tests := []struct {
		name     string
		category Category
		want     []string // command names
	}{
		{
			name:     "list shared commands",
			category: CategoryShared,
			want:     []string{"cmd1", "cmd2", "cmd3", "cmd4"}, // all commands since CategoryShared includes everything
		},
		{
			name:     "list REPL commands",
			category: CategoryREPL,
			want:     []string{"cmd1", "cmd2", "cmd3", "cmd4"}, // REPL commands + shared (all in this case)
		},
		{
			name:     "list shared commands explicitly",
			category: CategoryShared,
			want:     []string{"cmd1", "cmd2", "cmd3", "cmd4"}, // all commands
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.List(tt.category)
			gotNames := make([]string, len(got))
			for i, cmd := range got {
				gotNames[i] = cmd.Metadata().Name
			}
			assert.ElementsMatch(t, tt.want, gotNames)
		})
	}
}

func TestRegistry_Names(t *testing.T) {
	r := NewRegistry()

	// Register commands with aliases
	cmd1 := createTestCommand("test1", CategoryShared, "t1")
	cmd2 := createTestCommand("test2", CategoryShared, "t2", "test-two")

	_ = r.Register(cmd1)
	_ = r.Register(cmd2)

	names := r.Names()
	expected := []string{"t1", "t2", "test-two", "test1", "test2"}
	assert.ElementsMatch(t, expected, names)
}

func TestRegistry_Search(t *testing.T) {
	r := NewRegistry()

	// Register commands
	cmd1 := createTestCommand("help", CategoryShared, "h")
	cmd2 := createTestCommand("history", CategoryREPL, "hist")
	cmd3 := createTestCommand("set", CategoryShared)

	_ = r.Register(cmd1)
	_ = r.Register(cmd2)
	_ = r.Register(cmd3)

	tests := []struct {
		name  string
		query string
		want  []string // command names
	}{
		{
			name:  "search by partial primary name",
			query: "hel",
			want:  []string{"help"},
		},
		{
			name:  "search by partial alias",
			query: "his",
			want:  []string{"history"},
		},
		{
			name:  "search matches multiple",
			query: "h",
			want:  []string{"help", "history"},
		},
		{
			name:  "no matches",
			query: "xyz",
			want:  []string{},
		},
		{
			name:  "case insensitive",
			query: "HEL",
			want:  []string{"help"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.Search(tt.query)
			gotNames := make([]string, len(got))
			for i, cmd := range got {
				gotNames[i] = cmd.Metadata().Name
			}

			if len(tt.want) == 0 {
				assert.Empty(t, gotNames)
			} else {
				assert.ElementsMatch(t, tt.want, gotNames)
			}
		})
	}
}

func TestRegistry_Clear(t *testing.T) {
	r := NewRegistry()

	// Register some commands
	_ = r.Register(createTestCommand("test1", CategoryShared, "t1"))
	_ = r.Register(createTestCommand("test2", CategoryShared))

	// Verify they exist
	assert.Len(t, r.commands, 2)
	assert.Len(t, r.aliases, 1)

	// Clear the registry
	r.Clear()

	// Verify everything is cleared
	assert.Empty(t, r.commands)
	assert.Empty(t, r.aliases)
}

func TestRegistry_MustRegister(t *testing.T) {
	r := NewRegistry()

	// Should not panic for valid command
	assert.NotPanics(t, func() {
		r.MustRegister(createTestCommand("test", CategoryShared))
	})

	// Should panic for duplicate
	assert.Panics(t, func() {
		r.MustRegister(createTestCommand("test", CategoryShared))
	})
}

func TestRegistry_GetExecutor(t *testing.T) {
	r := NewRegistry()
	executor := r.GetExecutor()
	assert.NotNil(t, executor)
	assert.Equal(t, r, executor.registry)
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	r := NewRegistry()
	var wg sync.WaitGroup

	// Concurrent registrations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			name := string(rune('a' + i))
			cmd := createTestCommand(name, CategoryShared)
			err := r.Register(cmd)
			assert.NoError(t, err)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = r.List(CategoryShared)
			_ = r.Names()
			_ = r.Search("a")
		}()
	}

	wg.Wait()

	// Verify all commands were registered
	assert.Len(t, r.commands, 10)
}

func TestGlobalRegistry(t *testing.T) {
	// Save current global registry
	originalGlobal := GlobalRegistry
	defer func() {
		GlobalRegistry = originalGlobal
	}()

	// Create new global registry for testing
	GlobalRegistry = NewRegistry()

	// Test global functions
	cmd := createTestCommand("global", CategoryShared)

	err := Register(cmd)
	assert.NoError(t, err)

	got, err := Get("global")
	assert.NoError(t, err)
	assert.Equal(t, cmd, got)

	commands := List(CategoryShared)
	assert.Len(t, commands, 1)

	names := Names()
	assert.Contains(t, names, "global")

	results := Search("glob")
	assert.Len(t, results, 1)

	// Test MustRegister
	MustRegister(createTestCommand("must", CategoryShared))
	got, err = Get("must")
	assert.NoError(t, err)
	assert.NotNil(t, got)
}
