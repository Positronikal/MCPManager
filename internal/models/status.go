package models

import (
	"fmt"
	"time"
)

// ServerStatus represents the runtime status of an MCP server
type ServerStatus struct {
	State            StatusState `json:"state"`
	StartupAttempts  int         `json:"startupAttempts"`
	LastStateChange  time.Time   `json:"lastStateChange"`
	ErrorMessage     string      `json:"errorMessage,omitempty"`
	CrashRecoverable bool        `json:"crashRecoverable"`
}

// NewServerStatus creates a new ServerStatus in the stopped state
func NewServerStatus() *ServerStatus {
	return &ServerStatus{
		State:            StatusStopped,
		StartupAttempts:  0,
		LastStateChange:  time.Now(),
		CrashRecoverable: true,
	}
}

// State transition rules based on the state machine diagram:
// stopped -> starting (user initiates start)
// starting -> running (startup successful)
// starting -> error (startup failed)
// starting -> stopped (user cancels startup)
// running -> stopped (user initiates stop)
// running -> error (crash/unexpected termination)
// error -> starting (retry attempt)
// error -> stopped (user cancels retry)

var validTransitions = map[StatusState][]StatusState{
	StatusStopped:  {StatusStarting},
	StatusStarting: {StatusRunning, StatusError, StatusStopped}, // Allow canceling startup
	StatusRunning:  {StatusStopped, StatusError},
	StatusError:    {StatusStarting, StatusStopped},
}

// CanTransitionTo checks if transitioning to the new state is valid
func (s *ServerStatus) CanTransitionTo(newState StatusState) bool {
	if !newState.IsValid() {
		return false
	}

	allowedStates, exists := validTransitions[s.State]
	if !exists {
		return false
	}

	for _, allowed := range allowedStates {
		if allowed == newState {
			return true
		}
	}

	return false
}

// TransitionTo attempts to transition to a new state
func (s *ServerStatus) TransitionTo(newState StatusState, reason string) error {
	if !s.CanTransitionTo(newState) {
		return fmt.Errorf("invalid state transition from %s to %s", s.State, newState)
	}

	oldState := s.State
	s.State = newState
	s.LastStateChange = time.Now()

	// Handle startup attempts logic
	if newState == StatusStarting {
		// Increment startup attempts when entering starting state
		s.StartupAttempts++
	} else if newState == StatusRunning && oldState == StatusStarting {
		// Reset startup attempts on successful start
		s.StartupAttempts = 0
		s.ErrorMessage = ""
	} else if newState == StatusError {
		// Set error message
		s.ErrorMessage = reason
		// Determine if crash is recoverable
		s.CrashRecoverable = s.StartupAttempts < 3
	} else if newState == StatusStopped {
		// Reset error state when stopped
		s.ErrorMessage = ""
	}

	return nil
}

// Validate checks if the ServerStatus is in a valid state
func (s *ServerStatus) Validate() error {
	if !s.State.IsValid() {
		return fmt.Errorf("invalid status state: %s", s.State)
	}

	if s.StartupAttempts < 0 {
		return fmt.Errorf("startupAttempts cannot be negative: %d", s.StartupAttempts)
	}

	if s.State == StatusError && s.ErrorMessage == "" {
		return fmt.Errorf("error state requires an error message")
	}

	return nil
}
