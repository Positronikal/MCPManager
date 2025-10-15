//go:build linux

package platform

import (
	"os"
	"path/filepath"
)

// LinuxPathResolver implements PathResolver for Linux
type LinuxPathResolver struct{}

// NewPathResolver creates a new platform-specific PathResolver
func NewPathResolver() PathResolver {
	return &LinuxPathResolver{}
}

// GetConfigDir returns ~/.config on Linux
func (r *LinuxPathResolver) GetConfigDir() string {
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return xdgConfig
	}
	home := r.GetUserHomeDir()
	if home == "" {
		return ""
	}
	return filepath.Join(home, ".config")
}

// GetAppDataDir returns ~/.local/share on Linux
func (r *LinuxPathResolver) GetAppDataDir() string {
	if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
		return xdgData
	}
	home := r.GetUserHomeDir()
	if home == "" {
		return ""
	}
	return filepath.Join(home, ".local", "share")
}

// GetUserHomeDir returns the user's home directory
func (r *LinuxPathResolver) GetUserHomeDir() string {
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
