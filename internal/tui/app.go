package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/i-am-fran/markdowndo/internal/cli"
	"github.com/i-am-fran/markdowndo/internal/config"
	"github.com/i-am-fran/markdowndo/internal/core"
	"github.com/i-am-fran/markdowndo/internal/tui/colors"
	"github.com/i-am-fran/markdowndo/internal/tui/undo"
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
	showHelp       bool
	keys           KeyMap

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
	helpModel            help.Model
	statusBarModel       views.StatusBarModel
	paletteModel         views.PaletteModel
	toastModel           views.ToastModel
	showPalette          bool
	undoStack            *undo.Stack
}

// New creates a new TUI model
func New() Model {
	core.SetSectionAliases(config.GetSettings().SectionAliases)
	return Model{
		view:         ViewMenu,
		width:        80,
		height:       24,
		keys:         DefaultKeyMap(),
		helpModel:    help.New(),
		paletteModel: views.NewPaletteModel(80, 24),
		toastModel:   views.NewToastModel(80),
		undoStack:    undo.NewStack(5 * time.Second), // 5 second undo timeout
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return m.loadTodoFile()
}

func (m Model) loadTodoFile() tea.Cmd {
	return func() tea.Msg {
		todoFile, filePath, err := cli.LoadDefaultTodoFile()
		if err != nil {
			return todoFileLoadedMsg{err: err}
		}

		return todoFileLoadedMsg{todoFile: todoFile, filePath: filePath}
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
		m.statusBarModel.SetWidth(msg.Width)
		m.toastModel.SetWidth(msg.Width)
		m.paletteModel, _ = m.paletteModel.Update(msg)
		return m, nil

	case todoFileLoadedMsg:
		if msg.err != nil {
			m.message = &Message{Type: "error", Text: msg.err.Error()}
			return m, nil
		}
		m.todoFile = msg.todoFile
		m.filePath = msg.filePath
		m.menuModel = views.NewMenuModel(m.todoFile, m.width, m.height)
		m.statusBarModel = views.NewStatusBarModel(m.todoFile, m.filePath, m.width)
		return m, nil

	case reloadTodoFileMsg:
		if m.filePath != "" {
			todoFile, err := core.Load(m.filePath)
			if err == nil {
				m.todoFile = todoFile
				m.taskListModel.Refresh(todoFile)
				m.statusBarModel.Update(todoFile, m.filePath)
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

	case views.QuitMsg:
		return m, tea.Quit

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

	case views.ToggleShowCompletedMsg:
		settings := config.GetSettings()
		config.UpdateSettings(map[string]interface{}{
			"showCompleted": !settings.ShowCompleted,
		})
		config.ClearCache()
		m.taskListModel.Refresh(m.todoFile)
		return m, nil

	case views.PaletteSelectMsg:
		m.showPalette = false
		return m.handlePaletteAction(msg.Action)

	case views.ClosePaletteMsg:
		m.showPalette = false
		return m, nil

	case views.ToastTickMsg:
		var cmd tea.Cmd
		m.toastModel, cmd = m.toastModel.Update(msg)
		return m, cmd

	case views.AnimationTickMsg:
		// Update animations in task list
		if m.view == ViewTaskList {
			var cmd tea.Cmd
			m.taskListModel, cmd = m.taskListModel.Update(msg)
			return m, cmd
		}
		return m, nil

	case tea.KeyMsg:
		// Handle palette first if it's open
		if m.showPalette {
			var cmd tea.Cmd
			m.paletteModel, cmd = m.paletteModel.Update(msg)
			return m, cmd
		}

		// Global quit with ctrl+c
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		// Global help toggle with ?
		if msg.String() == "?" {
			m.showHelp = !m.showHelp
			return m, nil
		}
		// Global undo with u (if undo available)
		if msg.String() == "u" && !m.undoStack.IsEmpty() {
			return m.handleUndo()
		}
		// Global command palette toggle with ctrl+k
		if msg.String() == "ctrl+k" {
			m.showPalette = !m.showPalette
			return m, nil
		}
		// Global shortcut for adding new task with ctrl+n
		if msg.String() == "ctrl+n" {
			m.view = ViewAddTask
			m.addTaskModel = views.NewAddTaskModel(true, "", "", m.width, m.height)
			return m, m.addTaskModel.Init()
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

	// Create snapshot before toggle
	snapshot := m.todoFile.Snapshot()

	updated, err := cli.PerformToggleTask(m.todoFile, taskID)
	if err != nil {
		return m, nil
	}

	msg := "Task reopened"
	switch updated.Status {
	case core.TaskCompleted:
		msg = "Task completed"
	case core.TaskInProgress:
		msg = "Task started"
	}

	// Add to undo stack
	m.undoStack.Push(undo.Action{
		Description: msg,
		Snapshot:    snapshot,
		Timestamp:   time.Now(),
		FilePath:    m.filePath,
	})

	m.taskListModel.Refresh(m.todoFile)
	m.statusBarModel.Update(m.todoFile, m.filePath)

	// Start flash animation
	animCmd := m.taskListModel.StartFlashAnimation(taskID)

	// Show toast notification
	toastCmd := m.toastModel.Show(msg, 5*time.Second)

	// Return both commands
	return m, tea.Batch(toastCmd, animCmd)
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
		todoFile, err := cli.PerformAddNote(m.todoFile, m.filePath, text)
		if err != nil {
			m.message = &Message{Type: "error", Text: err.Error()}
			return m, nil
		}
		m.todoFile = todoFile
		m.message = &Message{Type: "success", Text: "Added note: " + text}
		m.view = ViewMenu
		m.menuModel = views.NewMenuModel(m.todoFile, m.width, m.height)
		m.statusBarModel.Update(m.todoFile, m.filePath)
		return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
			return clearMessageMsg{}
		})
	}

	todoFile, task, err := cli.PerformAddTask(m.todoFile, m.filePath, text)
	if err != nil {
		m.message = &Message{Type: "error", Text: err.Error()}
		return m, nil
	}
	m.todoFile = todoFile
	m.lastAddedTask = task.Text
	m.addTaskModel.SetLastAddedWithID(task.Text, task.ID)
	m.statusBarModel.Update(m.todoFile, m.filePath)

	return m, nil
}

func (m Model) handleTextPromptSubmit(text string) (tea.Model, tea.Cmd) {
	if m.view == ViewEditTask && m.selectedTaskID != nil && m.todoFile != nil {
		// Create snapshot before edit
		snapshot := m.todoFile.Snapshot()

		if _, err := cli.PerformEditTask(m.todoFile, *m.selectedTaskID, text); err != nil {
			return m, nil
		}

		// Add to undo stack
		m.undoStack.Push(undo.Action{
			Description: "Task updated",
			Snapshot:    snapshot,
			Timestamp:   time.Now(),
			FilePath:    m.filePath,
		})

		m.view = ViewTaskList
		m.selectedTaskID = nil
		m.taskListModel.Refresh(m.todoFile)
		m.statusBarModel.Update(m.todoFile, m.filePath)

		// Show toast notification
		return m, m.toastModel.Show("Task updated", 5*time.Second)
	}
	return m, nil
}

func (m Model) handleMoveTask(section *string, taskID int) (tea.Model, tea.Cmd) {
	if m.todoFile == nil {
		return m, nil
	}

	// Create snapshot before move
	snapshot := m.todoFile.Snapshot()

	if _, err := cli.PerformMoveTask(m.todoFile, taskID, section); err != nil {
		return m, nil
	}

	target := "Inbox"
	if section != nil {
		target = *section
	}

	// Add to undo stack
	m.undoStack.Push(undo.Action{
		Description: "Moved to " + target,
		Snapshot:    snapshot,
		Timestamp:   time.Now(),
		FilePath:    m.filePath,
	})

	m.view = ViewTaskList
	m.selectedTaskID = nil
	m.taskListModel.Refresh(m.todoFile)
	m.statusBarModel.Update(m.todoFile, m.filePath)

	// Show toast notification
	return m, m.toastModel.Show("Moved to "+target, 5*time.Second)
}

func (m Model) handleConfirm(confirmed bool) (tea.Model, tea.Cmd) {
	if !confirmed {
		m.view = ViewTaskActions
		return m, nil
	}

	if m.selectedTaskID != nil && m.todoFile != nil {
		taskID := *m.selectedTaskID

		// Create snapshot before delete
		snapshot := m.todoFile.Snapshot()

		if _, err := cli.PerformDeleteTask(m.todoFile, taskID); err != nil {
			return m, nil
		}

		// Add to undo stack
		m.undoStack.Push(undo.Action{
			Description: "Task deleted",
			Snapshot:    snapshot,
			Timestamp:   time.Now(),
			FilePath:    m.filePath,
		})

		m.view = ViewTaskList
		m.selectedTaskID = nil
		m.taskListModel.Refresh(m.todoFile)
		m.statusBarModel.Update(m.todoFile, m.filePath)

		// Start collapse animation
		animCmd := m.taskListModel.StartCollapseAnimation(taskID)

		// Show toast notification
		toastCmd := m.toastModel.Show("Task deleted", 5*time.Second)

		// Return both commands
		return m, tea.Batch(toastCmd, animCmd)
	}

	return m, nil
}

func (m Model) handleUndo() (tea.Model, tea.Cmd) {
	// Pop the most recent action
	action := m.undoStack.Pop()
	if action == nil {
		return m, nil
	}

	// Hide the toast
	m.toastModel.Hide()

	// Restore the snapshot
	if m.todoFile != nil && action.FilePath == m.filePath {
		if err := cli.PerformRestore(m.todoFile, action.Snapshot); err == nil {
			// Refresh UI
			m.taskListModel.Refresh(m.todoFile)
			m.statusBarModel.Update(m.todoFile, m.filePath)

			// Show success message
			m.message = &Message{Type: "success", Text: "Undo: " + action.Description}
		}
	}

	return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return clearMessageMsg{}
	})
}

