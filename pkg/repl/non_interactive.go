// ABOUTME: Non-interactive mode detection for REPL operations
// ABOUTME: Handles detection of pipes, background execution, and non-TTY environments

package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/lexlapax/magellai/internal/logging"
)

// NonInteractiveMode represents the various non-interactive states
type NonInteractiveMode struct {
	IsNonInteractive bool // Overall non-interactive flag
	IsPipedInput     bool // stdin is piped
	IsPipedOutput    bool // stdout is piped
	IsPipedError     bool // stderr is piped
	IsBackground     bool // Running as background process
	IsNotTTY         bool // Not a terminal
	IsCIEnvironment  bool // Running in CI/CD
}

// DetectNonInteractiveMode checks for various non-interactive conditions
func DetectNonInteractiveMode(reader io.Reader, writer io.Writer) NonInteractiveMode {
	mode := NonInteractiveMode{}

	// Check if running in a terminal - check both stdin and stdout
	stdinStat, stdinErr := os.Stdin.Stat()
	stdinIsTerminal := stdinErr == nil && (stdinStat.Mode()&os.ModeCharDevice) != 0

	stdoutStat, stdoutErr := os.Stdout.Stat()
	stdoutIsTerminal := stdoutErr == nil && (stdoutStat.Mode()&os.ModeCharDevice) != 0

	mode.IsNotTTY = !stdinIsTerminal || !stdoutIsTerminal

	// Check for piped input
	if file, ok := reader.(*os.File); ok {
		stat, err := file.Stat()
		if err == nil {
			mode.IsPipedInput = (stat.Mode() & os.ModeCharDevice) == 0
		}
	} else if reader != os.Stdin {
		// Custom reader indicates non-interactive
		mode.IsPipedInput = true
	}

	// Check for piped output
	if file, ok := writer.(*os.File); ok {
		stat, err := file.Stat()
		if err == nil {
			mode.IsPipedOutput = (stat.Mode() & os.ModeCharDevice) == 0
		}
	} else if writer != os.Stdout {
		// Custom writer indicates non-interactive
		mode.IsPipedOutput = true
	}

	// Check stderr
	stat, err := os.Stderr.Stat()
	if err == nil {
		mode.IsPipedError = (stat.Mode() & os.ModeCharDevice) == 0
	}

	// Check for CI environment variables
	mode.IsCIEnvironment = checkCIEnvironment()

	// Check if running as background process
	mode.IsBackground = isBackgroundProcess()

	// Overall non-interactive flag
	mode.IsNonInteractive = mode.IsNotTTY || mode.IsPipedInput ||
		mode.IsPipedOutput || mode.IsBackground || mode.IsCIEnvironment

	logging.LogDebug("Non-interactive mode detection",
		"isNonInteractive", mode.IsNonInteractive,
		"isNotTTY", mode.IsNotTTY,
		"isPipedInput", mode.IsPipedInput,
		"isPipedOutput", mode.IsPipedOutput,
		"isPipedError", mode.IsPipedError,
		"isBackground", mode.IsBackground,
		"isCIEnvironment", mode.IsCIEnvironment)

	return mode
}

// checkCIEnvironment checks for common CI environment variables
func checkCIEnvironment() bool {
	ciVars := []string{
		"CI",
		"CONTINUOUS_INTEGRATION",
		"GITHUB_ACTIONS",
		"GITLAB_CI",
		"JENKINS_URL",
		"CIRCLECI",
		"TRAVIS",
		"BUILDKITE",
		"DRONE",
		"TEAMCITY_VERSION",
	}

	for _, v := range ciVars {
		if os.Getenv(v) != "" {
			return true
		}
	}

	return false
}

// isBackgroundProcess attempts to detect if running as background process
func isBackgroundProcess() bool {
	// On Unix systems, check if process group differs from terminal process group
	// This is platform-specific and might need more sophisticated checks
	// For now, don't detect background processes
	return false
}

// ConfigureForNonInteractiveMode adjusts REPL settings for non-interactive use
func (r *REPL) ConfigureForNonInteractiveMode(mode NonInteractiveMode) {
	if !mode.IsNonInteractive {
		return
	}

	logging.LogInfo("Configuring REPL for non-interactive mode")

	// Disable interactive features
	r.isTerminal = false
	r.multiline = false
	r.exitOnEOF = true

	// Disable colors
	if r.colorFormatter != nil {
		r.colorFormatter.SetEnabled(false)
	}

	// Disable readline/tab completion
	if r.readline != nil {
		r.readline = nil
	}

	// Set simple prompt for non-interactive mode
	if mode.IsPipedInput || mode.IsPipedOutput {
		r.promptStyle = "" // No prompt when piped
	} else {
		r.promptStyle = "$ " // Simple prompt for other non-interactive cases
	}

	// Adjust auto-save behavior
	if mode.IsCIEnvironment || mode.IsBackground {
		// More aggressive saving in CI/background
		r.autoSave = true
	}

	logging.LogDebug("Non-interactive configuration applied",
		"isTerminal", r.isTerminal,
		"multiline", r.multiline,
		"exitOnEOF", r.exitOnEOF,
		"promptStyle", r.promptStyle,
		"autoSave", r.autoSave)
}

// ShouldAutoExit determines if REPL should exit automatically
func (r *REPL) ShouldAutoExit(mode NonInteractiveMode) bool {
	// Auto-exit when input is piped and we've processed all input
	return mode.IsPipedInput && r.exitOnEOF
}

// ProcessPipedInput handles reading and processing piped input
func (r *REPL) ProcessPipedInput(mode NonInteractiveMode) error {
	if !mode.IsPipedInput {
		return nil
	}

	logging.LogInfo("Processing piped input")

	// Read all input at once for piped mode
	scanner := bufio.NewScanner(r.reader)
	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		logging.LogError(err, "Error reading piped input")
		return fmt.Errorf("failed to read piped input: %w", err)
	}

	// Process all lines as a single input
	input := strings.Join(lines, "\n")
	if input == "" {
		return nil
	}

	// Process as a command or message
	if strings.HasPrefix(input, "/") || strings.HasPrefix(input, ":") {
		return r.handleCommand(input)
	}

	// Process as regular message
	return r.processMessage(input)
}
