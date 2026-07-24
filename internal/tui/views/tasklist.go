package views

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/paginator"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/i-am-fran/markdowndo/internal/config"
	"github.com/i-am-fran/markdowndo/internal/core"
	"github.com/i-am-fran/markdowndo/internal/tui/animation"
	"github.com/i-am-fran/markdowndo/internal/tui/colors"
)

type taskListItem struct {
	task         *core.Task
	section      string
	sectionCount int // Total tasks in section (for display)
	isSection    bool
	isAdd        bool
}

func (i taskListItem) Title() string {
	if i.isSection {
		// Use box drawing characters for better visual separation
		separator := "━━"
		if i.sectionCount > 0 {
			return fmt.Sprintf("  %s %s (%d) %s", separator, i.section, i.sectionCount, separator)
		}
		return fmt.Sprintf("  %s %s %s", separator, i.section, separator)
	}
	if i.isAdd {
		return "  + Add new task"
	}

	checkbox := "☐"
	switch i.task.Status {
	case core.TaskCompleted:
		checkbox = "☑"
	case core.TaskInProgress:
		checkbox = "◐"
	}
	return fmt.Sprintf("  %s %s", checkbox, i.task.Text)
}

func (i taskListItem) Description() string { return "" }
func (i taskListItem) FilterValue() string { return i.Title() }

// TaskListModel is the task list view model
type TaskListModel struct {
	list       list.Model
	paginator  paginator.Model
	todoFile   *core.TodoFile
	settings   config.Settings
	width      int
	height     int
	animations map[int]*animation.State // Task ID -> animation state
}

// NewTaskListModel creates a new task list model
func NewTaskListModel(todoFile *core.TodoFile, width, height int) TaskListModel {
	settings := config.GetSettings()

	items := buildTaskListItems(todoFile, settings)

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.SetSpacing(0)
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(colors.Selected).
		Bold(true)
	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.
		Foreground(lipgloss.NoColor{})

	l := list.New(items, delegate, width, height-6)
	l.Title = "Select a task:"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true) // Enable live filtering
	l.SetShowHelp(false)
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(colors.Accent).
		Bold(true).
		MarginBottom(1).
		MarginTop(1)
	l.Styles.NoItems = lipgloss.NewStyle().
		Foreground(colors.Muted).
		Italic(true)

	// Initialize paginator
	p := paginator.New()
	p.Type = paginator.Dots
	p.ActiveDot = lipgloss.NewStyle().Foreground(colors.Selected).Render("●")
	p.InactiveDot = lipgloss.NewStyle().Foreground(colors.Muted).Render("○")
	p.PerPage = height - 10 // Account for header, title, hints, status bar, etc.
	// Calculate number of pages (ceiling division)
	totalPages := (len(items) + p.PerPage - 1) / p.PerPage
	if p.PerPage <= 0 {
		p.PerPage = 1
		totalPages = len(items)
	}
	p.SetTotalPages(totalPages)

	return TaskListModel{
		list:       l,
		paginator:  p,
		todoFile:   todoFile,
		settings:   settings,
		width:      width,
		height:     height,
		animations: make(map[int]*animation.State),
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
			// Count total tasks in section (including completed if showing them)
			totalInSection := len(group.Tasks)
			items = append(items, taskListItem{
				isSection:    true,
				section:      *group.Section,
				sectionCount: totalInSection,
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

	case AnimationTickMsg:
		// Update all active animations
		anyActive := false
		for taskID, state := range m.animations {
			if state.Update() {
				anyActive = true
			} else {
				delete(m.animations, taskID)
			}
		}

		// Continue ticking if any animations are active
		if anyActive && m.settings.EnableAnimations {
			return m, m.animationTick()
		}
		return m, nil

	case tea.KeyMsg:
		// Handle filtering state first - let the list handle filter input
		if m.list.FilterState() == list.Filtering {
			// When filtering, ESC should exit filter mode, not go back
			if msg.String() == "esc" {
				m.list.ResetFilter()
				return m, nil
			}
			// Let the list handle all other keys during filtering
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			// Update paginator after list update
			if m.paginator.PerPage > 0 {
				currentPage := m.list.Index() / m.paginator.PerPage
				m.paginator.Page = currentPage
			}
			return m, cmd
		}

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

		case key.Matches(msg, key.NewBinding(key.WithKeys("v"))):
			return m, func() tea.Msg { return ToggleShowCompletedMsg{} }

		case key.Matches(msg, key.NewBinding(key.WithKeys("/"))):
			// Start filtering - let the list handle the / key to enter filter mode
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd

		case key.Matches(msg, key.NewBinding(key.WithKeys("pgup"))):
			// Page up - jump to previous page
			if m.paginator.PerPage > 0 && m.paginator.Page > 0 {
				newIndex := (m.paginator.Page - 1) * m.paginator.PerPage
				if newIndex >= 0 && newIndex < m.list.Index() {
					steps := m.list.Index() - newIndex
					for i := 0; i < steps; i++ {
						m.list.CursorUp()
					}
				}
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("pgdown"))):
			// Page down - jump to next page
			if m.paginator.PerPage > 0 && m.paginator.Page < m.paginator.TotalPages-1 {
				newIndex := (m.paginator.Page + 1) * m.paginator.PerPage
				totalItems := len(m.list.Items())
				if newIndex >= totalItems {
					newIndex = totalItems - 1
				}
				if newIndex > m.list.Index() {
					steps := newIndex - m.list.Index()
					for i := 0; i < steps; i++ {
						m.list.CursorDown()
					}
				}
			}
		}
	}

	// Update the list for all other messages
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	// Update paginator based on list cursor
	if m.paginator.PerPage > 0 {
		currentPage := m.list.Index() / m.paginator.PerPage
		m.paginator.Page = currentPage
	}

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
		// Better empty state with helpful message
		emptyCardStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colors.Border).
			Padding(2, 3).
			Width(m.width - 4).
			MarginTop(2)

		var title, hint string
		if m.settings.ShowCompleted {
			title = "No tasks found"
			hint = "Press + or Ctrl+N to add your first task"
		} else {
			title = "All done! No pending tasks"
			hint = "Press v to show completed tasks, or + to add a new task"
		}

		titleStyle := lipgloss.NewStyle().
			Foreground(colors.Accent).
			Bold(true).
			MarginBottom(1)

		hintStyle := lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true)

		content := titleStyle.Render(title) + "\n" + hintStyle.Render(hint)
		return emptyCardStyle.Render(content) + "\n\n" + m.list.View()
	}

	// Build hint with toggle state indicator
	toggleHint := "v show done"
	if m.settings.ShowCompleted {
		toggleHint = "v hide done"
	}
	filterHint := ""
	if m.list.FilterState() == list.Filtering {
		filterHint = "  / clear filter"
	} else {
		filterHint = "  / filter"
	}
	hint := fmt.Sprintf("c complete  d delete  e edit  m move  %s%s  esc back", toggleHint, filterHint)

	view := m.list.View()

	// Show paginator if there are multiple pages
	if m.paginator.TotalPages > 1 {
		paginatorView := "\n" + lipgloss.NewStyle().
			Foreground(colors.Muted).
			Align(lipgloss.Center).
			Width(m.width).
			Render(m.paginator.View())
		view += paginatorView
	}

	hintStyle := lipgloss.NewStyle().
		Foreground(colors.Hint).
		Italic(true)

	return view + "\n\n" + hintStyle.Render(hint)
}

