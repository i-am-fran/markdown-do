package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/i-am-fran/markdowndo/internal/tui/colors"
)

// Styles contains all the lipgloss styles for the TUI
type Styles struct {
	Header      lipgloss.Style
	Title       lipgloss.Style
	Subtitle    lipgloss.Style
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
	Border      lipgloss.Style
	Card        lipgloss.Style
}

// DefaultStyles returns the default styles
func DefaultStyles() Styles {
	return Styles{
		Header: lipgloss.NewStyle().
			Background(colors.HeaderBG).
			Foreground(colors.HeaderFG).
			Padding(0, 2).
			Bold(true),
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(colors.Accent).
			MarginBottom(1),
		Subtitle: lipgloss.NewStyle().
			Foreground(colors.Muted).
			MarginBottom(1),
		Hint: lipgloss.NewStyle().
			Foreground(colors.Hint).
			Italic(true),
		Selected: lipgloss.NewStyle().
			Foreground(colors.Selected).
			Bold(true),
		Section: lipgloss.NewStyle().
			Foreground(colors.Section).
			Bold(true).
			MarginTop(1).
			MarginBottom(1),
		TaskPending: lipgloss.NewStyle().
			Foreground(lipgloss.NoColor{}),
		TaskDone: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Strikethrough(true),
		Message: lipgloss.NewStyle().
			Foreground(colors.Info),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Success: lipgloss.NewStyle().
			Foreground(colors.Success),
		Input: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colors.Border).
			Padding(0, 1),
		Bold: lipgloss.NewStyle().
			Bold(true),
		Dim: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Border: lipgloss.NewStyle().
			Foreground(colors.Border),
		Card: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colors.Border).
			Padding(1, 2).
			Margin(1, 0),
	}
}
