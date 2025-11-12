//go:build windows
// +build windows

package platform

import (
	"os/exec"
)

// OpenFileExplorer opens Windows Explorer at the specified path
func OpenFileExplorer(path string) error {
	cmd := exec.Command("explorer", path)
	return cmd.Start()
}

// LaunchShell opens a Command Prompt window
func LaunchShell() error {
	cmd := exec.Command("cmd", "/c", "start", "cmd")
	return cmd.Start()
}
