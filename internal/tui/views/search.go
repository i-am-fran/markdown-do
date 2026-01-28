package views

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/i-am-fran/markdowndo/internal/core"
)

// SearchResult represents a search result
type SearchResult struct {
	FilePath     string
	RelativePath string
	Task         core.Task
}

type searchResultItem struct {
	result SearchResult
	index  int
}

func (i searchResultItem) Title() string {
	checkbox := "[ ]"
	if i.result.Task.Status == core.TaskCompleted {
		checkbox = "[x]"
	}
	return fmt.Sprintf("%d. %s %s", i.result.Task.ID, checkbox, i.result.Task.Text)
}
func (i searchResultItem) Description() string { return "" }
func (i searchResultItem) FilterValue() string { return i.Title() }

type backItem struct{}

func (i backItem) Title() string       { return "Back" }
func (i backItem) Description() string { return "" }
func (i backItem) FilterValue() string { return "back" }

type moveSectionItem struct {
	title   string
	section *string // nil means inbox
	isBack  bool
}

func (i moveSectionItem) Title() string       { return i.title }
func (i moveSectionItem) Description() string { return "" }
func (i moveSectionItem) FilterValue() string { return i.title }

// SearchModel is the search view model
type SearchModel struct {
	textInput      textinput.Model
	editInput      textinput.Model
	moveList       list.Model
	list           list.Model
	actionList     list.Model
	keyword        string
	recursive      bool
	results        []SearchResult
	selectedResult *SearchResult
	viewMode       string // "input", "recursive", "results", "taskActions", "confirmDelete", "edit", "move"
	searching      bool
	message        string
	width          int
	height         int
}

// NewSearchModel creates a new search model
func NewSearchModel(width, height int) SearchModel {
	ti := textinput.New()
	ti.Placeholder = "bug, feature, etc."
	ti.Focus()
	ti.CharLimit = 128
	ti.Width = width - 10
	
	// Edit input for editing task text
	ei := textinput.New()
	ei.CharLimit = 256
	ei.Width = width - 10

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.SetSpacing(0)

	l := list.New([]list.Item{}, delegate, width, height-6)
	l.Title = "Search results:"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.Styles.Title = lipgloss.NewStyle().Bold(true).MarginBottom(1)

	// Action list for task actions
	actionItems := []list.Item{
		settingsItem{title: "Mark as complete", action: "toggle"},
		settingsItem{title: "Edit text", action: "edit"},
		settingsItem{title: "Move to section", action: "move"},
		settingsItem{title: "Delete", action: "delete"},
		settingsItem{title: "Back", action: "back"},
	}
	al := list.New(actionItems, delegate, width, height-6)
	al.SetShowStatusBar(false)
	al.SetFilteringEnabled(false)
	al.SetShowHelp(false)
	al.Styles.Title = lipgloss.NewStyle().Bold(true).MarginBottom(1)
	
	// Move list (will be populated when needed)
	ml := list.New([]list.Item{}, delegate, width, height-6)
	ml.Title = "Move to section:"
	ml.SetShowStatusBar(false)
	ml.SetFilteringEnabled(false)
	ml.SetShowHelp(false)
	ml.Styles.Title = lipgloss.NewStyle().Bold(true).MarginBottom(1)

	return SearchModel{
		textInput:  ti,
		editInput:  ei,
		list:       l,
		actionList: al,
		moveList:   ml,
		viewMode:   "input",
		width:      width,
		height:     height,
	}
}

