// ABOUTME: ANSI color utilities for terminal output formatting
// ABOUTME: Provides wrapper functions for applying and stripping ANSI color codes

package stringutil

import (
	"fmt"
	"regexp"
	"strings"
)

// ANSI color code constants
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Italic    = "\033[3m"
	Underline = "\033[4m"
	Blink     = "\033[5m"
	Reverse   = "\033[7m"
	Hidden    = "\033[8m"

	Black   = "\033[30m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"

	BgBlack   = "\033[40m"
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
	BgWhite   = "\033[47m"
)

var (
	// Regular expression to match ANSI escape sequences
	ansiRegex = regexp.MustCompile("\033\\[[0-9;]*[a-zA-Z]")
)

// ColorMap maps color names to ANSI color codes
var ColorMap = map[string]string{
	"reset":     Reset,
	"bold":      Bold,
	"dim":       Dim,
	"italic":    Italic,
	"underline": Underline,
	"blink":     Blink,
	"reverse":   Reverse,
	"hidden":    Hidden,

	"black":   Black,
	"red":     Red,
	"green":   Green,
	"yellow":  Yellow,
	"blue":    Blue,
	"magenta": Magenta,
	"cyan":    Cyan,
	"white":   White,

	"bg_black":   BgBlack,
	"bg_red":     BgRed,
	"bg_green":   BgGreen,
	"bg_yellow":  BgYellow,
	"bg_blue":    BgBlue,
	"bg_magenta": BgMagenta,
	"bg_cyan":    BgCyan,
	"bg_white":   BgWhite,
}

// ColorText applies the specified color to the given text
func ColorText(color, text string) string {
	colorCode, exists := ColorMap[strings.ToLower(color)]
	if !exists {
		return text
	}

	return fmt.Sprintf("%s%s%s", colorCode, text, Reset)
}

// ColorTextf applies the specified color to a formatted string
func ColorTextf(color, format string, args ...interface{}) string {
	return ColorText(color, fmt.Sprintf(format, args...))
}

// StripColors removes ANSI color codes from a string
func StripColors(text string) string {
	return ansiRegex.ReplaceAllString(text, "")
}

// FormatCommand formats text as a command (typically cyan)
func FormatCommand(text string) string {
	return ColorText("cyan", text)
}

// FormatError formats text as an error (typically red)
func FormatError(text string) string {
	return ColorText("red", text)
}

// FormatWarning formats text as a warning (typically yellow)
func FormatWarning(text string) string {
	return ColorText("yellow", text)
}

// FormatSuccess formats text as a success message (typically green)
func FormatSuccess(text string) string {
	return ColorText("green", text)
}

// FormatHighlight formats text with highlighting (typically magenta or cyan)
func FormatHighlight(text string) string {
	return ColorText("magenta", text)
}

// FormatBold makes text bold
func FormatBold(text string) string {
	return fmt.Sprintf("%s%s%s", Bold, text, Reset)
}

// EnableColors determines if colors should be enabled based on environment
var EnableColors = true

// FormatWithColor conditionally applies color based on EnableColors setting
func FormatWithColor(color, text string) string {
	if !EnableColors {
		return text
	}
	return ColorText(color, text)
}
