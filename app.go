package main

import (
	"context"

	"github.com/hoytech/mcpmanager/internal/api"
	"github.com/hoytech/mcpmanager/internal/models"
)

// App struct contains references to backend services
type App struct {
	ctx      context.Context
	services *api.Services
}

// NewApp creates a new App application struct
func NewApp(services *api.Services) *App {
	return &App{
		services: services,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// ListServers returns all discovered servers
func (a *App) ListServers() ([]models.MCPServer, error) {
	servers, _, err := a.services.DiscoveryService.GetServers()
	return servers, err
}

// GetServer returns a specific server by ID
func (a *App) GetServer(serverID string) (*models.MCPServer, error) {
	server, exists := a.services.DiscoveryService.GetServerByID(serverID)
	if !exists {
		return nil, nil
	}
	return server, nil
}

// DiscoverServers triggers a manual discovery scan
func (a *App) DiscoverServers() error {
	_, err := a.services.DiscoveryService.Discover()
	return err
}

// StartServer starts a server by ID
func (a *App) StartServer(serverID string) error {
	server, exists := a.services.DiscoveryService.GetServerByID(serverID)
	if !exists {
		return nil
	}
	return a.services.LifecycleService.StartServer(server)
}

// StopServer stops a server by ID
func (a *App) StopServer(serverID string, force bool, timeout int) error {
	server, exists := a.services.DiscoveryService.GetServerByID(serverID)
	if !exists {
		return nil
	}
	return a.services.LifecycleService.StopServer(server, force, timeout)
}

// RestartServer restarts a server by ID
func (a *App) RestartServer(serverID string) error {
	server, exists := a.services.DiscoveryService.GetServerByID(serverID)
	if !exists {
		return nil
	}
	return a.services.LifecycleService.RestartServer(server)
}

// GetServerConfiguration returns configuration for a server
func (a *App) GetServerConfiguration(serverID string) (*models.ServerConfiguration, error) {
	return a.services.ConfigService.GetConfiguration(serverID)
}

// UpdateServerConfiguration updates configuration for a server
func (a *App) UpdateServerConfiguration(serverID string, config models.ServerConfiguration) error {
	return a.services.ConfigService.UpdateConfiguration(serverID, &config)
}

// GetServerLogs returns logs for a server
func (a *App) GetServerLogs(serverID string, offset, limit int) []models.LogEntry {
	return a.services.MonitoringService.GetLogs(serverID, offset, limit)
}

// GetApplicationState returns the application state
func (a *App) GetApplicationState() (*models.ApplicationState, error) {
	return a.services.StorageService.LoadState()
}

// SaveApplicationState saves the application state
func (a *App) SaveApplicationState(state models.ApplicationState) error {
	return a.services.StorageService.SaveState(&state)
}
