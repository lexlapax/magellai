// ABOUTME: Tests for shared context functionality
// ABOUTME: Validates thread-safety and correct behavior of shared state storage

package command

import (
	"sync"
	"testing"
)

func TestSharedContext(t *testing.T) {
	t.Run("Basic Operations", func(t *testing.T) {
		sc := NewSharedContext()

		// Test Set and Get
		sc.Set("key1", "value1")
		val, exists := sc.Get("key1")
		if !exists {
			t.Error("Expected value to exist")
		}
		if val != "value1" {
			t.Errorf("Expected 'value1', got %v", val)
		}

		// Test missing key
		_, exists = sc.Get("nonexistent")
		if exists {
			t.Error("Expected key to not exist")
		}

		// Test Delete
		sc.Delete("key1")
		_, exists = sc.Get("key1")
		if exists {
			t.Error("Expected key to be deleted")
		}
	})

	t.Run("Type-specific Getters", func(t *testing.T) {
		sc := NewSharedContext()

		// String
		sc.Set("string", "test")
		str, ok := sc.GetString("string")
		if !ok || str != "test" {
			t.Errorf("GetString failed: got %v, %v", str, ok)
		}

		// Int
		sc.Set("int", 42)
		num, ok := sc.GetInt("int")
		if !ok || num != 42 {
			t.Errorf("GetInt failed: got %v, %v", num, ok)
		}

		// Float64
		sc.Set("float", 3.14)
		f, ok := sc.GetFloat64("float")
		if !ok || f != 3.14 {
			t.Errorf("GetFloat64 failed: got %v, %v", f, ok)
		}

		// Bool
		sc.Set("bool", true)
		b, ok := sc.GetBool("bool")
		if !ok || !b {
			t.Errorf("GetBool failed: got %v, %v", b, ok)
		}

		// Wrong type
		sc.Set("wrongtype", "string")
		_, ok = sc.GetInt("wrongtype")
		if ok {
			t.Error("Expected type conversion to fail")
		}
	})

	t.Run("Clone", func(t *testing.T) {
		sc := NewSharedContext()
		sc.Set("key1", "value1")
		sc.Set("key2", 42)

		clone := sc.Clone()
		if len(clone) != 2 {
			t.Errorf("Expected 2 items in clone, got %d", len(clone))
		}
		if clone["key1"] != "value1" {
			t.Errorf("Expected clone to contain key1")
		}

		// Modify clone should not affect original
		clone["key1"] = "modified"
		val, _ := sc.Get("key1")
		if val != "value1" {
			t.Error("Original was modified")
		}
	})

	t.Run("Clear", func(t *testing.T) {
		sc := NewSharedContext()
		sc.Set("key1", "value1")
		sc.Set("key2", "value2")

		sc.Clear()

		clone := sc.Clone()
		if len(clone) != 0 {
			t.Errorf("Expected empty context after clear, got %d items", len(clone))
		}
	})

	t.Run("Thread Safety", func(t *testing.T) {
		sc := NewSharedContext()
		var wg sync.WaitGroup

		// Multiple writers
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				key := string(rune('a' + i))
				sc.Set(key, i)
			}(i)
		}

		// Multiple readers
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				key := string(rune('a' + i))
				sc.Get(key)
			}(i)
		}

		wg.Wait()

		// Verify all writes succeeded
		for i := 0; i < 10; i++ {
			key := string(rune('a' + i))
			val, exists := sc.Get(key)
			if !exists {
				t.Errorf("Expected key %s to exist", key)
			}
			if val != i {
				t.Errorf("Expected value %d, got %v", i, val)
			}
		}
	})
}

func TestSharedContextConstants(t *testing.T) {
	// Just verify constants are defined
	constants := []string{
		SharedContextModel,
		SharedContextProvider,
		SharedContextTemperature,
		SharedContextMaxTokens,
		SharedContextStream,
		SharedContextSessionID,
		SharedContextSessionName,
		SharedContextMultiline,
		SharedContextPrompt,
		SharedContextVerbosity,
		SharedContextOutput,
		SharedContextPendingAttachments,
	}

	for _, c := range constants {
		if c == "" {
			t.Error("Found empty constant")
		}
	}
}
