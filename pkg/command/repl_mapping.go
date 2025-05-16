// ABOUTME: REPL-specific command mappings and conversions
// ABOUTME: Maps CLI flags to REPL commands and handles conversions

package command

import (
	"fmt"
	"strings"
)

// REPLMapping defines how CLI flags map to REPL commands
type REPLMapping struct {
	// Flag is the CLI flag name
	Flag string

	// REPLCommand is the corresponding REPL command
	REPLCommand string

	// ValueTransform is an optional function to transform the flag value
	ValueTransform func(interface{}) (string, error)
}

// REPLMapper manages flag-to-command mappings for REPL
type REPLMapper struct {
	mappings map[string]REPLMapping
}

// NewREPLMapper creates a new REPL mapper
func NewREPLMapper() *REPLMapper {
	return &REPLMapper{
		mappings: make(map[string]REPLMapping),
	}
}

// AddMapping adds a flag-to-command mapping
func (m *REPLMapper) AddMapping(mapping REPLMapping) {
	m.mappings[mapping.Flag] = mapping
}

// ConvertFlagToCommand converts a CLI flag to a REPL command
func (m *REPLMapper) ConvertFlagToCommand(flag string, value interface{}) (string, error) {
	mapping, exists := m.mappings[flag]
	if !exists {
		return "", fmt.Errorf("no REPL mapping for flag '%s'", flag)
	}

	// Transform the value if needed
	var strValue string
	if mapping.ValueTransform != nil {
		transformed, err := mapping.ValueTransform(value)
		if err != nil {
			return "", fmt.Errorf("failed to transform value for flag '%s': %w", flag, err)
		}
		strValue = transformed
	} else {
		strValue = fmt.Sprintf("%v", value)
	}

	// Handle boolean flags specially
	if boolVal, ok := value.(bool); ok {
		// If there's a transform, use it
		if mapping.ValueTransform != nil {
			transformed, err := mapping.ValueTransform(value)
			if err != nil {
				return "", fmt.Errorf("failed to transform value for flag '%s': %w", flag, err)
			}
			return fmt.Sprintf("%s %s", mapping.REPLCommand, transformed), nil
		}

		// Default boolean handling
		if boolVal {
			return mapping.REPLCommand, nil
		}
		// For false boolean flags, we might want to return a different command
		// or no command at all
		return "", nil
	}

	// Build the REPL command
	return fmt.Sprintf("%s %s", mapping.REPLCommand, strValue), nil
}

// GetMapping returns the mapping for a given flag
func (m *REPLMapper) GetMapping(flag string) (REPLMapping, bool) {
	mapping, exists := m.mappings[flag]
	return mapping, exists
}

// DefaultREPLMappings returns the default flag-to-command mappings
func DefaultREPLMappings() *REPLMapper {
	mapper := NewREPLMapper()

	// Add default mappings
	mapper.AddMapping(REPLMapping{
		Flag:        "stream",
		REPLCommand: ":stream",
		ValueTransform: func(v interface{}) (string, error) {
			if boolVal, ok := v.(bool); ok {
				if boolVal {
					return "on", nil
				}
				return "off", nil
			}
			return "", fmt.Errorf("expected boolean value")
		},
	})

	mapper.AddMapping(REPLMapping{
		Flag:        "model",
		REPLCommand: ":model",
	})

	mapper.AddMapping(REPLMapping{
		Flag:        "temperature",
		REPLCommand: ":temperature",
	})

	mapper.AddMapping(REPLMapping{
		Flag:        "max-tokens",
		REPLCommand: ":max_tokens",
	})

	mapper.AddMapping(REPLMapping{
		Flag:        "profile",
		REPLCommand: ":profile",
	})

	mapper.AddMapping(REPLMapping{
		Flag:        "output",
		REPLCommand: ":output",
	})

	mapper.AddMapping(REPLMapping{
		Flag:        "verbosity",
		REPLCommand: ":verbosity",
	})

	mapper.AddMapping(REPLMapping{
		Flag:        "attach",
		REPLCommand: ":attach",
	})

	return mapper
}

// ParseREPLCommand parses a REPL command into command name and arguments
func ParseREPLCommand(input string) (string, []string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", nil, fmt.Errorf("empty command")
	}

	// Check for command prefix
	if !strings.HasPrefix(input, "/") && !strings.HasPrefix(input, ":") {
		return "", nil, fmt.Errorf("not a REPL command")
	}

	// Split into parts
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return "", nil, fmt.Errorf("invalid command format")
	}

	cmd := parts[0]
	args := parts[1:]

	return cmd, args, nil
}

// IsREPLCommand checks if input is a REPL command
func IsREPLCommand(input string) bool {
	input = strings.TrimSpace(input)
	return strings.HasPrefix(input, "/") || strings.HasPrefix(input, ":")
}
