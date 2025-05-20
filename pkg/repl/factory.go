// ABOUTME: Registers REPL factory with the replapi package
// ABOUTME: Implements the factory pattern for REPL creation

package repl

import (
	"github.com/lexlapax/magellai/pkg/replapi"
)

func init() {
	// Register our factory with the replapi package
	replapi.RegisterREPLFactory(func(opts *replapi.REPLOptions) (replapi.REPL, error) {
		// Convert API options to internal options
		internalOpts := &REPLOptions{
			Config:      opts.Config,
			StorageDir:  opts.StorageDir,
			PromptStyle: opts.PromptStyle,
			SessionID:   opts.SessionID,
			Model:       opts.Model,
			Writer:      opts.Writer,
			Reader:      opts.Reader,
		}

		// Create internal REPL
		repl, err := NewREPL(internalOpts)
		if err != nil {
			return nil, err
		}

		return repl, nil
	})
}
