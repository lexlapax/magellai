// ABOUTME: Provides factory functions for creating REPL instances
// ABOUTME: Implements dependency injection for REPL creation

package replapi

// DefaultREPLFactory holds the default REPL factory function
var DefaultREPLFactory Factory

// NewREPL creates a new REPL instance using the registered factory
func NewREPL(opts *REPLOptions) (REPL, error) {
	if DefaultREPLFactory == nil {
		panic("no REPL factory registered")
	}
	return DefaultREPLFactory(opts)
}

// RegisterREPLFactory registers a factory function for creating REPLs
func RegisterREPLFactory(factory Factory) {
	DefaultREPLFactory = factory
}
