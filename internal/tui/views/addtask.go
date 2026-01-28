package views

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AddTaskModel is the add task view model
type AddTaskModel struct {
	textInput   textinput.Model
	loop        bool
	lastAdded   string
	lastAddedID *int
	message     string
	placeholder string
	viewMode    string // "input" or "confirmOpen"
	width       int
	height      int
}

// NewAddTaskModel creates a new add task model
func NewAddTaskModel(loop bool, message, placeholder string, width, height int) AddTaskModel {
	if message == "" {
		message = "Enter task description (empty to finish):"
	}
	if placeholder == "" {
		placeholder = "Buy groceries or \"Fix bug @Features\" to add to section"
	}

	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = width - 10

	return AddTaskModel{
		textInput:   ti,
		loop:        loop,
		message:     message,
		placeholder: placeholder,
		viewMode:    "input",
		width:       width,
		height:      height,
	}
}

// Init implements tea.Model
func (m AddTaskModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model
func (m AddTaskModel) Update(msg tea.Msg) (AddTaskModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textInput.Width = msg.Width - 10
		return m, nil

	case tea.KeyMsg:
		if m.viewMode == "confirmOpen" {
			switch msg.String() {
			case "o", "O":
				// Open the task
				if m.lastAddedID != nil {
					return m, func() tea.Msg { return OpenTaskMsg{TaskID: *m.lastAddedID} }
				}
				m.viewMode = "input"
				m.lastAdded = ""
				m.lastAddedID = nil
				return m, nil
			case "enter", "n", "N", "c", "C":
				// Continue adding tasks or return to menu
				if !m.loop {
					return m, func() tea.Msg { return AddTaskCancelMsg{} }
				}
				m.viewMode = "input"
				m.lastAdded = ""
				m.lastAddedID = nil
				return m, nil
			case "esc":
				m.viewMode = "input"
				return m, func() tea.Msg { return AddTaskCancelMsg{} }
			}
			return m, nil
		}
		
		// Input mode
		switch msg.String() {
		case "enter":
			text := m.textInput.Value()
			if text == "" {
				return m, func() tea.Msg { return AddTaskCancelMsg{} }
			}
			return m, func() tea.Msg { return AddTaskSubmitMsg{Text: text} }

		case "esc":
			return m, func() tea.Msg { return AddTaskCancelMsg{} }
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// View implements tea.Model
func (m AddTaskModel) View() string {
	if m.viewMode == "confirmOpen" {
		var s string
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("✓ Added: "+m.lastAdded) + "\n\n"
		s += lipgloss.NewStyle().Bold(true).Render("Open this task?") + "\n\n"
		s += "Press " + lipgloss.NewStyle().Bold(true).Render("o") + " to open"
		if m.loop {
			s += " • " + lipgloss.NewStyle().Bold(true).Render("c") + " to continue adding"
		}
		s += " • " + lipgloss.NewStyle().Bold(true).Render("esc") + " to finish"
		return s
	}

	var s string

	if m.lastAdded != "" && m.viewMode == "input" {
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("✓ Added: "+m.lastAdded) + "\n\n"
	}

	s += lipgloss.NewStyle().Bold(true).Render(m.message) + "\n\n"
	s += lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("> ") + m.textInput.View() + "\n\n"

	hint := "enter submit • esc "
	if m.loop {
		hint += "finish"
	} else {
		hint += "cancel"
	}
	s += lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(hint)

	return s
}

// SetLastAdded sets the last added task text
func (m *AddTaskModel) SetLastAdded(text string) {
	m.lastAdded = text
	m.textInput.SetValue("")
}

// SetLastAddedWithID sets the last added task text and ID, and shows confirm prompt
func (m *AddTaskModel) SetLastAddedWithID(text string, id int) {
	m.lastAdded = text
	m.lastAddedID = &id
	m.viewMode = "confirmOpen"
	m.textInput.SetValue("")
}

// Reset resets the input
func (m *AddTaskModel) Reset() {
	m.textInput.SetValue("")
	m.lastAdded = ""
	m.lastAddedID = nil
	m.viewMode = "input"
}

// AddTaskSubmitMsg is sent when a task is submitted
type AddTaskSubmitMsg struct {
	Text string
}

// AddTaskCancelMsg is sent when adding is cancelled
type AddTaskCancelMsg struct{}

// OpenTaskMsg is sent when user wants to open a task
type OpenTaskMsg struct {
	TaskID int
}
