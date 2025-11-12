package performance

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

// TestMemoryUsageIdle tests memory usage when idle
// Requirement: FR-039 - System MUST consume less than 100MB of memory when idle
func TestMemoryUsageIdle(t *testing.T) {
	// Build the application first
	buildCmd := exec.Command("wails", "build", "-clean")
	buildCmd.Dir = filepath.Join("..", "..")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build application: %v\nOutput: %s", err, output)
	}

	// Determine the executable path
	exePath := filepath.Join("..", "..", "build", "bin", "mcpmanager.exe")
	if _, err := os.Stat(exePath); os.IsNotExist(err) {
		exePath = filepath.Join("..", "..", "build", "bin", "mcpmanager")
	}

	// Start the application
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, exePath)
	cmd.Env = append(os.Environ(), "HEADLESS=true")

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start application: %v", err)
	}
	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	// Wait for application to stabilize
	t.Log("Waiting 10 seconds for application to stabilize...")
	time.Sleep(10 * time.Second)

	// Get process info
	proc, err := process.NewProcess(int32(cmd.Process.Pid))
	if err != nil {
		t.Fatalf("Failed to get process info: %v", err)
	}

	// Measure memory usage
	memInfo, err := proc.MemoryInfo()
	if err != nil {
		t.Fatalf("Failed to get memory info: %v", err)
	}

	// Convert to MB
	rssMB := float64(memInfo.RSS) / 1024 / 1024
	vssMB := float64(memInfo.VMS) / 1024 / 1024

	t.Logf("\n=== Idle Memory Usage ===")
	t.Logf("RSS (Resident Set Size): %.2f MB", rssMB)
	t.Logf("VSS (Virtual Memory):    %.2f MB", vssMB)
	t.Logf("Target:                  < 100 MB (FR-039)")

	// Assert requirement: FR-039 - idle memory < 100MB
	const maxIdleMemoryMB = 100.0
	if rssMB > maxIdleMemoryMB {
		t.Errorf("FAIL: Idle memory usage %.2f MB exceeds requirement of %.0f MB (FR-039)", rssMB, maxIdleMemoryMB)
	} else {
		t.Logf("PASS: Idle memory usage meets FR-039 requirement")
	}

	// Check for memory leaks by monitoring over time
	t.Log("\nMonitoring for memory leaks over 5 minutes...")
	samples := []float64{rssMB}

	for i := 0; i < 5; i++ {
		time.Sleep(1 * time.Minute)

		memInfo, err := proc.MemoryInfo()
		if err != nil {
			t.Logf("WARNING: Failed to get memory info at minute %d: %v", i+1, err)
			continue
		}

		currentMB := float64(memInfo.RSS) / 1024 / 1024
		samples = append(samples, currentMB)
		t.Logf("Minute %d: %.2f MB", i+1, currentMB)
	}

	// Check for memory growth
	if len(samples) > 1 {
		growth := samples[len(samples)-1] - samples[0]
		growthPercent := (growth / samples[0]) * 100

		t.Logf("\nMemory Growth: %.2f MB (%.1f%%)", growth, growthPercent)

		if growthPercent > 10 {
			t.Errorf("WARNING: Memory grew by %.1f%% - possible memory leak", growthPercent)
		} else {
			t.Logf("PASS: Memory stable (growth within acceptable range)")
		}
	}
}

// TestMemoryUsageWith50Servers tests memory usage with 50 servers
// Requirement: FR-054 - System MUST efficiently handle up to 50 MCP servers
func TestMemoryUsageWith50Servers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping 50-server test in short mode")
	}

	// This test would require:
	// 1. Creating 50 mock server configurations
	// 2. Starting the application
	// 3. Waiting for discovery
	// 4. Measuring memory usage
	// 5. Verifying it's under reasonable limits (e.g., 300MB for 50 servers)

	t.Skip("TODO: Implement 50-server memory test (requires mock server setup)")

	// Expected implementation:
	// - Create temp config with 50 mock servers
	// - Start MCP Manager
	// - Wait for all servers to be discovered
	// - Measure RSS
	// - Assert RSS < 300MB (50 servers * ~4MB + 100MB base)
}

// BenchmarkMemoryFootprint benchmarks the memory footprint
func BenchmarkMemoryFootprint(b *testing.B) {
	exePath := filepath.Join("..", "..", "build", "bin", "mcpmanager.exe")
	if _, err := os.Stat(exePath); os.IsNotExist(err) {
		exePath = filepath.Join("..", "..", "build", "bin", "mcpmanager")
	}

	var totalMemory uint64

	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		cmd := exec.CommandContext(ctx, exePath)
		cmd.Env = append(os.Environ(), "HEADLESS=true")

		if err := cmd.Start(); err != nil {
			b.Fatalf("Failed to start: %v", err)
		}

		// Wait for stabilization
		time.Sleep(10 * time.Second)

		proc, err := process.NewProcess(int32(cmd.Process.Pid))
		if err == nil {
			if memInfo, err := proc.MemoryInfo(); err == nil {
				totalMemory += memInfo.RSS
			}
		}

		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		cancel()

		time.Sleep(2 * time.Second)
	}

	if b.N > 0 {
		avgMemoryMB := float64(totalMemory/uint64(b.N)) / 1024 / 1024
		b.ReportMetric(avgMemoryMB, "MB/idle")
	}
}

// getProcessMemoryStats returns memory statistics for the current process
func getProcessMemoryStats() (rss, vss uint64, err error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// For more accurate system-level memory, we'd use process package
	// This is a simplified version for the benchmark
	return m.Alloc, m.Sys, nil
}

// formatBytes formats bytes into human-readable format
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// Helper function to create mock server configurations
func createMockServerConfig(t *testing.T, count int, configPath string) {
	config := map[string]interface{}{
		"mcpServers": make(map[string]interface{}),
	}

	servers := config["mcpServers"].(map[string]interface{})

	for i := 0; i < count; i++ {
		serverName := fmt.Sprintf("mock-server-%d", i+1)
		servers[serverName] = map[string]interface{}{
			"command": "node",
			"args":    []string{"mock-server.js"},
			"env": map[string]string{
				"PORT": fmt.Sprintf("%d", 3000+i),
			},
		}
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
}