func (m Model) handlePaletteAction(action views.PaletteAction) (tea.Model, tea.Cmd) {
	switch action {
	case views.PaletteActionAddTask:
		m.view = ViewAddTask
		m.addTaskModel = views.NewAddTaskModel(true, "", "", m.width, m.height)
		return m, m.addTaskModel.Init()
	case views.PaletteActionListTasks:
		m.view = ViewTaskList
		m.taskListModel = views.NewTaskListModel(m.todoFile, m.width, m.height)
		return m, nil
	case views.PaletteActionSearch:
		m.view = ViewSearch
		m.searchModel = views.NewSearchModel(m.width, m.height)
		return m, m.searchModel.Init()
	case views.PaletteActionSubfolders:
		m.view = ViewSubfolders
		m.subfoldersModel = views.NewSubfoldersModel(m.width, m.height)
		return m, m.subfoldersModel.Init()
	case views.PaletteActionSettings:
		m.view = ViewSettings
		m.settingsModel = views.NewSettingsModel(m.width, m.height)
		return m, nil
	case views.PaletteActionLint:
		m.view = ViewLint
		m.lintModel = views.NewLintModel(m.width, m.height)
		return m, m.lintModel.Init()
	case views.PaletteActionQuit:
		return m, tea.Quit
	}
	return m, nil
}

