// ABOUTME: Path manipulation utilities for working with file and config paths
// ABOUTME: Provides common path operations for extracting names, extensions, and expanding user paths

package stringutil

import (
	"os"
	"path/filepath"
	"strings"
)

// ExpandPath expands ~ to user home directory in a file path
func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

// ExtractFileName returns the filename from a path, with or without extension
func ExtractFileName(path string, withExtension bool) string {
	if path == "" {
		return path
	}

	fileName := filepath.Base(path)
	if withExtension {
		return fileName
	}

	// Remove extension
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

// ExtractExtension returns the extension from a file path
func ExtractExtension(path string) string {
	if path == "" {
		return ""
	}

	return filepath.Ext(path)
}

// JoinConfigPath joins config key parts with dots
func JoinConfigPath(parts ...string) string {
	return strings.Join(parts, ".")
}

// SplitConfigPath splits a config key path into its component parts
func SplitConfigPath(path string) []string {
	if path == "" {
		return []string{}
	}

	return strings.Split(path, ".")
}

// SafeFilename ensures a string can be safely used as a filename
// It replaces potentially problematic characters with underscores
func SafeFilename(name string) string {
	// Replace any character that's not alphanumeric, dash, or underscore
	r := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		".", "_",
	)

	return r.Replace(name)
}
