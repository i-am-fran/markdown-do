package tui

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/i-am-fran/markdowndo/internal/cli"
	"github.com/i-am-fran/markdowndo/internal/config"
	"github.com/i-am-fran/markdowndo/internal/core"
	"github.com/i-am-fran/markdowndo/internal/tui/views"
)

// View represents the current view
type View string

const (
	ViewMenu            View = "menu"
	ViewTaskList        View = "taskList"
	ViewTaskActions     View = "taskActions"
	ViewAddTask         View = "addTask"
	ViewAddNote         View = "addNote"
	ViewEditTask        View = "editTask"
	ViewMoveTask        View = "moveTask"
	ViewDeleteTask      View = "deleteTask"
	ViewSettings        View = "settings"
	ViewSearch          View = "search"
	ViewSubfolders      View = "subfolders"
	ViewLint            View = "lint"
	ViewDeleteCompleted View = "deleteCompleted"
)

// Message represents a toast message
type Message struct {
	Type string // "success", "error", "info"
	Text string
}

const doubleEscWindowMS = 500

// Model is the main TUI model
type Model struct {
	view           View
	todoFile       *core.TodoFile
	filePath       string
	selectedTaskID *int
	message        *Message
	lastEscTime    time.Time
	escHintShown   bool
	lastAddedTask  string
	width          int
	height         int

	// Sub-models for each view
	menuModel            views.MenuModel
	taskListModel        views.TaskListModel
	taskActionsModel     views.TaskActionsModel
	addTaskModel         views.AddTaskModel
	textPromptModel      views.TextPromptModel
	moveTaskModel        views.MoveTaskModel
	settingsModel        views.SettingsModel
	searchModel          views.SearchModel
	subfoldersModel      views.SubfoldersModel
	lintModel            views.LintModel
	deleteCompletedModel views.DeleteCompletedModel
	confirmModel         views.ConfirmModel
}

