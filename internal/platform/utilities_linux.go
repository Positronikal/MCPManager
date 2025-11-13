//go:build linux
// +build linux

package platform

import (
	"os/exec"
)

// OpenFileExplorer opens the default file manager at the specified path
func OpenFileExplorer(path string) error {
	cmd := exec.Command("xdg-open", path)
	return cmd.Start()
}

// LaunchShell opens the default terminal emulator
func LaunchShell() error {
	// Try common terminal emulators
	terminals := []string{"x-terminal-emulator", "gnome-terminal", "xterm"}

	for _, terminal := range terminals {
		cmd := exec.Command(terminal)
		if err := cmd.Start(); err == nil {
			return nil
		}
	}

	return exec.Command("xterm").Start()
}
