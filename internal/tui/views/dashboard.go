package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/i-am-fran/markdowndo/internal/core"
	"github.com/i-am-fran/markdowndo/internal/tui/colors"
)

// DashboardModel is the dashboard view model
type DashboardModel struct {
	todoFile *core.TodoFile
	width    int
	height   int
}

// NewDashboardModel creates a new dashboard model
func NewDashboardModel(todoFile *core.TodoFile, width, height int) DashboardModel {
	return DashboardModel{
		todoFile: todoFile,
		width:    width,
		height:   height,
	}
}

// Init implements tea.Model
func (m DashboardModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m DashboardModel) Update(msg tea.Msg) (DashboardModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return m, func() tea.Msg { return BackMsg{} }
		}
	}

	return m, nil
}

// View implements tea.Model
func (m DashboardModel) View() string {
	if m.todoFile == nil {
		return ""
	}

	tasks := m.todoFile.GetTasks()
	total := len(tasks)
	pending := 0
	completed := 0
	
	// Count by section
	sectionStats := make(map[string]struct {
		total     int
		pending   int
		completed int
	})
	
	for _, task := range tasks {
		if task.Status == core.TaskPending {
			pending++
		} else {
			completed++
		}
		
		sectionName := "Inbox"
		if task.Section != nil {
			sectionName = *task.Section
		}
		
		stats := sectionStats[sectionName]
		stats.total++
		if task.Status == core.TaskPending {
			stats.pending++
		} else {
			stats.completed++
		}
		sectionStats[sectionName] = stats
	}

	var s string

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(colors.Accent).
		Bold(true).
		MarginBottom(1).
		MarginTop(1)
	s += titleStyle.Render("Dashboard") + "\n\n"

	// Overall stats card
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colors.Border).
		Padding(1, 2).
		Width(m.width - 4).
		MarginBottom(1)

	// Calculate completion percentage
	completionPercent := 0.0
	if total > 0 {
		completionPercent = float64(completed) / float64(total) * 100
	}

	// Progress bar
	progressBarWidth := 40
	filled := int(completionPercent / 100 * float64(progressBarWidth))
	if filled > progressBarWidth {
		filled = progressBarWidth
	}

	progressBar := strings.Repeat("█", filled) + strings.Repeat("░", progressBarWidth-filled)
	
	progressBarStyle := lipgloss.NewStyle().
		Foreground(colors.Success)
	if completionPercent < 50 {
		progressBarStyle = progressBarStyle.Foreground(colors.Warning)
	}

	statsContent := fmt.Sprintf(
		"Total: %d  |  Pending: %d  |  Completed: %d\n\n%s  %.1f%%",
		total, pending, completed,
		progressBarStyle.Render(progressBar),
		completionPercent,
	)

	s += cardStyle.Render(statsContent) + "\n\n"

	// Section stats
	if len(sectionStats) > 0 {
		sectionTitleStyle := lipgloss.NewStyle().
			Foreground(colors.Accent).
			Bold(true).
			MarginBottom(1)
		s += sectionTitleStyle.Render("By Section") + "\n\n"

		// Calculate card width (2 columns)
		cardWidth := (m.width - 8) / 2
		if cardWidth < 20 {
			cardWidth = m.width - 4
		}

		sectionCards := []string{}
		for sectionName, stats := range sectionStats {
			sectionCardStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colors.Border).
				Padding(1, 2).
				Width(cardWidth).
				MarginBottom(1)

			sectionPercent := 0.0
			if stats.total > 0 {
				sectionPercent = float64(stats.completed) / float64(stats.total) * 100
			}

			sectionBarWidth := 20
			sectionFilled := int(sectionPercent / 100 * float64(sectionBarWidth))
			if sectionFilled > sectionBarWidth {
				sectionFilled = sectionBarWidth
			}

			sectionBar := strings.Repeat("█", sectionFilled) + strings.Repeat("░", sectionBarWidth-sectionFilled)
			
			sectionBarStyle := lipgloss.NewStyle().
				Foreground(colors.Success)
			if sectionPercent < 50 {
				sectionBarStyle = sectionBarStyle.Foreground(colors.Warning)
			}

			sectionContent := fmt.Sprintf(
				"%s\n%d/%d tasks  %.1f%%\n%s",
				lipgloss.NewStyle().Foreground(colors.Section).Bold(true).Render(sectionName),
				stats.completed, stats.total,
				sectionPercent,
				sectionBarStyle.Render(sectionBar),
			)

			sectionCards = append(sectionCards, sectionCardStyle.Render(sectionContent))
		}
		
		// Render cards in 2 columns
		for i := 0; i < len(sectionCards); i += 2 {
			if i+1 < len(sectionCards) {
				s += lipgloss.JoinHorizontal(lipgloss.Top, sectionCards[i], "  ", sectionCards[i+1]) + "\n"
			} else {
				s += sectionCards[i] + "\n"
			}
		}
		s += "\n"
	}

	// Hint
	hintStyle := lipgloss.NewStyle().
		Foreground(colors.Hint).
		Italic(true)
	s += hintStyle.Render("esc back")

	return s
}
