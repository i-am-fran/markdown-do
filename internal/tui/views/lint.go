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

// LintModel is the lint view model
type LintModel struct {
	list       list.Model
	filePath   string
	issues     []core.LintIssue
	fixedCount int
	taskStats  struct {
		total     int
		pending   int
		completed int
	}
	loading bool
	width   int
	height  int
}

// NewLintModel creates a new lint model
func NewLintModel(width, height int) LintModel {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.SetSpacing(0)

	l := list.New([]list.Item{backItem{}}, delegate, width, height-6)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	return LintModel{
		list:    l,
		loading: true,
		width:   width,
		height:  height,
	}
}

// Init implements tea.Model
func (m LintModel) Init() tea.Cmd {
	return m.runLint()
}

func (m LintModel) runLint() tea.Cmd {
	return func() tea.Msg {
		cwd, _ := os.Getwd()
		path, err := core.FindDefaultTodoFile(cwd)
		if err != nil {
			return lintCompleteMsg{}
		}

		todoFile, err := core.Load(path)
		if err != nil {
			return lintCompleteMsg{}
		}

		tasks := todoFile.GetTasks()
		pending := 0
		completed := 0
		for _, t := range tasks {
			if t.Status == core.TaskPending {
				pending++
			} else {
				completed++
			}
		}

		result := todoFile.Lint()

		if result.FixedCount > 0 {
			todoFile.Save()
		}

		return lintCompleteMsg{
			filePath:   path,
			issues:     result.Issues,
			fixedCount: result.FixedCount,
			total:      len(tasks),
			pending:    pending,
			completed:  completed,
		}
	}
}

// Update implements tea.Model
func (m LintModel) Update(msg tea.Msg) (LintModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-6)
		return m, nil

	case lintCompleteMsg:
		m.loading = false
		m.filePath = msg.filePath
		m.issues = msg.issues
		m.fixedCount = msg.fixedCount
		m.taskStats.total = msg.total
		m.taskStats.pending = msg.pending
		m.taskStats.completed = msg.completed
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "esc":
			return m, func() tea.Msg { return BackMsg{} }
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View implements tea.Model
func (m LintModel) View() string {
	if m.loading {
		loadingStyle := lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true)
		return loadingStyle.Render("Linting...")
	}

	var s string

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(colors.Accent).
		Bold(true).
		MarginBottom(1).
		MarginTop(1)
	s += titleStyle.Render("Lint Results") + "\n"

	// File path
	pathStyle := lipgloss.NewStyle().
		Foreground(colors.Muted).
		Italic(true).
		MarginBottom(1)
	s += pathStyle.Render(m.filePath) + "\n\n"

	if len(m.issues) == 0 {
		successStyle := lipgloss.NewStyle().
			Foreground(colors.Success).
			Bold(true)
		s += successStyle.Render("✓ No issues found") + "\n\n"
	} else {
		resultsTitleStyle := lipgloss.NewStyle().
			Foreground(colors.Accent).
			Bold(true).
			MarginBottom(1)
		s += resultsTitleStyle.Render("Issues:") + "\n"

		for _, issue := range m.issues {
			status := lipgloss.NewStyle().Foreground(colors.Warning).Render("[found]")
			if issue.Fixed {
				status = lipgloss.NewStyle().Foreground(colors.Success).Bold(true).Render("[fixed]")
			}
			issueStyle := lipgloss.NewStyle().
				Foreground(colors.Hint)
			s += issueStyle.Render(fmt.Sprintf("  Line %d: %s %s\n", issue.Line, issue.Issue, status))
		}

		if m.fixedCount > 0 {
			suffix := "s"
			if m.fixedCount == 1 {
				suffix = ""
			}
			fixedStyle := lipgloss.NewStyle().
				Foreground(colors.Success).
				Bold(true).
				MarginTop(1)
			s += "\n" + fixedStyle.Render(fmt.Sprintf("✓ Fixed %d issue%s", m.fixedCount, suffix)) + "\n"
		}
		s += "\n"
	}

	suffix := "s"
	if m.taskStats.total == 1 {
		suffix = ""
	}
	statsStyle := lipgloss.NewStyle().
		Foreground(colors.Muted).
		Italic(true).
		MarginBottom(1)
	s += statsStyle.Render(
		fmt.Sprintf("Checked %d task%s (%d pending, %d completed)",
			m.taskStats.total, suffix, m.taskStats.pending, m.taskStats.completed)) + "\n"

	s += m.list.View() + "\n\n"
	
	hintStyle := lipgloss.NewStyle().
		Foreground(colors.Hint).
		Italic(true)
	s += hintStyle.Render("esc back")

	return s
}

type lintCompleteMsg struct {
	filePath   string
	issues     []core.LintIssue
	fixedCount int
	total      int
	pending    int
	completed  int
}
