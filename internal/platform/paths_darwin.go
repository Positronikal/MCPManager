//go:build darwin

package platform

import (
	"os"
	"path/filepath"
)

// DarwinPathResolver implements PathResolver for macOS
type DarwinPathResolver struct{}

// NewPathResolver creates a new platform-specific PathResolver
func NewPathResolver() PathResolver {
	return &DarwinPathResolver{}
}

// GetConfigDir returns ~/Library/Application Support on macOS
func (r *DarwinPathResolver) GetConfigDir() string {
	home := r.GetUserHomeDir()
	if home == "" {
		return ""
	}
	return filepath.Join(home, "Library", "Application Support")
}

// GetAppDataDir returns ~/Library/Application Support on macOS
func (r *DarwinPathResolver) GetAppDataDir() string {
	home := r.GetUserHomeDir()
	if home == "" {
		return ""
	}
	return filepath.Join(home, "Library", "Application Support")
}

// GetUserHomeDir returns the user's home directory
func (r *DarwinPathResolver) GetUserHomeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	return ""
}

// GetMCPManagerDir returns the MCP Manager config directory (~/.mcpmanager)
func GetMCPManagerDir() string {
	resolver := NewPathResolver()
	home := resolver.GetUserHomeDir()
	if home == "" {
		return ""
	}
	return filepath.Join(home, ".mcpmanager")
}
