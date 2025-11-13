package models

import (
	"testing"
	"time"
)

func TestNewServerStatus(t *testing.T) {
	status := NewServerStatus()

	if status.State != StatusStopped {
		t.Errorf("Expected initial state to be stopped, got %s", status.State)
	}
	if status.StartupAttempts != 0 {
		t.Errorf("Expected startup attempts to be 0, got %d", status.StartupAttempts)
	}
	if status.CrashRecoverable != true {
		t.Error("Expected crash recoverable to be true initially")
	}
}

func TestServerStatusStateTransitions(t *testing.T) {
	tests := []struct {
		name      string
		fromState StatusState
		toState   StatusState
		valid     bool
	}{
		// Valid transitions
		{"stopped to starting", StatusStopped, StatusStarting, true},
		{"starting to running", StatusStarting, StatusRunning, true},
		{"starting to error", StatusStarting, StatusError, true},
		{"starting to stopped", StatusStarting, StatusStopped, true}, // Allow canceling startup
		{"running to stopped", StatusRunning, StatusStopped, true},
		{"running to error (crash)", StatusRunning, StatusError, true},
		{"error to starting (retry)", StatusError, StatusStarting, true},
		{"error to stopped", StatusError, StatusStopped, true},

		// Invalid transitions
		{"stopped to running", StatusStopped, StatusRunning, false},
		{"stopped to error", StatusStopped, StatusError, false},
		{"error to running", StatusError, StatusRunning, false},
		{"running to starting", StatusRunning, StatusStarting, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := &ServerStatus{
				State:           tt.fromState,
				StartupAttempts: 0,
				LastStateChange: time.Now(),
			}

			canTransition := status.CanTransitionTo(tt.toState)
			if canTransition != tt.valid {
				t.Errorf("CanTransitionTo(%s -> %s) = %v, want %v",
					tt.fromState, tt.toState, canTransition, tt.valid)
			}

			// Test TransitionTo
			err := status.TransitionTo(tt.toState, "test reason")
			if tt.valid {
				if err != nil {
					t.Errorf("TransitionTo(%s -> %s) failed: %v", tt.fromState, tt.toState, err)
				}
				if status.State != tt.toState {
					t.Errorf("State not updated: got %s, want %s", status.State, tt.toState)
				}
			} else {
				if err == nil {
					t.Errorf("TransitionTo(%s -> %s) should have failed", tt.fromState, tt.toState)
				}
			}
		})
	}
}

func TestStartupAttemptsLogic(t *testing.T) {
	status := NewServerStatus()

	// First startup attempt
	if err := status.TransitionTo(StatusStarting, "initial start"); err != nil {
		t.Fatal(err)
	}
	if status.StartupAttempts != 1 {
		t.Errorf("Expected startup attempts to be 1, got %d", status.StartupAttempts)
	}

	// Successful start should reset attempts
	if err := status.TransitionTo(StatusRunning, "started successfully"); err != nil {
		t.Fatal(err)
	}
	if status.StartupAttempts != 0 {
		t.Errorf("Expected startup attempts to be reset to 0, got %d", status.StartupAttempts)
	}
	if status.ErrorMessage != "" {
		t.Error("Error message should be cleared on successful start")
	}

	// Stop the server
	if err := status.TransitionTo(StatusStopped, "user stopped"); err != nil {
		t.Fatal(err)
	}

	// Start again
	if err := status.TransitionTo(StatusStarting, "restart"); err != nil {
		t.Fatal(err)
	}
	if status.StartupAttempts != 1 {
		t.Errorf("Expected startup attempts to be 1 after restart, got %d", status.StartupAttempts)
	}

	// Fail to start
	if err := status.TransitionTo(StatusError, "startup failed"); err != nil {
		t.Fatal(err)
	}
	if status.ErrorMessage != "startup failed" {
		t.Errorf("Expected error message to be set, got %s", status.ErrorMessage)
	}

	// Retry
	if err := status.TransitionTo(StatusStarting, "retry"); err != nil {
		t.Fatal(err)
	}
	if status.StartupAttempts != 2 {
		t.Errorf("Expected startup attempts to be 2, got %d", status.StartupAttempts)
	}

	// Fail again
	if err := status.TransitionTo(StatusError, "startup failed again"); err != nil {
		t.Fatal(err)
	}

	// Retry again
	if err := status.TransitionTo(StatusStarting, "retry 2"); err != nil {
		t.Fatal(err)
	}
	if status.StartupAttempts != 3 {
		t.Errorf("Expected startup attempts to be 3, got %d", status.StartupAttempts)
	}

	// After 3 attempts, should still be recoverable but approaching limit
	if err := status.TransitionTo(StatusError, "failed after 3 attempts"); err != nil {
		t.Fatal(err)
	}
	if status.CrashRecoverable != false {
		t.Error("Expected crash to be unrecoverable after 3 attempts")
	}
}

func TestServerStatusValidation(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *ServerStatus
		wantErr bool
	}{
		{
			name: "valid stopped status",
			setup: func() *ServerStatus {
				return NewServerStatus()
			},
			wantErr: false,
		},
		{
			name: "invalid status state",
			setup: func() *ServerStatus {
				status := NewServerStatus()
				status.State = StatusState("invalid")
				return status
			},
			wantErr: true,
		},
		{
			name: "negative startup attempts",
			setup: func() *ServerStatus {
				status := NewServerStatus()
				status.StartupAttempts = -1
				return status
			},
			wantErr: true,
		},
		{
			name: "error state without message",
			setup: func() *ServerStatus {
				status := NewServerStatus()
				status.State = StatusError
				status.ErrorMessage = ""
				return status
			},
			wantErr: true,
		},
		{
			name: "error state with message",
			setup: func() *ServerStatus {
				status := NewServerStatus()
				status.State = StatusError
				status.ErrorMessage = "test error"
				return status
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := tt.setup()
			err := status.Validate()

			if tt.wantErr && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestLastStateChangeUpdated(t *testing.T) {
	status := NewServerStatus()
	initialTime := status.LastStateChange

	// Wait a bit to ensure time difference
	time.Sleep(10 * time.Millisecond)

	// Transition to a new state
	if err := status.TransitionTo(StatusStarting, "test"); err != nil {
		t.Fatal(err)
	}

	if !status.LastStateChange.After(initialTime) {
		t.Error("LastStateChange should be updated on state transition")
	}
}

func TestErrorStateBehavior(t *testing.T) {
	status := NewServerStatus()

	// Transition to starting then error
	if err := status.TransitionTo(StatusStarting, "start"); err != nil {
		t.Fatal(err)
	}
	if err := status.TransitionTo(StatusError, "connection refused"); err != nil {
		t.Fatal(err)
	}

	if status.ErrorMessage != "connection refused" {
		t.Errorf("Error message not set correctly: got %s", status.ErrorMessage)
	}

	// Transition to stopped should clear error
	if err := status.TransitionTo(StatusStopped, "user stopped"); err != nil {
		t.Fatal(err)
	}

	if status.ErrorMessage != "" {
		t.Error("Error message should be cleared when transitioning to stopped")
	}
}
