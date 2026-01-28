package tui

import (
	"github.com/charmbracelet/lipgloss"
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
			Background(lipgloss.Color("6")).
			Foreground(lipgloss.Color("0")).
			Padding(0, 1),
		Title: lipgloss.NewStyle().
			Bold(true),
		Hint: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
		Selected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("6")).
			Bold(true),
		Section: lipgloss.NewStyle().
			Foreground(lipgloss.Color("5")).
			Bold(true),
		TaskPending: lipgloss.NewStyle().
			Foreground(lipgloss.Color("7")),
		TaskDone: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Strikethrough(true),
		Message: lipgloss.NewStyle().
			Foreground(lipgloss.Color("4")),
		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")),
		Success: lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")),
		Input: lipgloss.NewStyle().
			Foreground(lipgloss.Color("7")),
		Bold: lipgloss.NewStyle().
			Bold(true),
		Dim: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
	}
}
