package undo

import (
	"time"
)

// Action represents an undoable action
type Action struct {
	Description string
	Snapshot    []string
	Timestamp   time.Time
	FilePath    string
}

// Stack manages undo actions
type Stack struct {
	actions []Action
	timeout time.Duration
}

// NewStack creates a new undo stack
func NewStack(timeout time.Duration) *Stack {
	return &Stack{
		actions: make([]Action, 0),
		timeout: timeout,
	}
}

// Push adds a new action to the stack
func (s *Stack) Push(action Action) {
	// Clean expired actions
	s.cleanExpired()

	// Add new action
	s.actions = append(s.actions, action)

	// Keep only last 10 actions
	if len(s.actions) > 10 {
		s.actions = s.actions[len(s.actions)-10:]
	}
}

// Pop removes and returns the most recent action
func (s *Stack) Pop() *Action {
	// Clean expired actions first
	s.cleanExpired()

	if len(s.actions) == 0 {
		return nil
	}

	action := s.actions[len(s.actions)-1]
	s.actions = s.actions[:len(s.actions)-1]
	return &action
}

// Peek returns the most recent action without removing it
func (s *Stack) Peek() *Action {
	// Clean expired actions first
	s.cleanExpired()

	if len(s.actions) == 0 {
		return nil
	}

	return &s.actions[len(s.actions)-1]
}

// Clear removes all actions
func (s *Stack) Clear() {
	s.actions = make([]Action, 0)
}

// IsEmpty returns true if the stack is empty
func (s *Stack) IsEmpty() bool {
	s.cleanExpired()
	return len(s.actions) == 0
}

// TimeRemaining returns the time remaining before the most recent action expires
func (s *Stack) TimeRemaining() time.Duration {
	action := s.Peek()
	if action == nil {
		return 0
	}

	elapsed := time.Since(action.Timestamp)
	remaining := s.timeout - elapsed

	if remaining < 0 {
		return 0
	}

	return remaining
}

// cleanExpired removes expired actions from the stack
func (s *Stack) cleanExpired() {
	now := time.Now()
	validActions := make([]Action, 0)

	for _, action := range s.actions {
		if now.Sub(action.Timestamp) < s.timeout {
			validActions = append(validActions, action)
		}
	}

	s.actions = validActions
}