// Init implements tea.Model
func (m SearchModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model
func (m SearchModel) Update(msg tea.Msg) (SearchModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textInput.Width = msg.Width - 10
		m.editInput.Width = msg.Width - 10
		m.list.SetSize(msg.Width, msg.Height-6)
		m.actionList.SetSize(msg.Width, msg.Height-6)
		m.moveList.SetSize(msg.Width, msg.Height-6)
		return m, nil

	case searchCompleteMsg:
		m.results = msg.results
		m.searching = false
		m.viewMode = "results"
		m.updateResultList()
		return m, nil

	case tea.KeyMsg:
		switch m.viewMode {
		case "input":
			switch msg.String() {
			case "enter":
				text := m.textInput.Value()
				if text == "" {
					return m, func() tea.Msg { return BackMsg{} }
				}
				m.keyword = text
				m.viewMode = "recursive"
				return m, nil
			case "esc":
				return m, func() tea.Msg { return BackMsg{} }
			}

			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd

		case "recursive":
			switch msg.String() {
			case "y", "Y":
				m.recursive = true
				m.searching = true
				m.viewMode = "searching"
				return m, m.performSearch()
			case "n", "N":
				m.recursive = false
				m.searching = true
				m.viewMode = "searching"
				return m, m.performSearch()
			case "esc":
				m.viewMode = "input"
				return m, nil
			}
			return m, nil

		case "searching":
			return m, nil

		case "results":
			switch msg.String() {
			case "enter":
				if item, ok := m.list.SelectedItem().(searchResultItem); ok {
					m.selectedResult = &item.result
					m.updateActionList()
					m.viewMode = "taskActions"
					return m, nil
				}
				if _, ok := m.list.SelectedItem().(backItem); ok {
					return m, func() tea.Msg { return BackMsg{} }
				}
			case "esc":
				m.viewMode = "input"
				m.results = nil
				return m, nil
			}

			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd

		case "taskActions":
			switch msg.String() {
			case "enter":
				if item, ok := m.actionList.SelectedItem().(settingsItem); ok {
					switch item.action {
					case "toggle":
						return m, m.toggleSelectedTask()
					case "edit":
						m.viewMode = "edit"
						m.setupEditMode()
						return m, textinput.Blink
					case "move":
						m.viewMode = "move"
						m.setupMoveMode()
						return m, nil
					case "delete":
						m.viewMode = "confirmDelete"
						return m, nil
					case "back":
						m.viewMode = "results"
						m.selectedResult = nil
						return m, nil
					}
				}
			case "esc":
				m.viewMode = "results"
				m.selectedResult = nil
				return m, nil
			}

			var cmd tea.Cmd
			m.actionList, cmd = m.actionList.Update(msg)
			return m, cmd

		case "confirmDelete":
			switch msg.String() {
			case "y", "Y":
				return m, m.deleteSelectedTask()
			case "n", "N", "esc":
				m.viewMode = "taskActions"
				return m, nil
			}
			return m, nil
			
		case "edit":
			switch msg.String() {
			case "enter":
				text := m.editInput.Value()
				if text != "" {
					return m, m.editSelectedTask(text)
				}
				m.viewMode = "taskActions"
				return m, nil
			case "esc":
				m.viewMode = "taskActions"
				return m, nil
			}
			
			var cmd tea.Cmd
			m.editInput, cmd = m.editInput.Update(msg)
			return m, cmd
			
		case "move":
			switch msg.String() {
			case "enter":
				if item, ok := m.moveList.SelectedItem().(moveSectionItem); ok {
					if item.isBack {
						m.viewMode = "taskActions"
						return m, nil
					}
					return m, m.moveSelectedTask(item.section)
				}
			case "esc":
				m.viewMode = "taskActions"
				return m, nil
			}
			
			var cmd tea.Cmd
			m.moveList, cmd = m.moveList.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m *SearchModel) performSearch() tea.Cmd {
	return func() tea.Msg {
		cwd, _ := os.Getwd()
		files, err := core.FindTodoFiles(cwd, m.recursive)
		if err != nil {
			return searchCompleteMsg{results: nil}
		}

		var results []SearchResult
		for _, file := range files {
			todoFile, err := core.Load(file.Path)
			if err != nil {
				continue
			}

			matches := todoFile.FindTasks(m.keyword)
			for _, task := range matches {
				results = append(results, SearchResult{
					FilePath:     file.Path,
					RelativePath: file.RelativePath,
					Task:         task,
				})
			}
		}

		return searchCompleteMsg{results: results}
	}
}

func (m *SearchModel) updateResultList() {
	var items []list.Item
	for i, result := range m.results {
		items = append(items, searchResultItem{result: result, index: i})
	}
	items = append(items, backItem{})
	m.list.SetItems(items)
	m.list.Title = fmt.Sprintf("Search results for: \"%s\"", m.keyword)
}

func (m *SearchModel) updateActionList() {
	if m.selectedResult == nil {
		return
	}

	toggleLabel := "Mark as complete"
	if m.selectedResult.Task.Status == core.TaskCompleted {
		toggleLabel = "Reopen task"
	}

	items := []list.Item{
		settingsItem{title: toggleLabel, action: "toggle"},
		settingsItem{title: "Edit text", action: "edit"},
		settingsItem{title: "Move to section", action: "move"},
		settingsItem{title: "Delete", action: "delete"},
		settingsItem{title: "Back", action: "back"},
	}
	m.actionList.SetItems(items)
	m.actionList.Title = fmt.Sprintf("Task: %s", m.selectedResult.Task.Text)
}

func (m *SearchModel) toggleSelectedTask() tea.Cmd {
	return func() tea.Msg {
		if m.selectedResult == nil {
			return nil
		}

		todoFile, err := core.Load(m.selectedResult.FilePath)
		if err != nil {
			return nil
		}

		task := todoFile.GetTask(m.selectedResult.Task.ID)
		if task == nil {
			return nil
		}

		prevStatus := task.Status
		todoFile.ToggleTask(m.selectedResult.Task.ID)
		todoFile.Save()

		msg := "Task completed"
		if prevStatus == core.TaskCompleted {
			msg = "Task reopened"
		}

		return SearchTaskActionCompleteMsg{Message: msg, Keyword: m.keyword, Recursive: m.recursive}
	}
}

func (m *SearchModel) deleteSelectedTask() tea.Cmd {
	return func() tea.Msg {
		if m.selectedResult == nil {
			return nil
		}

		todoFile, err := core.Load(m.selectedResult.FilePath)
		if err != nil {
			return nil
		}

		todoFile.DeleteTask(m.selectedResult.Task.ID)
		todoFile.Save()

		return SearchTaskActionCompleteMsg{Message: "Task deleted", Keyword: m.keyword, Recursive: m.recursive}
	}
}

func (m *SearchModel) editSelectedTask(newText string) tea.Cmd {
	return func() tea.Msg {
		if m.selectedResult == nil {
			return nil
		}

		todoFile, err := core.Load(m.selectedResult.FilePath)
		if err != nil {
			return nil
		}

		todoFile.UpdateTask(m.selectedResult.Task.ID, newText)
		todoFile.Save()

		return SearchTaskActionCompleteMsg{Message: "Task updated", Keyword: m.keyword, Recursive: m.recursive}
	}
}

func (m *SearchModel) moveSelectedTask(section *string) tea.Cmd {
	return func() tea.Msg {
		if m.selectedResult == nil {
			return nil
		}

		todoFile, err := core.Load(m.selectedResult.FilePath)
		if err != nil {
			return nil
		}

		todoFile.MoveTask(m.selectedResult.Task.ID, section)
		todoFile.Save()

		target := "Inbox"
		if section != nil {
			target = *section
		}

		return SearchTaskActionCompleteMsg{Message: "Moved to " + target, Keyword: m.keyword, Recursive: m.recursive}
	}
}

func (m *SearchModel) setupEditMode() {
	if m.selectedResult != nil {
		m.editInput.SetValue(m.selectedResult.Task.Text)
		m.editInput.Focus()
	}
}

func (m *SearchModel) setupMoveMode() {
	if m.selectedResult == nil {
		return
	}

	todoFile, err := core.Load(m.selectedResult.FilePath)
	if err != nil {
		return
	}

	sections := todoFile.GetSectionNames()
	var items []list.Item

	// Add inbox option if task is in a section
	task := todoFile.GetTask(m.selectedResult.Task.ID)
	if task != nil && task.Section != nil {
		items = append(items, moveSectionItem{title: "Inbox (no section)", section: nil})
	}

	// Add sections (except current one)
	for _, section := range sections {
		if task == nil || task.Section == nil || section != *task.Section {
			s := section
			items = append(items, moveSectionItem{title: section, section: &s})
		}
	}

	items = append(items, moveSectionItem{title: "Back", isBack: true})
	m.moveList.SetItems(items)
}

// View implements tea.Model
func (m SearchModel) View() string {
	switch m.viewMode {
	case "input":
		s := lipgloss.NewStyle().Bold(true).Render("Enter search keyword:") + "\n\n"
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("> ") + m.textInput.View() + "\n\n"
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("enter search • esc cancel")
		return s

	case "recursive":
		s := lipgloss.NewStyle().Bold(true).Render("Search in subfolders?") + "\n\n"
		s += "Press " + lipgloss.NewStyle().Bold(true).Render("y") + " for yes, " + lipgloss.NewStyle().Bold(true).Render("n") + " for no"
		return s

	case "searching":
		return "Searching..."

	case "results":
		if m.message != "" {
			msg := lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("✓ "+m.message) + "\n\n"
			m.message = ""
			return msg + m.list.View() + "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("↑↓ navigate • enter select • esc back")
		}

		if len(m.results) == 0 {
			return lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render("No matching tasks found") + "\n\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("esc back")
		}

		hint := "↑↓ navigate • enter select • esc back"
		return m.list.View() + "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(hint)

	case "taskActions":
		if m.selectedResult == nil {
			return "No task selected"
		}

		var s string
		if m.recursive {
			s = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("File: "+m.selectedResult.RelativePath) + "\n\n"
		}
		hint := "↑↓ navigate • enter select • esc back"
		return s + m.actionList.View() + "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(hint)

	case "confirmDelete":
		if m.selectedResult == nil {
			return "No task selected"
		}
		s := lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("Delete \"%s\"?", m.selectedResult.Task.Text)) + "\n\n"
		s += "Press " + lipgloss.NewStyle().Bold(true).Render("y") + " for yes, " + lipgloss.NewStyle().Bold(true).Render("n") + " for no"
		return s
		
	case "edit":
		if m.selectedResult == nil {
			return "No task selected"
		}
		s := lipgloss.NewStyle().Bold(true).Render("Edit task:") + "\n\n"
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("> ") + m.editInput.View() + "\n\n"
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("enter save • esc cancel")
		return s
		
	case "move":
		if m.selectedResult == nil {
			return "No task selected"
		}
		hint := "↑↓ navigate • enter select • esc back"
		return m.moveList.View() + "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(hint)
	}

	return ""
}

// SetMessage sets a message to display
func (m *SearchModel) SetMessage(msg string) {
	m.message = msg
}

// RerunSearch re-runs the current search
func (m *SearchModel) RerunSearch() tea.Cmd {
	m.viewMode = "results"
	m.selectedResult = nil
	return m.performSearch()
}

type searchCompleteMsg struct {
	results []SearchResult
}

// SearchTaskActionCompleteMsg is sent when a search task action completes
type SearchTaskActionCompleteMsg struct {
	Message   string
	Keyword   string
	Recursive bool
}

// HandleSearchTaskActionComplete handles the search task action complete message
func (m *SearchModel) HandleSearchTaskActionComplete(msg SearchTaskActionCompleteMsg) tea.Cmd {
	m.message = msg.Message
	m.keyword = msg.Keyword
	m.recursive = msg.Recursive
	m.selectedResult = nil
	return m.performSearch()
}
