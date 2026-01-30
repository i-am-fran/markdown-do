package views

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/i-am-fran/markdowndo/internal/tui/colors"
)

// ToastModel represents a toast notification with countdown
type ToastModel struct {
	message       string
	timeRemaining time.Duration
	width         int
	visible       bool
}

// NewToastModel creates a new toast model
func NewToastModel(width int) ToastModel {
	return ToastModel{
		width:   width,
		visible: false,
	}
}

// Show displays the toast with a message and duration
func (m *ToastModel) Show(message string, duration time.Duration) tea.Cmd {
	m.message = message
	m.timeRemaining = duration
	m.visible = true
	return m.tickCmd()
}

// Hide hides the toast
func (m *ToastModel) Hide() {
	m.visible = false
	m.timeRemaining = 0
}

// IsVisible returns whether the toast is currently visible
func (m *ToastModel) IsVisible() bool {
	return m.visible
}

// SetWidth updates the width
func (m *ToastModel) SetWidth(width int) {
	m.width = width
}

// Update implements tea.Model
func (m ToastModel) Update(msg tea.Msg) (ToastModel, tea.Cmd) {
	switch msg.(type) {
	case ToastTickMsg:
		if !m.visible {
			return m, nil
		}
		
		m.timeRemaining -= 100 * time.Millisecond
		
		if m.timeRemaining <= 0 {
			m.visible = false
			return m, nil
		}
		
		return m, m.tickCmd()
	}
	
	return m, nil
}

// View implements tea.Model
func (m ToastModel) View() string {
	if !m.visible {
		return ""
	}
	
	seconds := int(m.timeRemaining.Seconds())
	if seconds < 0 {
		seconds = 0
	}
	
	// Create progress bar
	totalWidth := 30
	filled := int(float64(m.timeRemaining) / float64(5*time.Second) * float64(totalWidth))
	if filled < 0 {
		filled = 0
	}
	if filled > totalWidth {
		filled = totalWidth
	}
	
	progressBar := ""
	for i := 0; i < totalWidth; i++ {
		if i < filled {
			progressBar += "█"
		} else {
			progressBar += "░"
		}
	}
	
	// Build toast content
	content := fmt.Sprintf("%s\n\nPress 'u' to undo (%ds) %s", m.message, seconds+1, progressBar)
	
	// Style the toast
	toast := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colors.Warning).
		Padding(1, 2).
		Foreground(colors.HeaderFG).
		Background(colors.HeaderBG).
		Width(m.width - 10).
		Render(content)
	
	return toast
}

// tickCmd returns a tick command for the countdown
func (m *ToastModel) tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return ToastTickMsg{Time: t}
	})
}

// ToastTickMsg is sent to update the toast countdown
type ToastTickMsg struct {
	Time time.Time
}
