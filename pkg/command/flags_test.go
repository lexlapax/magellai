// ABOUTME: Unit tests for the Flags wrapper
// ABOUTME: Tests type-safe access methods and conversion behavior

package command

import (
	"testing"
	"time"
)

func TestNewFlags(t *testing.T) {
	t.Run("Nil map creates empty flags", func(t *testing.T) {
		flags := NewFlags(nil)
		if flags == nil {
			t.Fatal("expected non-nil flags")
		}
		if len(flags.values) != 0 {
			t.Error("expected empty values map")
		}
	})

	t.Run("With values", func(t *testing.T) {
		values := map[string]interface{}{
			"test": "value",
		}
		flags := NewFlags(values)
		if flags.GetString("test") != "value" {
			t.Error("expected value to be preserved")
		}
	})
}

func TestFlagsGetString(t *testing.T) {
	flags := NewFlags(map[string]interface{}{
		"string":  "hello",
		"number":  42,
		"boolean": true,
	})

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{"existing string", "string", "hello"},
		{"non-existing", "missing", ""},
		{"wrong type", "number", ""},
		{"boolean", "boolean", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := flags.GetString(tt.key); got != tt.expected {
				t.Errorf("GetString(%q) = %v, want %v", tt.key, got, tt.expected)
			}
		})
	}
}

func TestFlagsGetInt(t *testing.T) {
	flags := NewFlags(map[string]interface{}{
		"int":      42,
		"string":   "123",
		"invalid":  "abc",
		"float":    3.14,
		"negative": "-10",
	})

	tests := []struct {
		name     string
		key      string
		expected int
	}{
		{"existing int", "int", 42},
		{"string number", "string", 123},
		{"negative string", "negative", -10},
		{"invalid string", "invalid", 0},
		{"float", "float", 0},
		{"non-existing", "missing", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := flags.GetInt(tt.key); got != tt.expected {
				t.Errorf("GetInt(%q) = %v, want %v", tt.key, got, tt.expected)
			}
		})
	}
}

func TestFlagsGetFloat(t *testing.T) {
	flags := NewFlags(map[string]interface{}{
		"float64":    3.14,
		"float32":    float32(2.5),
		"string":     "1.23",
		"invalid":    "abc",
		"int":        42,
		"negative":   "-3.14",
		"scientific": "1.23e-4",
	})

	tests := []struct {
		name     string
		key      string
		expected float64
	}{
		{"existing float64", "float64", 3.14},
		{"existing float32", "float32", 2.5},
		{"string float", "string", 1.23},
		{"negative string", "negative", -3.14},
		{"scientific notation", "scientific", 1.23e-4},
		{"invalid string", "invalid", 0},
		{"non-existing", "missing", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := flags.GetFloat(tt.key); got != tt.expected {
				t.Errorf("GetFloat(%q) = %v, want %v", tt.key, got, tt.expected)
			}
		})
	}
}

func TestFlagsGetBool(t *testing.T) {
	flags := NewFlags(map[string]interface{}{
		"true_bool":    true,
		"false_bool":   false,
		"true_string":  "true",
		"false_string": "false",
		"1_string":     "1",
		"0_string":     "0",
		"invalid":      "maybe",
		"number":       42,
	})

	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{"existing true bool", "true_bool", true},
		{"existing false bool", "false_bool", false},
		{"string true", "true_string", true},
		{"string false", "false_string", false},
		{"string 1", "1_string", true},
		{"string 0", "0_string", false},
		{"invalid string", "invalid", false},
		{"wrong type", "number", false},
		{"non-existing", "missing", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := flags.GetBool(tt.key); got != tt.expected {
				t.Errorf("GetBool(%q) = %v, want %v", tt.key, got, tt.expected)
			}
		})
	}
}

