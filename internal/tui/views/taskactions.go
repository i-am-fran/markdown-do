package views

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/i-am-fran/markdowndo/internal/core"
	"github.com/i-am-fran/markdowndo/internal/tui/colors"
	"github.com/i-am-fran/markdowndo/internal/tui/markdown"
)

// TaskAction represents a task action
type TaskAction string

const (
	TaskActionToggle TaskAction = "toggle"
	TaskActionEdit   TaskAction = "edit"
	TaskActionMove   TaskAction = "move"
	TaskActionDelete TaskAction = "delete"
	TaskActionBack   TaskAction = "back"
)

type taskActionItem struct {
	title  string
	action TaskAction
}

func (i taskActionItem) Title() string       { return i.title }
func (i taskActionItem) Description() string { return "" }
func (i taskActionItem) FilterValue() string { return i.title }

// TaskActionsModel is the task actions view model
type TaskActionsModel struct {
	list   list.Model
	task   *core.Task
	width  int
	height int
}

// NewTaskActionsModel creates a new task actions model
func NewTaskActionsModel(task *core.Task, width, height int) TaskActionsModel {
	toggleLabel := "Mark as complete"
	if task.Status == core.TaskCompleted {
		toggleLabel = "Reopen task"
	}

	items := []list.Item{
		taskActionItem{title: toggleLabel, action: TaskActionToggle},
		taskActionItem{title: "Edit text", action: TaskActionEdit},
		taskActionItem{title: "Move to section", action: TaskActionMove},
		taskActionItem{title: "Delete", action: TaskActionDelete},
		taskActionItem{title: "Back", action: TaskActionBack},
	}

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.SetSpacing(0)

	l := list.New(items, delegate, width, height-6)
	l.Title = fmt.Sprintf("Task: %s", task.Text)
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

	return TaskActionsModel{
		list:   l,
		task:   task,
		width:  width,
		height: height,
	}
}

// Init implements tea.Model
func (m TaskActionsModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m TaskActionsModel) Update(msg tea.Msg) (TaskActionsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-6)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if item, ok := m.list.SelectedItem().(taskActionItem); ok && m.task != nil {
				return m, func() tea.Msg {
					return TaskActionSelectMsg{Action: item.action, TaskID: m.task.ID}
				}
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
func (m TaskActionsModel) View() string {
	if m.task == nil {
		return "No task selected"
	}

	// Render task text with markdown styling in a card
	taskPreview := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colors.Border).
		Padding(1, 2).
		Width(m.width - 4).
		MarginBottom(1).
		Render(markdown.Render(m.task.Text))

	hint := lipgloss.NewStyle().
		Foreground(colors.Hint).
		Italic(true).
		Render("enter select  esc back")
	return taskPreview + "\n" + m.list.View() + "\n\n" + hint
}

// TaskActionSelectMsg is sent when a task action is selected
type TaskActionSelectMsg struct {
	Action TaskAction
	TaskID int
}