// New creates a new TUI model
func New() Model {
	return Model{
		view:   ViewMenu,
		width:  80,
		height: 24,
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return m.loadTodoFile()
}

func (m Model) loadTodoFile() tea.Cmd {
	return func() tea.Msg {
		cwd, err := os.Getwd()
		if err != nil {
			return todoFileLoadedMsg{err: err}
		}

		path, err := core.FindDefaultTodoFile(cwd)
		if err != nil {
			return todoFileLoadedMsg{err: err}
		}

		todoFile, err := core.Load(path)
		if err != nil {
			return todoFileLoadedMsg{err: err}
		}

		return todoFileLoadedMsg{todoFile: todoFile, filePath: path}
	}
}

type todoFileLoadedMsg struct {
	todoFile *core.TodoFile
	filePath string
	err      error
}

type reloadTodoFileMsg struct{}

type clearMessageMsg struct{}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case todoFileLoadedMsg:
		if msg.err != nil {
			m.message = &Message{Type: "error", Text: msg.err.Error()}
			return m, nil
		}
		m.todoFile = msg.todoFile
		m.filePath = msg.filePath
		m.menuModel = views.NewMenuModel(m.todoFile, m.width, m.height)
		return m, nil

	case reloadTodoFileMsg:
		if m.filePath != "" {
			todoFile, err := core.Load(m.filePath)
			if err == nil {
				m.todoFile = todoFile
				m.taskListModel.Refresh(todoFile)
			}
		}
		return m, nil

	case clearMessageMsg:
		m.message = nil
		return m, nil

	case views.MenuSelectMsg:
		return m.handleMenuSelect(msg.Action)

	case views.EscPressedMsg:
		return m.handleDoubleEsc()

	case views.TaskSelectedMsg:
		m.selectedTaskID = &msg.TaskID
		m.view = ViewTaskActions
		if m.todoFile != nil {
			if task := m.todoFile.GetTask(msg.TaskID); task != nil {
				m.taskActionsModel = views.NewTaskActionsModel(task, m.width, m.height)
			}
		}
		return m, nil

	case views.AddTaskMsg:
		m.view = ViewAddTask
		m.addTaskModel = views.NewAddTaskModel(true, "", "", m.width, m.height)
		return m, nil

	case views.BackMsg:
		return m.handleBack()

	case views.ToggleTaskMsg:
		return m.handleToggleTask(msg.TaskID)

	case views.DeleteTaskMsg:
		m.selectedTaskID = &msg.TaskID
		if m.todoFile != nil {
			if task := m.todoFile.GetTask(msg.TaskID); task != nil {
				m.confirmModel = views.NewConfirmModel(fmt.Sprintf("Delete \"%s\"?", task.Text), m.width, m.height)
			}
		}
		m.view = ViewDeleteTask
		return m, nil

	case views.EditTaskMsg:
		m.selectedTaskID = &msg.TaskID
		m.view = ViewEditTask
		if m.todoFile != nil {
			if task := m.todoFile.GetTask(msg.TaskID); task != nil {
				m.textPromptModel = views.NewTextPromptModel("Enter new text:", task.Text, "", m.width, m.height)
			}
		}
		return m, nil

	case views.MoveTaskMsg:
		m.selectedTaskID = &msg.TaskID
		m.view = ViewMoveTask
		if m.todoFile != nil {
			if task := m.todoFile.GetTask(msg.TaskID); task != nil {
				m.moveTaskModel = views.NewMoveTaskModel(m.todoFile, task, m.width, m.height)
			}
		}
		return m, nil

	case views.TaskActionSelectMsg:
		return m.handleTaskAction(msg.Action, msg.TaskID)

	case views.AddTaskSubmitMsg:
		return m.handleAddTaskSubmit(msg.Text)

	case views.AddTaskCancelMsg:
		m.lastAddedTask = ""
		m.view = ViewMenu
		m.menuModel = views.NewMenuModel(m.todoFile, m.width, m.height)
		return m, nil
	
	case views.OpenTaskMsg:
		// Open the task that was just created
		m.selectedTaskID = &msg.TaskID
		m.view = ViewTaskActions
		if m.todoFile != nil {
			if task := m.todoFile.GetTask(msg.TaskID); task != nil {
				m.taskActionsModel = views.NewTaskActionsModel(task, m.width, m.height)
			}
		}
		return m, nil

	case views.TextPromptSubmitMsg:
		return m.handleTextPromptSubmit(msg.Text)

	case views.TextPromptCancelMsg:
		if m.view == ViewEditTask {
			m.view = ViewTaskActions
		} else {
			m.view = ViewMenu
		}
		return m, nil

	case views.MoveTaskSelectMsg:
		return m.handleMoveTask(msg.Section, msg.TaskID)

	case views.ConfirmMsg:
		return m.handleConfirm(msg.Confirmed)

	case views.SettingsChangedMsg:
		config.ClearCache()
		return m, nil

	case views.DeleteCompletedMsg:
		return m, m.loadTodoFile()

	case views.SearchTaskActionCompleteMsg:
		m.searchModel.SetMessage(msg.Message)
		return m, m.searchModel.RerunSearch()

	case tea.KeyMsg:
		// Global quit with ctrl+c
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	// Update the current view's model
	var cmd tea.Cmd
	switch m.view {
	case ViewMenu:
		m.menuModel, cmd = m.menuModel.Update(msg)
	case ViewTaskList:
		m.taskListModel, cmd = m.taskListModel.Update(msg)
	case ViewTaskActions:
		m.taskActionsModel, cmd = m.taskActionsModel.Update(msg)
	case ViewAddTask, ViewAddNote:
		m.addTaskModel, cmd = m.addTaskModel.Update(msg)
	case ViewEditTask:
		m.textPromptModel, cmd = m.textPromptModel.Update(msg)
	case ViewMoveTask:
		m.moveTaskModel, cmd = m.moveTaskModel.Update(msg)
	case ViewDeleteTask:
		m.confirmModel, cmd = m.confirmModel.Update(msg)
	case ViewSettings:
		m.settingsModel, cmd = m.settingsModel.Update(msg)
	case ViewSearch:
		m.searchModel, cmd = m.searchModel.Update(msg)
	case ViewSubfolders:
		m.subfoldersModel, cmd = m.subfoldersModel.Update(msg)
	case ViewLint:
		m.lintModel, cmd = m.lintModel.Update(msg)
	case ViewDeleteCompleted:
		m.deleteCompletedModel, cmd = m.deleteCompletedModel.Update(msg)
	}

	return m, cmd
}

