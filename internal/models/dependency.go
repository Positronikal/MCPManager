package models

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// Dependency represents a dependency required by an MCP server
type Dependency struct {
	Name                    string         `json:"name"`
	Type                    DependencyType `json:"type"`
	RequiredVersion         string         `json:"requiredVersion,omitempty"`
	DetectedVersion         string         `json:"detectedVersion,omitempty"`
	InstallationInstructions string        `json:"installationInstructions,omitempty"` // Markdown format
}

// IsInstalled checks if the dependency is installed and satisfies the version requirement
func (d *Dependency) IsInstalled() bool {
	// If no detected version, not installed
	if d.DetectedVersion == "" {
		return false
	}

	// If no required version, any version is acceptable
	if d.RequiredVersion == "" {
		return true
	}

	// Try to parse as semver constraint
	constraint, err := semver.NewConstraint(d.RequiredVersion)
	if err != nil {
		// If not a valid semver constraint, do string comparison
		return strings.EqualFold(d.DetectedVersion, d.RequiredVersion)
	}

	// Parse detected version
	version, err := semver.NewVersion(d.DetectedVersion)
	if err != nil {
		// If detected version is not valid semver, do string comparison
		return strings.EqualFold(d.DetectedVersion, d.RequiredVersion)
	}

	// Check if version satisfies constraint
	return constraint.Check(version)
}

// Validate checks if the Dependency is valid
func (d *Dependency) Validate() error {
	if d.Name == "" {
		return fmt.Errorf("dependency name cannot be empty")
	}

	if !d.Type.IsValid() {
		return fmt.Errorf("invalid dependency type: %s", d.Type)
	}

	// If a required version is specified, try to validate it as a semver constraint
	if d.RequiredVersion != "" {
		// Try to parse as constraint - if it fails, it's okay (might be exact version)
		_, err := semver.NewConstraint(d.RequiredVersion)
		if err != nil {
			// Try as exact version
			_, err := semver.NewVersion(d.RequiredVersion)
			if err != nil {
				// Neither constraint nor version format - might be custom format, allow it
			}
		}
	}

	return nil
}
