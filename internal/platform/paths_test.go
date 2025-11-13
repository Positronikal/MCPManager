package platform

import (
	"runtime"
	"strings"
	"testing"
)

func TestPathResolver(t *testing.T) {
	resolver := NewPathResolver()

	t.Run("GetConfigDir", func(t *testing.T) {
		configDir := resolver.GetConfigDir()
		if configDir == "" {
			t.Error("GetConfigDir returned empty string")
		}

		// Verify platform-specific path format
		switch runtime.GOOS {
		case "windows":
			if !strings.Contains(configDir, "AppData") {
				t.Errorf("Windows config dir should contain 'AppData', got: %s", configDir)
			}
		case "darwin":
			if !strings.Contains(configDir, "Library/Application Support") {
				t.Errorf("macOS config dir should contain 'Library/Application Support', got: %s", configDir)
			}
		case "linux":
			if !strings.Contains(configDir, ".config") {
				t.Errorf("Linux config dir should contain '.config', got: %s", configDir)
			}
		}
	})

	t.Run("GetAppDataDir", func(t *testing.T) {
		appDataDir := resolver.GetAppDataDir()
		if appDataDir == "" {
			t.Error("GetAppDataDir returned empty string")
		}
	})

	t.Run("GetUserHomeDir", func(t *testing.T) {
		homeDir := resolver.GetUserHomeDir()
		if homeDir == "" {
			t.Error("GetUserHomeDir returned empty string")
		}
	})
}

func TestGetMCPManagerDir(t *testing.T) {
	mcpDir := GetMCPManagerDir()
	if mcpDir == "" {
		t.Error("GetMCPManagerDir returned empty string")
	}

	if !strings.HasSuffix(mcpDir, ".mcpmanager") {
		t.Errorf("MCP Manager dir should end with '.mcpmanager', got: %s", mcpDir)
	}
}
