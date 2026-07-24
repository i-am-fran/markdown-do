package views

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/i-am-fran/markdowndo/internal/core"
	"github.com/i-am-fran/markdowndo/internal/tui/colors"
)

// StatusBarModel represents the status bar at the bottom
type StatusBarModel struct {
	todoFile *core.TodoFile
	filePath string
	width    int
}

// NewStatusBarModel creates a new status bar model
func NewStatusBarModel(todoFile *core.TodoFile, filePath string, width int) StatusBarModel {
	return StatusBarModel{
		todoFile: todoFile,
		filePath: filePath,
		width:    width,
	}
}

// Update updates the status bar with new data
func (m *StatusBarModel) Update(todoFile *core.TodoFile, filePath string) {
	m.todoFile = todoFile
	m.filePath = filePath
}

// SetWidth updates the width of the status bar
func (m *StatusBarModel) SetWidth(width int) {
	m.width = width
}

// View renders the status bar
func (m StatusBarModel) View() string {
	if m.todoFile == nil {
		return ""
	}

	// Count tasks
	totalTasks := len(m.todoFile.GetTasks())
	pendingTasks := 0
	for _, task := range m.todoFile.GetTasks() {
		if task.Status != core.TaskCompleted {
			pendingTasks++
		}
	}

	// Format file path (show folder + filename for better context)
	displayPath := m.filePath
	if len(m.filePath) >= 40 {
		// Show parent folder + filename instead of just filename
		dir := filepath.Dir(m.filePath)
		filename := filepath.Base(m.filePath)
		parentDir := filepath.Base(dir)
		displayPath = filepath.Join(parentDir, filename)
	}

	// Calculate completion percentage for visual indicator
	completionPercent := 0.0
	if totalTasks > 0 {
		completionPercent = float64(totalTasks-pendingTasks) / float64(totalTasks) * 100
	}

	// Visual progress indicator (small bar)
	progressWidth := 10
	filled := int(completionPercent / 100 * float64(progressWidth))
	if filled > progressWidth {
		filled = progressWidth
	}
	progressBar := strings.Repeat("█", filled) + strings.Repeat("░", progressWidth-filled)

	// Left side: task count with progress
	leftText := fmt.Sprintf("%d/%d [%s] %.0f%%", pendingTasks, totalTasks, progressBar, completionPercent)

	// Right side: file path
	rightText := displayPath

	// Create modern status bar with better styling
	leftPart := lipgloss.NewStyle().
		Foreground(colors.HeaderFG).
		Background(colors.HeaderBG).
		Padding(0, 2).
		Bold(true).
		Render(leftText)

	rightPart := lipgloss.NewStyle().
		Foreground(colors.HeaderFG).
		Background(colors.HeaderBG).
		Padding(0, 2).
		Render(rightText)

	// Calculate space between left and right
	leftWidth := lipgloss.Width(leftPart)
	rightWidth := lipgloss.Width(rightPart)
	spacerWidth := m.width - leftWidth - rightWidth
	if spacerWidth < 0 {
		spacerWidth = 0
	}

	spacer := lipgloss.NewStyle().
		Background(colors.HeaderBG).
		Width(spacerWidth).
		Render("")

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPart, spacer, rightPart)
}
