package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/i-am-fran/markdowndo/internal/config"
	"github.com/i-am-fran/markdowndo/internal/core"
	"github.com/i-am-fran/markdowndo/internal/tui/colors"
)

type taskListItem struct {
	task      *core.Task
	section   string
	isSection bool
	isAdd     bool
}

func (i taskListItem) Title() string {
	if i.isSection {
		return fmt.Sprintf("── %s ──", i.section)
	}
	if i.isAdd {
		return "+ Add new task"
	}

	checkbox := "[ ]"
	if i.task.Status == core.TaskCompleted {
		checkbox = "[x]"
	}
	return fmt.Sprintf("%d. %s %s", i.task.ID, checkbox, i.task.Text)
}

func (i taskListItem) Description() string { return "" }
func (i taskListItem) FilterValue() string { return i.Title() }

// TaskListModel is the task list view model
type TaskListModel struct {
	list     list.Model
	todoFile *core.TodoFile
	settings config.Settings
	width    int
	height   int
}

// NewTaskListModel creates a new task list model
func NewTaskListModel(todoFile *core.TodoFile, width, height int) TaskListModel {
	settings := config.GetSettings()

	items := buildTaskListItems(todoFile, settings)

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.SetSpacing(0)
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Foreground(colors.Selected)
	// Use default color for normal titles (no explicit foreground)

	l := list.New(items, delegate, width, height-6)
	l.Title = "Select a task:"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.Styles.Title = lipgloss.NewStyle().Bold(true).MarginBottom(1)

	return TaskListModel{
		list:     l,
		todoFile: todoFile,
		settings: settings,
		width:    width,
		height:   height,
	}
}

func buildTaskListItems(todoFile *core.TodoFile, settings config.Settings) []list.Item {
	var items []list.Item
	groups := todoFile.GetTasksGroupedBySectionOrdered()

	for _, group := range groups {
		var visibleTasks []core.Task
		for _, task := range group.Tasks {
			if settings.ShowCompleted || task.Status != core.TaskCompleted {
				visibleTasks = append(visibleTasks, task)
			}
		}

		if len(visibleTasks) == 0 {
			continue
		}

		if group.Section != nil {
			items = append(items, taskListItem{
				isSection: true,
				section:   *group.Section,
			})
		}

		for i := range visibleTasks {
			items = append(items, taskListItem{task: &visibleTasks[i]})
		}
	}

	items = append(items, taskListItem{isAdd: true})

	return items
}

// Init implements tea.Model
func (m TaskListModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m TaskListModel) Update(msg tea.Msg) (TaskListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-6)
		return m, nil

	case tea.KeyMsg:
		// Get current item for hotkey actions
		item, ok := m.list.SelectedItem().(taskListItem)

		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if ok {
				if item.isAdd {
					return m, func() tea.Msg { return AddTaskMsg{} }
				}
				if item.task != nil {
					return m, func() tea.Msg { return TaskSelectedMsg{TaskID: item.task.ID} }
				}
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			return m, func() tea.Msg { return BackMsg{} }

		case key.Matches(msg, key.NewBinding(key.WithKeys("c"))):
			if ok && item.task != nil {
				return m, func() tea.Msg { return ToggleTaskMsg{TaskID: item.task.ID} }
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("d"))):
			if ok && item.task != nil {
				return m, func() tea.Msg { return DeleteTaskMsg{TaskID: item.task.ID} }
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("e"))):
			if ok && item.task != nil {
				return m, func() tea.Msg { return EditTaskMsg{TaskID: item.task.ID} }
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("m"))):
			if ok && item.task != nil {
				return m, func() tea.Msg { return MoveTaskMsg{TaskID: item.task.ID} }
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View implements tea.Model
func (m TaskListModel) View() string {
	groups := m.todoFile.GetTasksGroupedBySectionOrdered()
	hasVisibleTasks := false
	for _, group := range groups {
		for _, task := range group.Tasks {
			if m.settings.ShowCompleted || task.Status != core.TaskCompleted {
				hasVisibleTasks = true
				break
			}
		}
		if hasVisibleTasks {
			break
		}
	}

	if !hasVisibleTasks {
		var msg string
		if m.settings.ShowCompleted {
			msg = "No tasks found"
		} else {
			msg = "No pending tasks (completed tasks hidden)"
		}
		return lipgloss.NewStyle().Foreground(colors.Hint).Render(msg) + "\n\n" + m.list.View()
	}

	hint := "↑↓ • enter • c complete • d delete • e edit • m move • esc back"
	return m.list.View() + "\n\n" + lipgloss.NewStyle().Foreground(colors.Hint).Render(hint)
}

// Refresh refreshes the task list
func (m *TaskListModel) Refresh(todoFile *core.TodoFile) {
	m.todoFile = todoFile
	m.settings = config.GetSettings()
	items := buildTaskListItems(todoFile, m.settings)
	m.list.SetItems(items)
}

// SelectedTask returns the currently selected task ID, if any
func (m TaskListModel) SelectedTask() *int {
	if item, ok := m.list.SelectedItem().(taskListItem); ok && item.task != nil {
		return &item.task.ID
	}
	return nil
}

// TaskSelectedMsg is sent when a task is selected
type TaskSelectedMsg struct {
	TaskID int
}

// AddTaskMsg is sent when "Add new task" is selected
type AddTaskMsg struct{}

// BackMsg is sent to go back
type BackMsg struct{}

// ToggleTaskMsg is sent to toggle a task
type ToggleTaskMsg struct {
	TaskID int
}

// DeleteTaskMsg is sent to delete a task
type DeleteTaskMsg struct {
	TaskID int
}

// EditTaskMsg is sent to edit a task
type EditTaskMsg struct {
	TaskID int
}

// MoveTaskMsg is sent to move a task
type MoveTaskMsg struct {
	TaskID int
}

// FormatTaskListHints returns the keyboard hints
func FormatTaskListHints() string {
	var hints []string
	hints = append(hints, "↑↓")
	hints = append(hints, "enter")
	hints = append(hints, "c complete")
	hints = append(hints, "d delete")
	hints = append(hints, "e edit")
	hints = append(hints, "m move")
	hints = append(hints, "esc back")
	return strings.Join(hints, " • ")
}
