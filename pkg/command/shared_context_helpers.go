// ABOUTME: Helper methods for SharedContext to provide convenient access
// ABOUTME: Adds methods like Model(), Temperature() for easier usage

package command

// Model returns the current model setting
func (sc *SharedContext) Model() string {
	if model, ok := sc.GetString(SharedContextModel); ok {
		return model
	}
	return ""
}

// SetModel sets the model
func (sc *SharedContext) SetModel(model string) {
	sc.Set(SharedContextModel, model)
}

// Temperature returns the current temperature setting
func (sc *SharedContext) Temperature() float64 {
	if temp, ok := sc.GetFloat64(SharedContextTemperature); ok {
		return temp
	}
	return -1
}

// SetTemperature sets the temperature
func (sc *SharedContext) SetTemperature(temp float64) {
	sc.Set(SharedContextTemperature, temp)
}

// MaxTokens returns the current max tokens setting
func (sc *SharedContext) MaxTokens() int {
	if tokens, ok := sc.GetInt(SharedContextMaxTokens); ok {
		return tokens
	}
	return -1
}

// SetMaxTokens sets the max tokens
func (sc *SharedContext) SetMaxTokens(tokens int) {
	sc.Set(SharedContextMaxTokens, tokens)
}

// Stream returns the current stream setting
func (sc *SharedContext) Stream() bool {
	if stream, ok := sc.GetBool(SharedContextStream); ok {
		return stream
	}
	return false
}

// SetStream sets the stream setting
func (sc *SharedContext) SetStream(stream bool) {
	sc.Set(SharedContextStream, stream)
}

// Verbose returns the current verbose setting
func (sc *SharedContext) Verbose() bool {
	// Check if verbosity is set to "verbose" or higher
	if verbosity, ok := sc.GetString(SharedContextVerbosity); ok {
		return verbosity == "debug" || verbosity == "info"
	}
	return false
}

// SetVerbose sets the verbose setting
func (sc *SharedContext) SetVerbose(verbose bool) {
	if verbose {
		sc.Set(SharedContextVerbosity, "info")
	} else {
		sc.Set(SharedContextVerbosity, "error")
	}
}

// Debug returns the current debug setting
func (sc *SharedContext) Debug() bool {
	if verbosity, ok := sc.GetString(SharedContextVerbosity); ok {
		return verbosity == "debug"
	}
	return false
}

// SetDebug sets the debug setting
func (sc *SharedContext) SetDebug(debug bool) {
	if debug {
		sc.Set(SharedContextVerbosity, "debug")
	} else {
		sc.Set(SharedContextVerbosity, "info")
	}
}
