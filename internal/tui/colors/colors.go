package colors

import "github.com/charmbracelet/lipgloss"

// Adaptive color definitions for light/dark terminal themes
var (
	Hint     = lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#5C5C5C"}
	Success  = lipgloss.AdaptiveColor{Light: "#00A000", Dark: "#00FF00"}
	Error    = lipgloss.AdaptiveColor{Light: "#FF0000", Dark: "#FF0000"}
	Info     = lipgloss.AdaptiveColor{Light: "#00008B", Dark: "#00BFFF"}
	Warning  = lipgloss.AdaptiveColor{Light: "#B8860B", Dark: "#FFD700"}
	HeaderBG = lipgloss.AdaptiveColor{Light: "#7D56F4", Dark: "#7D56F4"}
	HeaderFG = lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"}
	Selected = lipgloss.AdaptiveColor{Light: "#7D56F4", Dark: "#7D56F4"}
	Section  = lipgloss.AdaptiveColor{Light: "#9B30FF", Dark: "#FF00FF"}
)