func (m Model) handleMenuSelect(action views.MenuAction) (tea.Model, tea.Cmd) {
	m.lastEscTime = time.Time{}

	switch action {
	case views.ActionList:
		m.view = ViewTaskList
		m.taskListModel = views.NewTaskListModel(m.todoFile, m.width, m.height)
	case views.ActionAdd:
		m.view = ViewAddTask
		m.addTaskModel = views.NewAddTaskModel(true, "", "", m.width, m.height)
	case views.ActionAddNote:
		m.view = ViewAddNote
		m.addTaskModel = views.NewAddTaskModel(false, "Enter note text:", "Remember to review the PR", m.width, m.height)
	case views.ActionFind:
		m.view = ViewSearch
		m.searchModel = views.NewSearchModel(m.width, m.height)
	case views.ActionDeleteCompleted:
		m.view = ViewDeleteCompleted
		m.deleteCompletedModel = views.NewDeleteCompletedModel(m.width, m.height)
		return m, m.deleteCompletedModel.Init()
	case views.ActionLint:
		m.view = ViewLint
		m.lintModel = views.NewLintModel(m.width, m.height)
		return m, m.lintModel.Init()
	case views.ActionOpen:
		cli.OpenInEditor()
		return m, m.loadTodoFile()
	case views.ActionSubfolders:
		m.view = ViewSubfolders
		m.subfoldersModel = views.NewSubfoldersModel(m.width, m.height)
		return m, m.subfoldersModel.Init()
	case views.ActionSettings:
		m.view = ViewSettings
		m.settingsModel = views.NewSettingsModel(m.width, m.height)
	case views.ActionQuit:
		return m, tea.Quit
	}

	return m, nil
}

func (m Model) handleDoubleEsc() (tea.Model, tea.Cmd) {
	now := time.Now()
	if now.Sub(m.lastEscTime) < time.Duration(doubleEscWindowMS)*time.Millisecond {
		return m, tea.Quit
	}

	m.lastEscTime = now
	m.escHintShown = true

	return m, tea.Tick(time.Duration(doubleEscWindowMS)*time.Millisecond, func(t time.Time) tea.Msg {
		return hideEscHintMsg{}
	})
}

type hideEscHintMsg struct{}

func (m Model) handleBack() (tea.Model, tea.Cmd) {
	switch m.view {
	case ViewTaskList, ViewSettings, ViewSearch, ViewSubfolders, ViewLint, ViewDeleteCompleted:
		m.view = ViewMenu
		m.menuModel = views.NewMenuModel(m.todoFile, m.width, m.height)
	case ViewTaskActions:
		m.view = ViewTaskList
		m.selectedTaskID = nil
	case ViewMoveTask, ViewEditTask:
		m.view = ViewTaskActions
	case ViewDeleteTask:
		m.view = ViewTaskActions
	}
	return m, nil
}

func (m Model) handleToggleTask(taskID int) (tea.Model, tea.Cmd) {
	if m.todoFile == nil {
		return m, nil
	}

	task := m.todoFile.GetTask(taskID)
	if task == nil {
		return m, nil
	}

	prevStatus := task.Status
	m.todoFile.ToggleTask(taskID)
	m.todoFile.Save()

	msg := "Task completed"
	if prevStatus == core.TaskCompleted {
		msg = "Task reopened"
	}
	m.message = &Message{Type: "success", Text: msg}

	m.taskListModel.Refresh(m.todoFile)

	return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return clearMessageMsg{}
	})
}

func (m Model) handleTaskAction(action views.TaskAction, taskID int) (tea.Model, tea.Cmd) {
	switch action {
	case views.TaskActionToggle:
		return m.handleToggleTask(taskID)
	case views.TaskActionEdit:
		m.selectedTaskID = &taskID
		m.view = ViewEditTask
		if task := m.todoFile.GetTask(taskID); task != nil {
			m.textPromptModel = views.NewTextPromptModel("Enter new text:", task.Text, "", m.width, m.height)
		}
	case views.TaskActionMove:
		m.selectedTaskID = &taskID
		m.view = ViewMoveTask
		if task := m.todoFile.GetTask(taskID); task != nil {
			m.moveTaskModel = views.NewMoveTaskModel(m.todoFile, task, m.width, m.height)
		}
	case views.TaskActionDelete:
		m.selectedTaskID = &taskID
		m.view = ViewDeleteTask
		if task := m.todoFile.GetTask(taskID); task != nil {
			m.confirmModel = views.NewConfirmModel(fmt.Sprintf("Delete \"%s\"?", task.Text), m.width, m.height)
		}
	case views.TaskActionBack:
		m.view = ViewTaskList
		m.selectedTaskID = nil
	}
	return m, nil
}

func (m Model) handleAddTaskSubmit(text string) (tea.Model, tea.Cmd) {
	if m.todoFile == nil {
		return m, nil
	}

	if m.view == ViewAddNote {
		m.todoFile.AddNote(text)
		m.todoFile.Save()
		m.message = &Message{Type: "success", Text: "Added note: " + text}
		m.view = ViewMenu
		m.menuModel = views.NewMenuModel(m.todoFile, m.width, m.height)
		return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
			return clearMessageMsg{}
		})
	}

	task, err := m.todoFile.AddTask(text)
	if err != nil {
		m.message = &Message{Type: "error", Text: err.Error()}
		return m, nil
	}
	m.todoFile.Save()
	m.lastAddedTask = task.Text
	m.addTaskModel.SetLastAddedWithID(task.Text, task.ID)

	return m, nil
}