func TestFlagsGetDuration(t *testing.T) {
	flags := NewFlags(map[string]interface{}{
		"duration":       5 * time.Second,
		"string":         "10s",
		"complex_string": "1h30m",
		"invalid":        "abc",
		"number":         42,
	})

	tests := []struct {
		name     string
		key      string
		expected time.Duration
	}{
		{"existing duration", "duration", 5 * time.Second},
		{"string duration", "string", 10 * time.Second},
		{"complex string", "complex_string", 90 * time.Minute},
		{"invalid string", "invalid", 0},
		{"wrong type", "number", 0},
		{"non-existing", "missing", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := flags.GetDuration(tt.key); got != tt.expected {
				t.Errorf("GetDuration(%q) = %v, want %v", tt.key, got, tt.expected)
			}
		})
	}
}

func TestFlagsGetStringSlice(t *testing.T) {
	flags := NewFlags(map[string]interface{}{
		"slice":  []string{"a", "b", "c"},
		"single": "one",
		"number": 42,
		"empty":  []string{},
	})

	t.Run("existing slice", func(t *testing.T) {
		slice := flags.GetStringSlice("slice")
		if len(slice) != 3 || slice[0] != "a" || slice[1] != "b" || slice[2] != "c" {
			t.Errorf("GetStringSlice(\"slice\") = %v, want [a b c]", slice)
		}
	})

	t.Run("single string", func(t *testing.T) {
		slice := flags.GetStringSlice("single")
		if len(slice) != 1 || slice[0] != "one" {
			t.Errorf("GetStringSlice(\"single\") = %v, want [one]", slice)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		slice := flags.GetStringSlice("empty")
		if len(slice) != 0 {
			t.Errorf("GetStringSlice(\"empty\") = %v, want []", slice)
		}
	})

	t.Run("wrong type", func(t *testing.T) {
		slice := flags.GetStringSlice("number")
		if slice != nil {
			t.Errorf("GetStringSlice(\"number\") = %v, want nil", slice)
		}
	})

	t.Run("non-existing", func(t *testing.T) {
		slice := flags.GetStringSlice("missing")
		if slice != nil {
			t.Errorf("GetStringSlice(\"missing\") = %v, want nil", slice)
		}
	})
}

func TestFlagsGet(t *testing.T) {
	values := map[string]interface{}{
		"string": "hello",
		"number": 42,
		"bool":   true,
		"slice":  []string{"a", "b"},
	}
	flags := NewFlags(values)

	for key, expected := range values {
		t.Run(key, func(t *testing.T) {
			got := flags.Get(key)
			// Special handling for slices
			if key == "slice" {
				gotSlice, ok1 := got.([]string)
				expectedSlice, ok2 := expected.([]string)
				if !ok1 || !ok2 {
					t.Errorf("Get(%q) type mismatch", key)
					return
				}
				if len(gotSlice) != len(expectedSlice) {
					t.Errorf("Get(%q) = %v, want %v", key, got, expected)
					return
				}
				for i := range gotSlice {
					if gotSlice[i] != expectedSlice[i] {
						t.Errorf("Get(%q) = %v, want %v", key, got, expected)
						return
					}
				}
			} else if got != expected {
				t.Errorf("Get(%q) = %v, want %v", key, got, expected)
			}
		})
	}

	t.Run("non-existing", func(t *testing.T) {
		if got := flags.Get("missing"); got != nil {
			t.Errorf("Get(\"missing\") = %v, want nil", got)
		}
	})
}

func TestFlagsHas(t *testing.T) {
	flags := NewFlags(map[string]interface{}{
		"exists": "value",
		"zero":   0,
		"nil":    nil,
	})

	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{"existing", "exists", true},
		{"zero value", "zero", true},
		{"nil value", "nil", true},
		{"non-existing", "missing", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := flags.Has(tt.key); got != tt.expected {
				t.Errorf("Has(%q) = %v, want %v", tt.key, got, tt.expected)
			}
		})
	}
}

func TestFlagsAll(t *testing.T) {
	values := map[string]interface{}{
		"string": "hello",
		"number": 42,
		"bool":   true,
	}
	flags := NewFlags(values)

	all := flags.All()
	if len(all) != len(values) {
		t.Errorf("All() returned %d items, want %d", len(all), len(values))
	}

	for key, expected := range values {
		if got, ok := all[key]; !ok || got != expected {
			t.Errorf("All()[%q] = %v, want %v", key, got, expected)
		}
	}
}