// Refresh refreshes the task list
func (m *TaskListModel) Refresh(todoFile *core.TodoFile) {
	m.todoFile = todoFile
	m.settings = config.GetSettings()
	items := buildTaskListItems(todoFile, m.settings)
	m.list.SetItems(items)
	// Calculate number of pages (ceiling division)
	totalPages := (len(items) + m.paginator.PerPage - 1) / m.paginator.PerPage
	if m.paginator.PerPage <= 0 {
		m.paginator.PerPage = 1
		totalPages = len(items)
	}
	m.paginator.SetTotalPages(totalPages)
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

// ToggleShowCompletedMsg is sent to toggle showing completed tasks
type ToggleShowCompletedMsg struct{}

// AnimationTickMsg is sent to update animations
type AnimationTickMsg struct {
	Time time.Time
}

// animationTick returns a tick command for animations
func (m TaskListModel) animationTick() tea.Cmd {
	return tea.Tick(16*time.Millisecond, func(t time.Time) tea.Msg {
		return AnimationTickMsg{Time: t}
	})
}

// StartFlashAnimation starts a flash animation for a task
func (m *TaskListModel) StartFlashAnimation(taskID int) tea.Cmd {
	if !m.settings.EnableAnimations {
		return nil
	}

	m.animations[taskID] = animation.NewFlashAnimation()
	return m.animationTick()
}

// StartCollapseAnimation starts a collapse animation for a task
func (m *TaskListModel) StartCollapseAnimation(taskID int) tea.Cmd {
	if !m.settings.EnableAnimations {
		return nil
	}

	m.animations[taskID] = animation.NewCollapseAnimation()
	return m.animationTick()
}

// FormatTaskListHints returns the keyboard hints
func FormatTaskListHints(showCompleted bool) string {
	toggleHint := "v show done"
	if showCompleted {
		toggleHint = "v hide done"
	}
	return fmt.Sprintf("c complete  d delete  e edit  m move  %s  esc back", toggleHint)
}
