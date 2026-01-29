package views

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/i-am-fran/markdowndo/internal/core"
	"github.com/i-am-fran/markdowndo/internal/tui/colors"
)

// DeleteCompletedModel is the delete completed tasks view model
type DeleteCompletedModel struct {
	list           list.Model
	completedCount int
	deletedCount   int
	viewMode       string // "confirm" or "result"
	loading        bool
	width          int
	height         int
}

// NewDeleteCompletedModel creates a new delete completed model
func NewDeleteCompletedModel(width, height int) DeleteCompletedModel {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.SetSpacing(0)

	l := list.New([]list.Item{backItem{}}, delegate, width, height-6)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	return DeleteCompletedModel{
		list:     l,
		viewMode: "confirm",
		loading:  true,
		width:    width,
		height:   height,
	}
}

// Init implements tea.Model
func (m DeleteCompletedModel) Init() tea.Cmd {
	return m.loadCompletedCount()
}

func (m DeleteCompletedModel) loadCompletedCount() tea.Cmd {
	return func() tea.Msg {
		cwd, _ := os.Getwd()
		path, err := core.FindDefaultTodoFile(cwd)
		if err != nil {
			return deleteCompletedCountMsg{count: 0}
		}

		todoFile, err := core.Load(path)
		if err != nil {
			return deleteCompletedCountMsg{count: 0}
		}

		completed := todoFile.GetCompletedTasks()
		return deleteCompletedCountMsg{count: len(completed)}
	}
}

// Update implements tea.Model
func (m DeleteCompletedModel) Update(msg tea.Msg) (DeleteCompletedModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-6)
		return m, nil

	case deleteCompletedCountMsg:
		m.loading = false
		m.completedCount = msg.count
		return m, nil

	case deleteCompletedDoneMsg:
		m.deletedCount = msg.count
		m.viewMode = "result"
		return m, func() tea.Msg { return DeleteCompletedMsg{Count: msg.count} }

	case tea.KeyMsg:
		if m.viewMode == "result" {
			switch msg.String() {
			case "enter", "esc":
				return m, func() tea.Msg { return BackMsg{} }
			}
			return m, nil
		}

		switch msg.String() {
		case "y", "Y":
			return m, m.performDelete()
		case "n", "N", "esc":
			return m, func() tea.Msg { return BackMsg{} }
		}
	}

	return m, nil
}

func (m DeleteCompletedModel) performDelete() tea.Cmd {
	return func() tea.Msg {
		cwd, _ := os.Getwd()
		path, err := core.FindDefaultTodoFile(cwd)
		if err != nil {
			return deleteCompletedDoneMsg{count: 0}
		}

		todoFile, err := core.Load(path)
		if err != nil {
			return deleteCompletedDoneMsg{count: 0}
		}

		count := todoFile.DeleteCompletedTasks()
		todoFile.Save()

		return deleteCompletedDoneMsg{count: count}
	}
}

// View implements tea.Model
func (m DeleteCompletedModel) View() string {
	if m.loading {
		return lipgloss.NewStyle().Foreground(colors.Hint).Render("Loading...")
	}

	if m.completedCount == 0 {
		return lipgloss.NewStyle().Foreground(colors.Warning).Render("No completed tasks to delete") + "\n\n" +
			m.list.View() + "\n\n" +
			lipgloss.NewStyle().Foreground(colors.Hint).Render("esc back")
	}

	if m.viewMode == "result" {
		suffix := "s"
		if m.deletedCount == 1 {
			suffix = ""
		}
		return lipgloss.NewStyle().Foreground(colors.Success).Render(
			fmt.Sprintf("✓ Deleted %d completed task%s", m.deletedCount, suffix)) + "\n\n" +
			m.list.View() + "\n\n" +
			lipgloss.NewStyle().Foreground(colors.Hint).Render("esc back")
	}

	suffix := "s"
	if m.completedCount == 1 {
		suffix = ""
	}
	s := lipgloss.NewStyle().Bold(true).Render(
		fmt.Sprintf("Delete %d completed task%s?", m.completedCount, suffix)) + "\n\n"
	s += lipgloss.NewStyle().Foreground(colors.Hint).Render("y yes  n no")

	return s
}

type deleteCompletedCountMsg struct {
	count int
}

type deleteCompletedDoneMsg struct {
	count int
}

// DeleteCompletedMsg is sent when completed tasks are deleted
type DeleteCompletedMsg struct {
	Count int
}
