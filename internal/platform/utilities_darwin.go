//go:build darwin
// +build darwin

package platform

import (
	"os/exec"
)

// OpenFileExplorer opens Finder at the specified path
func OpenFileExplorer(path string) error {
	cmd := exec.Command("open", path)
	return cmd.Start()
}

// LaunchShell opens Terminal.app
func LaunchShell() error {
	cmd := exec.Command("open", "-a", "Terminal")
	return cmd.Start()
}
