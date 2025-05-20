// ABOUTME: REPL API interfaces for dependency inversion
// ABOUTME: Breaks circular dependencies between command and repl packages

/*
Package replapi defines interfaces for the REPL component of Magellai.

This package serves as a dependency inversion layer between the command and
repl packages, breaking circular dependencies while providing clean interfaces
for REPL functionality. It contains only interfaces and factory functions,
with the actual implementations residing in the repl package.

Key Components:
  - REPL Interface: Core interface defining the interactive REPL functionality
  - ConfigInterface: Minimal configuration interface needed by REPL
  - REPLOptions: Initialization options for creating REPL instances
  - Factory: Factory function type for dependency injection

The package enables the command system to utilize REPL functionality without
directly importing the repl package, establishing a clean dependency direction:
command -> replapi <- repl

Usage:
    // Register a REPL factory
    replapi.RegisterREPLFactory(func(opts *replapi.REPLOptions) (replapi.REPL, error) {
        return myreplimplementation.NewREPL(opts)
    })

    // Create a new REPL using the registered factory
    repl, err := replapi.NewREPL(&replapi.REPLOptions{
        Config: config,
        Writer: os.Stdout,
        Reader: os.Stdin,
    })

This package is part of the domain-driven design pattern used in Magellai,
facilitating proper separation of concerns and dependency management.
*/
package replapi