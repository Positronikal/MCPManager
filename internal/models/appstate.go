package models

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// UserPreferences contains user-configurable preferences
type UserPreferences struct {
	Theme                 string `json:"theme"` // "dark" or "light"
	LogRetentionPerServer int    `json:"logRetentionPerServer"`
	AutoStartServers      bool   `json:"autoStartServers"`
	MinimizeToTray        bool   `json:"minimizeToTray"`
	ShowNotifications     bool   `json:"showNotifications"`
}

// WindowLayout stores the window position and size
type WindowLayout struct {
	Width          int  `json:"width"`
	Height         int  `json:"height"`
	X              int  `json:"x"`
	Y              int  `json:"y"`
	Maximized      bool `json:"maximized"`
	LogPanelHeight int  `json:"logPanelHeight"`
}

// Filters stores the current UI filter state
type Filters struct {
	SelectedServer   string      `json:"selectedServer,omitempty"`   // Server ID
	SelectedSeverity LogSeverity `json:"selectedSeverity,omitempty"` // Log severity filter
	SearchQuery      string      `json:"searchQuery,omitempty"`
}

// ApplicationState represents the complete application state
type ApplicationState struct {
	Version              string           `json:"version"`
	LastSaved            time.Time        `json:"lastSaved"`
	Preferences          UserPreferences  `json:"preferences"`
	WindowLayout         WindowLayout     `json:"windowLayout"`
	Filters              Filters          `json:"filters"`
	DiscoveredServers    []string         `json:"discoveredServers"` // List of server IDs
	MonitoredConfigPaths []string         `json:"monitoredConfigPaths"`
	LastDiscoveryScan    time.Time        `json:"lastDiscoveryScan"`
}

// NewApplicationState creates a new ApplicationState with default values
func NewApplicationState() *ApplicationState {
	return &ApplicationState{
		Version:   "1.0.0",
		LastSaved: time.Now(),
		Preferences: UserPreferences{
			Theme:                 "dark",
			LogRetentionPerServer: 1000,
			AutoStartServers:      false,
			MinimizeToTray:        true,
			ShowNotifications:     true,
		},
		WindowLayout: WindowLayout{
			Width:          1024,
			Height:         768,
			X:              100,
			Y:              100,
			Maximized:      false,
			LogPanelHeight: 300,
		},
		Filters: Filters{
			SelectedServer:   "",
			SelectedSeverity: "",
			SearchQuery:      "",
		},
		DiscoveredServers:    []string{},
		MonitoredConfigPaths: []string{},
		LastDiscoveryScan:    time.Now(),
	}
}

// Validate checks if the ApplicationState is valid
func (s *ApplicationState) Validate() error {
	// Validate window layout dimensions
	if s.WindowLayout.Width < 640 {
		return fmt.Errorf("window width must be at least 640, got: %d", s.WindowLayout.Width)
	}
	if s.WindowLayout.Height < 480 {
		return fmt.Errorf("window height must be at least 480, got: %d", s.WindowLayout.Height)
	}

	// Validate log panel height
	if s.WindowLayout.LogPanelHeight < 0 {
		return fmt.Errorf("log panel height cannot be negative, got: %d", s.WindowLayout.LogPanelHeight)
	}
	if s.WindowLayout.LogPanelHeight > s.WindowLayout.Height {
		return fmt.Errorf("log panel height (%d) cannot exceed window height (%d)",
			s.WindowLayout.LogPanelHeight, s.WindowLayout.Height)
	}

	// Validate theme
	if s.Preferences.Theme != "dark" && s.Preferences.Theme != "light" {
		return fmt.Errorf("theme must be 'dark' or 'light', got: %s", s.Preferences.Theme)
	}

	// Validate log retention
	if s.Preferences.LogRetentionPerServer < 100 || s.Preferences.LogRetentionPerServer > 10000 {
		return fmt.Errorf("logRetentionPerServer must be between 100 and 10000, got: %d",
			s.Preferences.LogRetentionPerServer)
	}

	// Validate discovered servers are UUIDs
	for i, serverID := range s.DiscoveredServers {
		if _, err := uuid.Parse(serverID); err != nil {
			return fmt.Errorf("discoveredServers[%d] is not a valid UUID: %s", i, serverID)
		}
	}

	// Validate monitored config paths are absolute
	for i, path := range s.MonitoredConfigPaths {
		if !filepath.IsAbs(path) {
			return fmt.Errorf("monitoredConfigPaths[%d] is not an absolute path: %s", i, path)
		}
	}

	// Validate selected severity if set
	if s.Filters.SelectedSeverity != "" && !s.Filters.SelectedSeverity.IsValid() {
		return fmt.Errorf("invalid selected severity: %s", s.Filters.SelectedSeverity)
	}

	// Validate selected server is a UUID if set
	if s.Filters.SelectedServer != "" {
		if _, err := uuid.Parse(s.Filters.SelectedServer); err != nil {
			return fmt.Errorf("selected server is not a valid UUID: %s", s.Filters.SelectedServer)
		}
	}

	return nil
}

// AddDiscoveredServer adds a server ID to the discovered servers list
func (s *ApplicationState) AddDiscoveredServer(serverID string) error {
	// Validate UUID
	if _, err := uuid.Parse(serverID); err != nil {
		return fmt.Errorf("invalid server ID: %w", err)
	}

	// Check if already exists
	for _, id := range s.DiscoveredServers {
		if id == serverID {
			return nil // Already exists
		}
	}

	s.DiscoveredServers = append(s.DiscoveredServers, serverID)
	return nil
}

// RemoveDiscoveredServer removes a server ID from the discovered servers list
func (s *ApplicationState) RemoveDiscoveredServer(serverID string) {
	for i, id := range s.DiscoveredServers {
		if id == serverID {
			s.DiscoveredServers = append(s.DiscoveredServers[:i], s.DiscoveredServers[i+1:]...)
			return
		}
	}
}

// AddMonitoredPath adds a config path to the monitored paths list
func (s *ApplicationState) AddMonitoredPath(path string) error {
	// Validate absolute path
	if !filepath.IsAbs(path) {
		return fmt.Errorf("path must be absolute: %s", path)
	}

	// Check if already exists
	for _, p := range s.MonitoredConfigPaths {
		if p == path {
			return nil // Already exists
		}
	}

	s.MonitoredConfigPaths = append(s.MonitoredConfigPaths, path)
	return nil
}
