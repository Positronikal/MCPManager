//go:build windows

package platform

import (
	"os"
	"path/filepath"
)

// WindowsPathResolver implements PathResolver for Windows
type WindowsPathResolver struct{}

// NewPathResolver creates a new platform-specific PathResolver
func NewPathResolver() PathResolver {
	return &WindowsPathResolver{}
}

// GetConfigDir returns %APPDATA% on Windows
func (r *WindowsPathResolver) GetConfigDir() string {
	if appData := os.Getenv("APPDATA"); appData != "" {
		return appData
	}
	// Fallback to user profile
	if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
		return filepath.Join(userProfile, "AppData", "Roaming")
	}
	return ""
}

// GetAppDataDir returns %LOCALAPPDATA% on Windows
func (r *WindowsPathResolver) GetAppDataDir() string {
	if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
		return localAppData
	}
	// Fallback to user profile
	if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
		return filepath.Join(userProfile, "AppData", "Local")
	}
	return ""
}

// GetUserHomeDir returns the user's home directory
func (r *WindowsPathResolver) GetUserHomeDir() string {
	if home := os.Getenv("USERPROFILE"); home != "" {
		return home
	}
	// Fallback
	if homeDrive := os.Getenv("HOMEDRIVE"); homeDrive != "" {
		if homePath := os.Getenv("HOMEPATH"); homePath != "" {
			return filepath.Join(homeDrive, homePath)
		}
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
