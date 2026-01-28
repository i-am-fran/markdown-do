package views

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TextPromptModel is a generic text input prompt
type TextPromptModel struct {
	textInput   textinput.Model
	message     string
	placeholder string
	width       int
	height      int
}

// NewTextPromptModel creates a new text prompt model
func NewTextPromptModel(message, initialValue, placeholder string, width, height int) TextPromptModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.SetValue(initialValue)
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = width - 10

	return TextPromptModel{
		textInput:   ti,
		message:     message,
		placeholder: placeholder,
		width:       width,
		height:      height,
	}
}

// Init implements tea.Model
func (m TextPromptModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model
func (m TextPromptModel) Update(msg tea.Msg) (TextPromptModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textInput.Width = msg.Width - 10
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			text := m.textInput.Value()
			if text == "" {
				return m, func() tea.Msg { return TextPromptCancelMsg{} }
			}
			return m, func() tea.Msg { return TextPromptSubmitMsg{Text: text} }

		case "esc":
			return m, func() tea.Msg { return TextPromptCancelMsg{} }
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// View implements tea.Model
func (m TextPromptModel) View() string {
	var s string

	s += lipgloss.NewStyle().Bold(true).Render(m.message) + "\n\n"
	s += lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#5C5C5C"}).Render("> ") + m.textInput.View() + "\n\n"
	s += lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#5C5C5C"}).Render("enter submit • esc cancel")

	return s
}

// TextPromptSubmitMsg is sent when text is submitted
type TextPromptSubmitMsg struct {
	Text string
}

// TextPromptCancelMsg is sent when prompt is cancelled
type TextPromptCancelMsg struct{}
