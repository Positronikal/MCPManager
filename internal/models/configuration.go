package models

import (
	"fmt"
	"os"
	"regexp"
)

// ServerConfiguration contains the configuration for launching and managing an MCP server
type ServerConfiguration struct {
	EnvironmentVariables map[string]string `json:"environmentVariables,omitempty"`
	CommandLineArguments []string          `json:"commandLineArguments,omitempty"`
	WorkingDirectory     string            `json:"workingDirectory,omitempty"`
	AutoStart            bool              `json:"autoStart"`
	RestartOnCrash       bool              `json:"restartOnCrash"`
	MaxRestartAttempts   int               `json:"maxRestartAttempts"`
	StartupTimeout       int               `json:"startupTimeout"`                // seconds
	ShutdownTimeout      int               `json:"shutdownTimeout"`               // seconds
	HealthCheckInterval  int               `json:"healthCheckInterval,omitempty"` // seconds
	HealthCheckEndpoint  string            `json:"healthCheckEndpoint,omitempty"`
}

// envVarRegex matches valid environment variable names (uppercase letters, digits, underscores)
var envVarRegex = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)

// NewServerConfiguration creates a new ServerConfiguration with default values
func NewServerConfiguration() *ServerConfiguration {
	return &ServerConfiguration{
		EnvironmentVariables: make(map[string]string),
		CommandLineArguments: []string{},
		AutoStart:            false,
		RestartOnCrash:       false,
		MaxRestartAttempts:   3,
		StartupTimeout:       30,
		ShutdownTimeout:      10,
	}
}

// Validate checks if the ServerConfiguration is valid
func (c *ServerConfiguration) Validate() error {
	// Validate environment variable names
	for key := range c.EnvironmentVariables {
		if !envVarRegex.MatchString(key) {
			return fmt.Errorf("invalid environment variable name: %s (must match ^[A-Z_][A-Z0-9_]*$)", key)
		}
	}

	// Validate working directory exists if provided
	if c.WorkingDirectory != "" {
		info, err := os.Stat(c.WorkingDirectory)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("working directory does not exist: %s", c.WorkingDirectory)
			}
			return fmt.Errorf("cannot access working directory: %w", err)
		}
		if !info.IsDir() {
			return fmt.Errorf("working directory is not a directory: %s", c.WorkingDirectory)
		}
	}

	// Validate MaxRestartAttempts is in range 0-10
	if c.MaxRestartAttempts < 0 || c.MaxRestartAttempts > 10 {
		return fmt.Errorf("maxRestartAttempts must be between 0 and 10, got: %d", c.MaxRestartAttempts)
	}

	// Validate timeouts are positive
	if c.StartupTimeout <= 0 {
		return fmt.Errorf("startupTimeout must be positive, got: %d", c.StartupTimeout)
	}
	if c.ShutdownTimeout <= 0 {
		return fmt.Errorf("shutdownTimeout must be positive, got: %d", c.ShutdownTimeout)
	}

	// Validate health check interval if health checking is enabled
	if c.HealthCheckEndpoint != "" && c.HealthCheckInterval <= 0 {
		return fmt.Errorf("healthCheckInterval must be positive when healthCheckEndpoint is set")
	}

	return nil
}
