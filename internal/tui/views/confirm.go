package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/i-am-fran/markdowndo/internal/tui/colors"
)

// ConfirmModel is a confirmation dialog model
type ConfirmModel struct {
	message string
	width   int
	height  int
}

// NewConfirmModel creates a new confirm model
func NewConfirmModel(message string, width, height int) ConfirmModel {
	return ConfirmModel{
		message: message,
		width:   width,
		height:  height,
	}
}

// Init implements tea.Model
func (m ConfirmModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m ConfirmModel) Update(msg tea.Msg) (ConfirmModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			return m, func() tea.Msg { return ConfirmMsg{Confirmed: true} }
		case "n", "N", "esc":
			return m, func() tea.Msg { return ConfirmMsg{Confirmed: false} }
		}
	}

	return m, nil
}

// View implements tea.Model
func (m ConfirmModel) View() string {
	s := lipgloss.NewStyle().Bold(true).Render(m.message) + "\n\n"
	s += lipgloss.NewStyle().Foreground(colors.Hint).Render("y yes  n no")
	return s
}

// ConfirmMsg is sent when the user confirms or cancels
type ConfirmMsg struct {
	Confirmed bool
}
