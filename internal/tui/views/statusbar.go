package views

import (
	"fmt"
	"path/filepath"

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
		if task.Status == core.TaskPending {
			pendingTasks++
		}
	}

	// Format file path (show just filename or relative path if short enough)
	displayPath := filepath.Base(m.filePath)
	if len(m.filePath) < 40 {
		displayPath = m.filePath
	}

	// Left side: task count
	leftText := fmt.Sprintf("%d/%d tasks", pendingTasks, totalTasks)
	
	// Right side: file path
	rightText := displayPath

	// Create status bar with padding
	leftPart := lipgloss.NewStyle().
		Foreground(colors.HeaderFG).
		Background(colors.HeaderBG).
		Padding(0, 1).
		Render(leftText)

	rightPart := lipgloss.NewStyle().
		Foreground(colors.HeaderFG).
		Background(colors.HeaderBG).
		Padding(0, 1).
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