// View implements tea.Model
func (m Model) View() string {
	if m.todoFile == nil {
		return lipgloss.NewStyle().Foreground(colors.Hint).Render("Loading...")
	}

	var s string

	// Modern header with better spacing
	header := lipgloss.NewStyle().
		Background(colors.HeaderBG).
		Foreground(colors.HeaderFG).
		Padding(0, 2).
		Bold(true).
		Render(" MarkdownDO ")
	s += header + "\n"

	// Message toast with better styling
	if m.message != nil {
		var icon string
		var color lipgloss.TerminalColor
		switch m.message.Type {
		case "success":
			icon = "✓"
			color = colors.Success
		case "error":
			icon = "✗"
			color = colors.Error
		default:
			icon = "ℹ"
			color = colors.Info
		}
		messageStyle := lipgloss.NewStyle().
			Foreground(color).
			Bold(true).
			Padding(0, 1)
		s += messageStyle.Render(icon+" "+m.message.Text) + "\n\n"
	}

	// Escape hint
	if m.escHintShown && m.view == ViewMenu {
		s += lipgloss.NewStyle().Foreground(colors.Hint).Render("Press ESC again to quit") + "\n\n"
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

	// Status bar at the bottom (if enabled in settings)
	settings := config.GetSettings()
	if settings.ShowStatusBar {
		s += "\n" + m.statusBarModel.View()
	}

	// Help overlay
	if m.showHelp {
		helpView := m.helpModel.View(m.keys)
		helpStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colors.Border).
			Padding(1, 2).
			MaxWidth(m.width - 4)
		overlay := helpStyle.Render(helpView)

		// Center the help overlay
		lines := lipgloss.Height(s)
		helpHeight := lipgloss.Height(overlay)
		padding := (lines - helpHeight) / 2
		if padding < 0 {
			padding = 0
		}

		s = lipgloss.PlaceVertical(lines, lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Left,
				lipgloss.NewStyle().Height(padding).Render(""),
				lipgloss.PlaceHorizontal(m.width, lipgloss.Center, overlay),
			),
		)
	}

	// Command palette overlay (higher priority than help)
	if m.showPalette {
		s = m.paletteModel.View()
	}

	// Toast notification (highest priority, appears over everything)
	if m.toastModel.IsVisible() {
		toastView := m.toastModel.View()
		// Place toast at bottom center
		s += "\n\n" + lipgloss.PlaceHorizontal(m.width, lipgloss.Center, toastView)
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
