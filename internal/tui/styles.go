package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/i-am-fran/markdowndo/internal/tui/colors"
)

// Styles contains all the lipgloss styles for the TUI
type Styles struct {
	Header      lipgloss.Style
	Title       lipgloss.Style
	Hint        lipgloss.Style
	Selected    lipgloss.Style
	Section     lipgloss.Style
	TaskPending lipgloss.Style
	TaskDone    lipgloss.Style
	Message     lipgloss.Style
	Error       lipgloss.Style
	Success     lipgloss.Style
	Input       lipgloss.Style
	Bold        lipgloss.Style
	Dim         lipgloss.Style
}

// DefaultStyles returns the default styles
func DefaultStyles() Styles {
	return Styles{
		Header: lipgloss.NewStyle().
			Background(colors.HeaderBG).
			Foreground(colors.HeaderFG).
			Padding(0, 1),
		Title: lipgloss.NewStyle().
			Bold(true),
		Hint: lipgloss.NewStyle().
			Foreground(colors.Hint),
		Selected: lipgloss.NewStyle().
			Foreground(colors.Selected).
			Bold(true),
		Section: lipgloss.NewStyle().
			Foreground(colors.Section).
			Bold(true),
		TaskPending: lipgloss.NewStyle(),
		TaskDone: lipgloss.NewStyle().
			Foreground(colors.Hint).
			Strikethrough(true),
		Message: lipgloss.NewStyle().
			Foreground(colors.Info),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Success: lipgloss.NewStyle().
			Foreground(colors.Success),
		Input: lipgloss.NewStyle(),
		Bold: lipgloss.NewStyle().
			Bold(true),
		Dim: lipgloss.NewStyle().
			Foreground(colors.Hint),
	}
}
