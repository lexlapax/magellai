// ABOUTME: ANSI color output support for TTY environments
// ABOUTME: Provides color formatting for CLI and REPL output when terminal supports it

package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/lexlapax/magellai/internal/logging"
)

// ANSI color codes
const (
	ColorReset  = "\033[0m"
	ColorBold   = "\033[1m"
	ColorDim    = "\033[2m"
	ColorItalic = "\033[3m"

	ColorBlack   = "\033[30m"
	ColorRed     = "\033[31m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorBlue    = "\033[34m"
	ColorMagenta = "\033[35m"
	ColorCyan    = "\033[36m"
	ColorWhite   = "\033[37m"

	ColorBrightBlack   = "\033[90m"
	ColorBrightRed     = "\033[91m"
	ColorBrightGreen   = "\033[92m"
	ColorBrightYellow  = "\033[93m"
	ColorBrightBlue    = "\033[94m"
	ColorBrightMagenta = "\033[95m"
	ColorBrightCyan    = "\033[96m"
	ColorBrightWhite   = "\033[97m"

	// Background colors
	ColorBgBlack   = "\033[40m"
	ColorBgRed     = "\033[41m"
	ColorBgGreen   = "\033[42m"
	ColorBgYellow  = "\033[43m"
	ColorBgBlue    = "\033[44m"
	ColorBgMagenta = "\033[45m"
	ColorBgCyan    = "\033[46m"
	ColorBgWhite   = "\033[47m"
)

// ColorTheme defines the color scheme for REPL output
type ColorTheme struct {
	Command          string
	CommandPrefix    string
	Error            string
	Warning          string
	Success          string
	Info             string
	UserMessage      string
	AssistantMessage string
	SystemMessage    string
	CodeBlock        string
	Highlight        string
}

// DefaultColorTheme returns the default color theme
func DefaultColorTheme() *ColorTheme {
	return &ColorTheme{
		Command:          ColorBrightCyan,
		CommandPrefix:    ColorBrightMagenta,
		Error:            ColorRed,
		Warning:          ColorYellow,
		Success:          ColorGreen,
		Info:             ColorBlue,
		UserMessage:      ColorBrightWhite,
		AssistantMessage: ColorWhite,
		SystemMessage:    ColorBrightBlack,
		CodeBlock:        ColorDim,
		Highlight:        ColorBrightYellow,
	}
}

// ColorFormatter handles color formatting for REPL output
type ColorFormatter struct {
	enabled bool
	theme   *ColorTheme
}

// NewColorFormatter creates a new color formatter
func NewColorFormatter(enabled bool, theme *ColorTheme) *ColorFormatter {
	if theme == nil {
		theme = DefaultColorTheme()
	}

	logging.LogDebug("Creating color formatter", "enabled", enabled)

	return &ColorFormatter{
		enabled: enabled,
		theme:   theme,
	}
}

// Enabled returns whether colors are enabled
func (f *ColorFormatter) Enabled() bool {
	return f.enabled
}

// SetEnabled sets whether colors are enabled
func (f *ColorFormatter) SetEnabled(enabled bool) {
	logging.LogDebug("Color output changed", "enabled", enabled)
	f.enabled = enabled
}

// Format applies color formatting if enabled
func (f *ColorFormatter) Format(color, text string) string {
	if !f.enabled {
		return text
	}
	return color + text + ColorReset
}

// FormatCommand formats a command
func (f *ColorFormatter) FormatCommand(text string) string {
	return f.Format(f.theme.Command, text)
}

// FormatCommandPrefix formats a command prefix (/, :)
func (f *ColorFormatter) FormatCommandPrefix(text string) string {
	return f.Format(f.theme.CommandPrefix, text)
}

// FormatError formats an error message
func (f *ColorFormatter) FormatError(text string) string {
	return f.Format(f.theme.Error, text)
}

// FormatWarning formats a warning message
func (f *ColorFormatter) FormatWarning(text string) string {
	return f.Format(f.theme.Warning, text)
}

// FormatSuccess formats a success message
func (f *ColorFormatter) FormatSuccess(text string) string {
	return f.Format(f.theme.Success, text)
}

// FormatInfo formats an info message
func (f *ColorFormatter) FormatInfo(text string) string {
	return f.Format(f.theme.Info, text)
}

// FormatUserMessage formats a user message
func (f *ColorFormatter) FormatUserMessage(text string) string {
	return f.Format(f.theme.UserMessage, text)
}

// FormatAssistantMessage formats an assistant message
func (f *ColorFormatter) FormatAssistantMessage(text string) string {
	return f.Format(f.theme.AssistantMessage, text)
}

// FormatSystemMessage formats a system message
func (f *ColorFormatter) FormatSystemMessage(text string) string {
	return f.Format(f.theme.SystemMessage, text)
}

// FormatCodeBlock formats a code block
func (f *ColorFormatter) FormatCodeBlock(text string) string {
	return f.Format(f.theme.CodeBlock, text)
}

// FormatHighlight formats highlighted text
func (f *ColorFormatter) FormatHighlight(text string) string {
	return f.Format(f.theme.Highlight, text)
}

// FormatPrompt formats the REPL prompt
func (f *ColorFormatter) FormatPrompt(prompt string) string {
	if !f.enabled {
		return prompt
	}

	// Make the prompt bold and colored
	return f.Format(ColorBold+f.theme.Info, prompt)
}

// FormatCommandHelp formats command help text with colors
func (f *ColorFormatter) FormatCommandHelp(name, description string) string {
	if !f.enabled {
		return fmt.Sprintf("  %-20s %s", name, description)
	}

	coloredName := f.Format(f.theme.Command, name)
	coloredDesc := f.Format(ColorDim, description)
	return fmt.Sprintf("  %-20s %s", coloredName, coloredDesc)
}

// StripColors removes ANSI color codes from text
func StripColors(text string) string {
	var result strings.Builder
	i := 0

	for i < len(text) {
		if i+1 < len(text) && text[i] == '\033' && text[i+1] == '[' {
			// Potential ANSI escape sequence
			start := i
			i += 2

			// Valid sequence must have either:
			// 1. At least one parameter then command, or
			// 2. Just a command character (some sequences like \033[m are valid)
			foundParam := false
			foundValid := false

			// Skip parameters (numbers and semicolons)
			for i < len(text) && (text[i] >= '0' && text[i] <= '9' || text[i] == ';') {
				foundParam = true
				i++
			}

			// Check for command character
			if i < len(text) && ((text[i] >= 'A' && text[i] <= 'Z') || (text[i] >= 'a' && text[i] <= 'z')) {
				// Valid ANSI sequence only if we had params or it's a special single-letter command
				ch := text[i]
				// Common ANSI commands that can appear without parameters
				if foundParam || ch == 'm' || ch == 'H' || ch == 'J' || ch == 'K' || ch == 'h' || ch == 'l' {
					foundValid = true
					i++ // Skip the command character
				}
			}

			if !foundValid {
				// Not a valid sequence, write from start and continue
				result.WriteString(text[start:])
				return result.String()
			}
			// If valid, we've already skipped past it
		} else {
			result.WriteByte(text[i])
			i++
		}
	}

	return result.String()
}

// IsTerminal checks if we're running in a terminal
func IsTerminal() bool {
	fileInfo, _ := os.Stdin.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