func (m Model) handleTextPromptSubmit(text string) (tea.Model, tea.Cmd) {
	if m.view == ViewEditTask && m.selectedTaskID != nil && m.todoFile != nil {
		m.todoFile.UpdateTask(*m.selectedTaskID, text)
		m.todoFile.Save()
		m.message = &Message{Type: "success", Text: "Task updated"}
		m.view = ViewTaskList
		m.selectedTaskID = nil
		m.taskListModel.Refresh(m.todoFile)
		return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
			return clearMessageMsg{}
		})
	}
	return m, nil
}

func (m Model) handleMoveTask(section *string, taskID int) (tea.Model, tea.Cmd) {
	if m.todoFile == nil {
		return m, nil
	}

	m.todoFile.MoveTask(taskID, section)
	m.todoFile.Save()

	target := "Inbox"
	if section != nil {
		target = *section
	}
	m.message = &Message{Type: "success", Text: "Moved to " + target}
	m.view = ViewTaskList
	m.selectedTaskID = nil
	m.taskListModel.Refresh(m.todoFile)

	return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return clearMessageMsg{}
	})
}

func (m Model) handleConfirm(confirmed bool) (tea.Model, tea.Cmd) {
	if !confirmed {
		m.view = ViewTaskActions
		return m, nil
	}

	if m.selectedTaskID != nil && m.todoFile != nil {
		m.todoFile.DeleteTask(*m.selectedTaskID)
		m.todoFile.Save()
		m.message = &Message{Type: "success", Text: "Task deleted"}
		m.view = ViewTaskList
		m.selectedTaskID = nil
		m.taskListModel.Refresh(m.todoFile)
		return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
			return clearMessageMsg{}
		})
	}

	return m, nil
}

// View implements tea.Model
func (m Model) View() string {
	if m.todoFile == nil {
		return "Loading..."
	}

	var s string

	// Header
	header := lipgloss.NewStyle().
		Background(lipgloss.AdaptiveColor{Light: "#7D56F4", Dark: "#7D56F4"}).
		Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"}).
		Padding(0, 1).
		Render(" MarkdownDO ")
	s += header + "\n\n"

	// Message toast
	if m.message != nil {
		var icon string
		var color lipgloss.AdaptiveColor
		switch m.message.Type {
		case "success":
			icon = "✓"
			color = lipgloss.AdaptiveColor{Light: "#00A000", Dark: "#00FF00"}
		case "error":
			icon = "✗"
			color = lipgloss.AdaptiveColor{Light: "#FF0000", Dark: "#FF0000"}
		default:
			icon = "ℹ"
			color = lipgloss.AdaptiveColor{Light: "#0000FF", Dark: "#00BFFF"}
		}
		s += lipgloss.NewStyle().Foreground(color).Render(icon+" "+m.message.Text) + "\n\n"
	}

	// Escape hint
	if m.escHintShown && m.view == ViewMenu {
		s += lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#5C5C5C"}).Render("Press ESC again to quit") + "\n\n"
	}

	// Current view
	switch m.view {
	case ViewMenu:
		s += m.menuModel.View()
	case ViewTaskList:
		s += m.taskListModel.View()
	case ViewTaskActions:
		s += m.taskActionsModel.View()
	case ViewAddTask, ViewAddNote:
		s += m.addTaskModel.View()
	case ViewEditTask:
		s += m.textPromptModel.View()
	case ViewMoveTask:
		s += m.moveTaskModel.View()
	case ViewDeleteTask:
		s += m.confirmModel.View()
	case ViewSettings:
		s += m.settingsModel.View()
	case ViewSearch:
		s += m.searchModel.View()
	case ViewSubfolders:
		s += m.subfoldersModel.View()
	case ViewLint:
		s += m.lintModel.View()
	case ViewDeleteCompleted:
		s += m.deleteCompletedModel.View()
	default:
		s += "Unknown view"
	}

	return s
}

// Run starts the TUI
func Run() error {
	settings := config.GetSettings()

	p := tea.NewProgram(
		New(),
		tea.WithAltScreen(),
	)

	if !settings.Fullscreen {
		p = tea.NewProgram(New())
	}

	_, err := p.Run()
	return err
}
