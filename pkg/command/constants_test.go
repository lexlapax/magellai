// ABOUTME: Tests for command package constants
// ABOUTME: Verifies constant values and type definitions

package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOutputFormatConstants(t *testing.T) {
	// Test that constants have expected values
	assert.Equal(t, OutputFormat("text"), OutputFormatText)
	assert.Equal(t, OutputFormat("json"), OutputFormatJSON)
	assert.Equal(t, OutputFormat("yaml"), OutputFormatYAML)
}

func TestOutputFormatType(t *testing.T) {
	// Test that the type works as expected
	var format OutputFormat
	
	// Test assignment
	format = OutputFormatText
	assert.Equal(t, "text", string(format))
	
	format = OutputFormatJSON
	assert.Equal(t, "json", string(format))
	
	format = OutputFormatYAML
	assert.Equal(t, "yaml", string(format))
}

func TestOutputFormatString(t *testing.T) {
	tests := []struct {
		format   OutputFormat
		expected string
	}{
		{OutputFormatText, "text"},
		{OutputFormatJSON, "json"},
		{OutputFormatYAML, "yaml"},
	}
	
	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.format))
		})
	}
}

func TestOutputFormatValues(t *testing.T) {
	// Test that we can create variables with these constants
	formats := []OutputFormat{
		OutputFormatText,
		OutputFormatJSON,
		OutputFormatYAML,
	}
	
	assert.Len(t, formats, 3)
	assert.Contains(t, formats, OutputFormatText)
	assert.Contains(t, formats, OutputFormatJSON)
	assert.Contains(t, formats, OutputFormatYAML)
}

func TestOutputFormatSwitch(t *testing.T) {
	// Test that we can use the constants in switch statements
	testFormat := func(format OutputFormat) string {
		switch format {
		case OutputFormatText:
			return "text output"
		case OutputFormatJSON:
			return "json output"
		case OutputFormatYAML:
			return "yaml output"
		default:
			return "unknown"
		}
	}
	
	assert.Equal(t, "text output", testFormat(OutputFormatText))
	assert.Equal(t, "json output", testFormat(OutputFormatJSON))
	assert.Equal(t, "yaml output", testFormat(OutputFormatYAML))
	assert.Equal(t, "unknown", testFormat(OutputFormat("xml")))
}

func TestOutputFormatComparison(t *testing.T) {
	// Test that constants can be compared
	assert.True(t, OutputFormatText == OutputFormatText)
	assert.False(t, OutputFormatText == OutputFormatJSON)
	assert.False(t, OutputFormatJSON == OutputFormatYAML)
	
	// Test with variables
	format1 := OutputFormatText
	format2 := OutputFormatText
	format3 := OutputFormatJSON
	
	assert.True(t, format1 == format2)
	assert.False(t, format1 == format3)
}

func TestOutputFormatCasting(t *testing.T) {
	// Test casting from string
	strFormat := "json"
	format := OutputFormat(strFormat)
	
	assert.Equal(t, OutputFormatJSON, format)
	
	// Test casting to string
	str := string(OutputFormatYAML)
	assert.Equal(t, "yaml", str)
}

func TestOutputFormatMap(t *testing.T) {
	// Test that constants can be used as map keys
	formatDescriptions := map[OutputFormat]string{
		OutputFormatText: "Plain text output",
		OutputFormatJSON: "JSON formatted output",
		OutputFormatYAML: "YAML formatted output",
	}
	
	assert.Equal(t, "Plain text output", formatDescriptions[OutputFormatText])
	assert.Equal(t, "JSON formatted output", formatDescriptions[OutputFormatJSON])
	assert.Equal(t, "YAML formatted output", formatDescriptions[OutputFormatYAML])
	
	// Test that non-existent key returns zero value
	assert.Equal(t, "", formatDescriptions[OutputFormat("xml")])
}

func TestOutputFormatCustom(t *testing.T) {
	// Test that we can create custom OutputFormat values
	customFormat := OutputFormat("custom")
	
	assert.Equal(t, "custom", string(customFormat))
	assert.NotEqual(t, OutputFormatText, customFormat)
	assert.NotEqual(t, OutputFormatJSON, customFormat)
	assert.NotEqual(t, OutputFormatYAML, customFormat)
}

// Benchmark tests
func BenchmarkOutputFormatString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = string(OutputFormatJSON)
	}
}

func BenchmarkOutputFormatComparison(b *testing.B) {
	format1 := OutputFormatText
	format2 := OutputFormatJSON
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = format1 == format2
	}
}

func BenchmarkOutputFormatSwitch(b *testing.B) {
	format := OutputFormatJSON
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		switch format {
		case OutputFormatText:
			// Do nothing
		case OutputFormatJSON:
			// Do nothing
		case OutputFormatYAML:
			// Do nothing
		}
	}
}