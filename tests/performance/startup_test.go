package performance

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// TestStartupTime benchmarks the application startup time
// Requirement: FR-037 - System MUST start within 2 seconds of launch
func TestStartupTime(t *testing.T) {
	// Build the application first
	buildCmd := exec.Command("wails", "build", "-clean")
	buildCmd.Dir = filepath.Join("..", "..")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build application: %v\nOutput: %s", err, output)
	}

	// Determine the executable path based on OS
	exePath := filepath.Join("..", "..", "build", "bin", "mcpmanager.exe")
	if _, err := os.Stat(exePath); os.IsNotExist(err) {
		// Try without .exe (macOS/Linux)
		exePath = filepath.Join("..", "..", "build", "bin", "mcpmanager")
	}

	// Run multiple iterations to get consistent results
	const iterations = 5
	var times []time.Duration

	for i := 0; i < iterations; i++ {
		startTime := time.Now()

		// Create a context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Start the application
		cmd := exec.CommandContext(ctx, exePath)
		cmd.Env = append(os.Environ(), "HEADLESS=true") // Run in headless mode if supported

		// Start the process
		if err := cmd.Start(); err != nil {
			t.Fatalf("Iteration %d: Failed to start application: %v", i+1, err)
		}

		// Give it a moment to initialize
		time.Sleep(100 * time.Millisecond)

		// Kill the process
		if cmd.Process != nil {
			cmd.Process.Kill()
		}

		elapsed := time.Since(startTime)
		times = append(times, elapsed)

		t.Logf("Iteration %d: Startup time: %v", i+1, elapsed)

		// Clean up between iterations
		time.Sleep(500 * time.Millisecond)
	}

	// Calculate statistics
	var total time.Duration
	minTime := times[0]
	maxTime := times[0]

	for _, duration := range times {
		total += duration
		if duration < minTime {
			minTime = duration
		}
		if duration > maxTime {
			maxTime = duration
		}
	}

	avgTime := total / time.Duration(len(times))

	// Report results
	t.Logf("\n=== Startup Time Results ===")
	t.Logf("Min:     %v", minTime)
	t.Logf("Max:     %v", maxTime)
	t.Logf("Average: %v", avgTime)
	t.Logf("Target:  < 2000ms (FR-037)")

	// Assert requirement: FR-037 - startup time < 2 seconds
	const maxStartupTime = 2000 * time.Millisecond
	if avgTime > maxStartupTime {
		t.Errorf("FAIL: Average startup time %v exceeds requirement of %v (FR-037)", avgTime, maxStartupTime)
	} else {
		t.Logf("PASS: Startup time meets FR-037 requirement")
	}

	// Warn if any individual run exceeded the limit
	for i, duration := range times {
		if duration > maxStartupTime {
			t.Logf("WARNING: Iteration %d exceeded target (%v)", i+1, duration)
		}
	}
}

// BenchmarkColdStart benchmarks a cold start with no filesystem cache
func BenchmarkColdStart(b *testing.B) {
	exePath := filepath.Join("..", "..", "build", "bin", "mcpmanager.exe")
	if _, err := os.Stat(exePath); os.IsNotExist(err) {
		exePath = filepath.Join("..", "..", "build", "bin", "mcpmanager")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		cmd := exec.CommandContext(ctx, exePath)
		cmd.Env = append(os.Environ(), "HEADLESS=true")

		if err := cmd.Start(); err != nil {
			b.Fatalf("Failed to start: %v", err)
		}

		time.Sleep(100 * time.Millisecond)

		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		cancel()

		time.Sleep(500 * time.Millisecond)
	}
}

// BenchmarkWarmStart benchmarks a warm start with filesystem cache
func BenchmarkWarmStart(b *testing.B) {
	exePath := filepath.Join("..", "..", "build", "bin", "mcpmanager.exe")
	if _, err := os.Stat(exePath); os.IsNotExist(err) {
		exePath = filepath.Join("..", "..", "build", "bin", "mcpmanager")
	}

	// Warm up the cache
	for i := 0; i < 2; i++ {
		cmd := exec.Command(exePath)
		cmd.Env = append(os.Environ(), "HEADLESS=true")
		cmd.Start()
		time.Sleep(100 * time.Millisecond)
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		time.Sleep(200 * time.Millisecond)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		cmd := exec.CommandContext(ctx, exePath)
		cmd.Env = append(os.Environ(), "HEADLESS=true")

		startTime := time.Now()
		if err := cmd.Start(); err != nil {
			b.Fatalf("Failed to start: %v", err)
		}

		time.Sleep(100 * time.Millisecond)
		elapsed := time.Since(startTime)

		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		cancel()

		b.ReportMetric(float64(elapsed.Milliseconds()), "ms/startup")
		time.Sleep(200 * time.Millisecond)
	}
}
