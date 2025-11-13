package models

import "fmt"

// StatusState represents the lifecycle state of an MCP server
type StatusState string

const (
	StatusStopped  StatusState = "stopped"
	StatusStarting StatusState = "starting"
	StatusRunning  StatusState = "running"
	StatusError    StatusState = "error"
)

// ValidStatusStates contains all valid status states
var ValidStatusStates = []StatusState{
	StatusStopped,
	StatusStarting,
	StatusRunning,
	StatusError,
}

// IsValid validates if the status state is valid
func (s StatusState) IsValid() bool {
	for _, valid := range ValidStatusStates {
		if s == valid {
			return true
		}
	}
	return false
}

// LogSeverity represents the severity level of a log entry
type LogSeverity string

const (
	LogInfo    LogSeverity = "info"
	LogSuccess LogSeverity = "success"
	LogWarning LogSeverity = "warning"
	LogError   LogSeverity = "error"
)

// ValidLogSeverities contains all valid log severities
var ValidLogSeverities = []LogSeverity{
	LogInfo,
	LogSuccess,
	LogWarning,
	LogError,
}

// IsValid validates if the log severity is valid
func (l LogSeverity) IsValid() bool {
	for _, valid := range ValidLogSeverities {
		if l == valid {
			return true
		}
	}
	return false
}

// DependencyType represents the type of dependency
type DependencyType string

const (
	DependencyRuntime     DependencyType = "runtime"
	DependencyLibrary     DependencyType = "library"
	DependencyTool        DependencyType = "tool"
	DependencyEnvironment DependencyType = "environment"
)

// ValidDependencyTypes contains all valid dependency types
var ValidDependencyTypes = []DependencyType{
	DependencyRuntime,
	DependencyLibrary,
	DependencyTool,
	DependencyEnvironment,
}

// IsValid validates if the dependency type is valid
func (d DependencyType) IsValid() bool {
	for _, valid := range ValidDependencyTypes {
		if d == valid {
			return true
		}
	}
	return false
}

// DiscoverySource represents where a server was discovered from
type DiscoverySource string

const (
	DiscoveryClientConfig DiscoverySource = "client_config"
	DiscoveryExtension    DiscoverySource = "extension"
	DiscoveryFilesystem   DiscoverySource = "filesystem"
	DiscoveryProcess      DiscoverySource = "process"
)

// ValidDiscoverySources contains all valid discovery sources
var ValidDiscoverySources = []DiscoverySource{
	DiscoveryClientConfig,
	DiscoveryExtension,
	DiscoveryFilesystem,
	DiscoveryProcess,
}

// IsValid validates if the discovery source is valid
func (d DiscoverySource) IsValid() bool {
	for _, valid := range ValidDiscoverySources {
		if d == valid {
			return true
		}
	}
	return false
}

// ValidateEnum validates any enum type
func ValidateEnum(value interface{}, validValues []interface{}, fieldName string) error {
	for _, valid := range validValues {
		if value == valid {
			return nil
		}
	}
	return fmt.Errorf("%s has invalid value: %v", fieldName, value)
}
