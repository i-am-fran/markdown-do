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
			Background(lipgloss.AdaptiveColor{Light: "#7D56F4", Dark: "#7D56F4"}).
			Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"}).
			Padding(0, 1),
		Title: lipgloss.NewStyle().
			Bold(true),
		Hint: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#5C5C5C"}),
		Selected: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#7D56F4", Dark: "#7D56F4"}).
			Bold(true),
		Section: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#FF00FF", Dark: "#FF00FF"}).
			Bold(true),
		TaskPending: lipgloss.NewStyle(),
		TaskDone: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#5C5C5C"}).
			Strikethrough(true),
		Message: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#0000FF", Dark: "#00BFFF"}),
		Error: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#FF0000", Dark: "#FF0000"}),
		Success: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#00A000", Dark: "#00FF00"}),
		Input: lipgloss.NewStyle(),
		Bold: lipgloss.NewStyle().
			Bold(true),
		Dim: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#5C5C5C"}),
	}
}
