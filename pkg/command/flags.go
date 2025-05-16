// ABOUTME: Flags wrapper providing type-safe access to command flags
// ABOUTME: Implements convenience methods for getting typed values from flag map

package command

import (
	"strconv"
	"time"
)

// Flags provides type-safe access to command flags
type Flags struct {
	values map[string]interface{}
}

// NewFlags creates a new Flags wrapper
func NewFlags(values map[string]interface{}) *Flags {
	if values == nil {
		values = make(map[string]interface{})
	}
	return &Flags{values: values}
}

// GetString returns a string flag value
func (f *Flags) GetString(key string) string {
	if val, ok := f.values[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// GetInt returns an integer flag value
func (f *Flags) GetInt(key string) int {
	if val, ok := f.values[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case string:
			if intVal, err := strconv.Atoi(v); err == nil {
				return intVal
			}
		}
	}
	return 0
}

// GetFloat returns a float flag value
func (f *Flags) GetFloat(key string) float64 {
	if val, ok := f.values[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case string:
			if floatVal, err := strconv.ParseFloat(v, 64); err == nil {
				return floatVal
			}
		}
	}
	return 0
}

// GetBool returns a boolean flag value
func (f *Flags) GetBool(key string) bool {
	if val, ok := f.values[key]; ok {
		switch v := val.(type) {
		case bool:
			return v
		case string:
			if boolVal, err := strconv.ParseBool(v); err == nil {
				return boolVal
			}
		}
	}
	return false
}

// GetDuration returns a duration flag value
func (f *Flags) GetDuration(key string) time.Duration {
	if val, ok := f.values[key]; ok {
		switch v := val.(type) {
		case time.Duration:
			return v
		case string:
			if duration, err := time.ParseDuration(v); err == nil {
				return duration
			}
		}
	}
	return 0
}

// GetStringSlice returns a string slice flag value
func (f *Flags) GetStringSlice(key string) []string {
	if val, ok := f.values[key]; ok {
		switch v := val.(type) {
		case []string:
			return v
		case string:
			return []string{v}
		}
	}
	return nil
}

// Get returns the raw value for a key
func (f *Flags) Get(key string) interface{} {
	return f.values[key]
}

// Has checks if a flag exists
func (f *Flags) Has(key string) bool {
	_, ok := f.values[key]
	return ok
}

// All returns all flag values
func (f *Flags) All() map[string]interface{} {
	return f.values
}

// Set sets a flag value
func (f *Flags) Set(key string, value interface{}) {
	f.values[key] = value
}
