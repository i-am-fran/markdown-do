package animation

import (
	"time"
)

// AnimationType represents the type of animation
type AnimationType string

const (
	AnimationFlash    AnimationType = "flash"
	AnimationCollapse AnimationType = "collapse"
)

// State represents the state of an animation
type State struct {
	Type      AnimationType
	StartTime time.Time
	Duration  time.Duration
	Active    bool
}

// NewFlashAnimation creates a new flash animation state
func NewFlashAnimation() *State {
	return &State{
		Type:      AnimationFlash,
		StartTime: time.Now(),
		Duration:  500 * time.Millisecond,
		Active:    true,
	}
}

// NewCollapseAnimation creates a new collapse animation state
func NewCollapseAnimation() *State {
	return &State{
		Type:      AnimationCollapse,
		StartTime: time.Now(),
		Duration:  800 * time.Millisecond,
		Active:    true,
	}
}

// Update updates the animation state
func (s *State) Update() bool {
	if !s.Active {
		return false
	}

	elapsed := time.Since(s.StartTime)
	if elapsed >= s.Duration {
		s.Active = false
		return false
	}

	return true
}

// Value returns the current animation value (0.0 to 1.0)
func (s *State) Value() float64 {
	if !s.Active {
		return 0.0
	}

	progress := float64(time.Since(s.StartTime)) / float64(s.Duration)

	switch s.Type {
	case AnimationFlash:
		// Flash: quick pulse up and down with easing
		if progress < 0.5 {
			// Ease out: start fast, slow down
			p := progress * 2
			return easeOutQuad(p)
		}
		// Ease in: start slow, speed up
		p := (progress - 0.5) * 2
		return 1.0 - easeInQuad(p)

	case AnimationCollapse:
		// Collapse: smooth decrease from 1 to 0 with spring-like easing
		return 1.0 - easeOutElastic(progress)
	}

	return 0.0
}

// IsActive returns whether the animation is still active
func (s *State) IsActive() bool {
	return s.Active
}

// easeOutQuad applies quadratic easing out (fast to slow)
func easeOutQuad(t float64) float64 {
	return t * (2 - t)
}

// easeInQuad applies quadratic easing in (slow to fast)
func easeInQuad(t float64) float64 {
	return t * t
}

// easeOutElastic applies elastic easing out (spring effect)
func easeOutElastic(t float64) float64 {
	if t == 0 || t == 1 {
		return t
	}
	p := 0.3
	s := p / 4
	return 1 - (0.1*(-10*t)+1)*(-10*t)*s
}
