package views

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/i-am-fran/markdowndo/internal/core"
	"github.com/i-am-fran/markdowndo/internal/tui/colors"
)

type moveSectionItem struct {
	title   string
	section *string // nil means inbox
	isNew   bool
	isBack  bool
}

func (i moveSectionItem) Title() string       { return i.title }
func (i moveSectionItem) Description() string { return "" }
func (i moveSectionItem) FilterValue() string { return i.title }

// MoveTaskModel is the move task view model
type MoveTaskModel struct {
	list      list.Model
	textInput textinput.Model
	todoFile  *core.TodoFile
	task      *core.Task
	viewMode  string // "select" or "new"
	width     int
	height    int
}

// NewMoveTaskModel creates a new move task model
func NewMoveTaskModel(todoFile *core.TodoFile, task *core.Task, width, height int) MoveTaskModel {
	sections := todoFile.GetSectionNames()

	var items []list.Item

	// Add inbox option if task is in a section
	if task.Section != nil {
		items = append(items, moveSectionItem{title: "Inbox (no section)", section: nil})
	}

	// Add sections (except current one)
	for _, section := range sections {
		if task.Section == nil || section != *task.Section {
			s := section
			items = append(items, moveSectionItem{title: section, section: &s})
		}
	}

	items = append(items, moveSectionItem{title: "+ Create new section", isNew: true})
	items = append(items, moveSectionItem{title: "Back", isBack: true})

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.SetSpacing(0)

	l := list.New(items, delegate, width, height-6)
	l.Title = "Move to section:"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(colors.Accent).
		Bold(true).
		MarginBottom(1).
		MarginTop(1)
	l.Styles.NoItems = lipgloss.NewStyle().
		Foreground(colors.Muted).
		Italic(true)

	ti := textinput.New()
	ti.Placeholder = "Features, Bugs, etc."
	ti.CharLimit = 64
	ti.Width = width - 10

	return MoveTaskModel{
		list:      l,
		textInput: ti,
		todoFile:  todoFile,
		task:      task,
		viewMode:  "select",
		width:     width,
		height:    height,
	}
}

// Init implements tea.Model
func (m MoveTaskModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m MoveTaskModel) Update(msg tea.Msg) (MoveTaskModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-6)
		m.textInput.Width = msg.Width - 10
		return m, nil

	case tea.KeyMsg:
		if m.viewMode == "new" {
			switch msg.String() {
			case "enter":
				text := m.textInput.Value()
				if text == "" {
					m.viewMode = "select"
					return m, nil
				}
				return m, func() tea.Msg { return MoveTaskSelectMsg{Section: &text, TaskID: m.task.ID} }

			case "esc":
				m.viewMode = "select"
				m.textInput.SetValue("")
				return m, nil
			}

			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "enter":
			if item, ok := m.list.SelectedItem().(moveSectionItem); ok {
				if item.isBack {
					return m, func() tea.Msg { return BackMsg{} }
				}
				if item.isNew {
					m.viewMode = "new"
					m.textInput.Focus()
					return m, textinput.Blink
				}
				return m, func() tea.Msg { return MoveTaskSelectMsg{Section: item.section, TaskID: m.task.ID} }
			}
		case "esc":
			return m, func() tea.Msg { return BackMsg{} }
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View implements tea.Model
func (m MoveTaskModel) View() string {
	if m.viewMode == "new" {
		titleStyle := lipgloss.NewStyle().
			Foreground(colors.Accent).
			Bold(true).
			MarginBottom(1)
		
		inputStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colors.Border).
			Padding(0, 1).
			Width(m.width - 6)
		
		hintStyle := lipgloss.NewStyle().
			Foreground(colors.Hint).
			Italic(true)
		
		s := titleStyle.Render("New section name:") + "\n"
		s += inputStyle.Render(m.textInput.View()) + "\n\n"
		s += hintStyle.Render("enter create  esc back")
		return s
	}

	hintStyle := lipgloss.NewStyle().
		Foreground(colors.Hint).
		Italic(true)
	hint := hintStyle.Render("enter select  esc back")
	return m.list.View() + "\n\n" + hint
}

// MoveTaskSelectMsg is sent when a section is selected
type MoveTaskSelectMsg struct {
	Section *string
	TaskID  int
}
