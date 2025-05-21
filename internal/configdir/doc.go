// ABOUTME: Internal package for determining configuration directories
// ABOUTME: Handles platform-specific config and data directories

/*
Package configdir provides utilities for determining configuration directories.

This package handles the platform-specific logic for locating configuration and
data directories, following OS-specific conventions for application data storage.
It ensures that configuration files and session data are stored in appropriate
locations across different operating systems.

Key Components:
  - GetConfigDir: Determines the configuration directory
  - GetCacheDir: Locates the cache directory
  - GetDataDir: Identifies the data directory
  - EnsureDir: Creates directories if they don't exist

The package handles the following platform conventions:
  - macOS: ~/Library/Application Support/Magellai
  - Linux: ~/.config/magellai
  - Windows: %APPDATA%\Magellai

Usage:

	// Get the configuration directory
	configDir, err := configdir.GetConfigDir()
	if err != nil {
	    // Handle error
	}

	// Ensure a directory exists
	dataDir, err := configdir.GetDataDir()
	err = configdir.EnsureDir(dataDir)

This package is internal to the application and should not be imported by
external code. Configuration paths should be accessed through the config package.
*/
package configdir
