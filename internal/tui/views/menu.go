package views

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/i-am-fran/markdowndo/internal/core"
	"github.com/i-am-fran/markdowndo/internal/tui/colors"
)

// MenuAction represents a main menu action
type MenuAction string

const (
	ActionList            MenuAction = "list"
	ActionAdd             MenuAction = "add"
	ActionAddNote         MenuAction = "addNote"
	ActionFind            MenuAction = "find"
	ActionDeleteCompleted MenuAction = "deleteCompleted"
	ActionLint            MenuAction = "lint"
	ActionOpen            MenuAction = "open"
	ActionSubfolders      MenuAction = "subfolders"
	ActionSettings        MenuAction = "settings"
)

type menuItem struct {
	title  string
	action MenuAction
}

func (i menuItem) Title() string       { return i.title }
func (i menuItem) Description() string { return "" }
func (i menuItem) FilterValue() string { return i.title }

// MenuModel is the main menu view model
type MenuModel struct {
	list        list.Model
	todoFile    *core.TodoFile
	width       int
	height      int
	lastEscTime int64
	escHint     bool
}

// NewMenuModel creates a new menu model
func NewMenuModel(todoFile *core.TodoFile, width, height int) MenuModel {
	taskSummary := " (no tasks)"
	if todoFile != nil {
		tasks := todoFile.GetTasks()
		total := len(tasks)
		pending := 0
		for _, t := range tasks {
			if t.Status == core.TaskPending {
				pending++
			}
		}
		if total > 0 {
			taskSummary = fmt.Sprintf(" (%d pending, %d total)", pending, total)
		}
	}

	items := []list.Item{
		menuItem{title: "List tasks" + taskSummary, action: ActionList},
		menuItem{title: "Add new task", action: ActionAdd},
		menuItem{title: "Find tasks", action: ActionFind},
		menuItem{title: "Delete completed tasks", action: ActionDeleteCompleted},
		menuItem{title: "Lint and fix formatting", action: ActionLint},
		menuItem{title: "Add note", action: ActionAddNote},
		menuItem{title: "Open in editor", action: ActionOpen},
		menuItem{title: "List subfolders [WIP]", action: ActionSubfolders},
		menuItem{title: "Settings", action: ActionSettings},
	}

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.SetSpacing(0)

	l := list.New(items, delegate, width, height-6)
	l.Title = "What would you like to do?"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.Styles.Title = lipgloss.NewStyle().Bold(true).MarginBottom(1)

	return MenuModel{
		list:     l,
		todoFile: todoFile,
		width:    width,
		height:   height,
	}
}

// Init implements tea.Model
func (m MenuModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m MenuModel) Update(msg tea.Msg) (MenuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-6)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if item, ok := m.list.SelectedItem().(menuItem); ok {
				return m, func() tea.Msg {
					return MenuSelectMsg{Action: item.action}
				}
			}
		case "esc":
			return m, func() tea.Msg {
				return EscPressedMsg{}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View implements tea.Model
func (m MenuModel) View() string {
	hint := "↑↓ navigate • enter select • esc esc quit"
	return m.list.View() + "\n\n" + lipgloss.NewStyle().Foreground(colors.Hint).Render(hint)
}

// SetEscHint sets the escape hint visibility
func (m *MenuModel) SetEscHint(show bool) {
	m.escHint = show
}

// MenuSelectMsg is sent when a menu item is selected
type MenuSelectMsg struct {
	Action MenuAction
}

// EscPressedMsg is sent when escape is pressed
type EscPressedMsg struct{}
