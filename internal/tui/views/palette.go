package views

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/i-am-fran/markdowndo/internal/tui/colors"
)

// PaletteAction represents a command palette action
type PaletteAction string

const (
	PaletteActionAddTask    PaletteAction = "add"
	PaletteActionListTasks  PaletteAction = "list"
	PaletteActionSearch     PaletteAction = "search"
	PaletteActionSettings   PaletteAction = "settings"
	PaletteActionSubfolders PaletteAction = "subfolders"
	PaletteActionLint       PaletteAction = "lint"
	PaletteActionQuit       PaletteAction = "quit"
)

// paletteItem represents an item in the command palette
type paletteItem struct {
	title       string
	description string
	action      PaletteAction
}

func (i paletteItem) Title() string       { return i.title }
func (i paletteItem) Description() string { return i.description }
func (i paletteItem) FilterValue() string { return i.title + " " + i.description }

// PaletteModel is the command palette model
type PaletteModel struct {
	list   list.Model
	width  int
	height int
}

// NewPaletteModel creates a new palette model
func NewPaletteModel(width, height int) PaletteModel {
	items := []list.Item{
		paletteItem{title: "Add Task", description: "Create a new task", action: PaletteActionAddTask},
		paletteItem{title: "List Tasks", description: "View all tasks", action: PaletteActionListTasks},
		paletteItem{title: "Search Tasks", description: "Find tasks by keyword", action: PaletteActionSearch},
		paletteItem{title: "Browse Subfolders", description: "View TODO files in subdirectories", action: PaletteActionSubfolders},
		paletteItem{title: "Settings", description: "Configure preferences", action: PaletteActionSettings},
		paletteItem{title: "Lint File", description: "Check and fix formatting", action: PaletteActionLint},
		paletteItem{title: "Quit", description: "Exit application", action: PaletteActionQuit},
	}

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.SetSpacing(0)
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Foreground(colors.Selected).Bold(true)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Foreground(colors.Hint)

	// Create a smaller list for the palette overlay
	paletteWidth := 60
	if width < 70 {
		paletteWidth = width - 10
	}
	paletteHeight := 15
	if height < 20 {
		paletteHeight = height - 5
	}

	l := list.New(items, delegate, paletteWidth, paletteHeight)
	l.Title = "Quick Actions"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true) // Enable fuzzy search
	l.SetShowHelp(false)
	l.Styles.Title = lipgloss.NewStyle().Bold(true).Foreground(colors.HeaderFG).Background(colors.HeaderBG).Padding(0, 1)

	return PaletteModel{
		list:   l,
		width:  width,
		height: height,
	}
}

// Init implements tea.Model
func (m PaletteModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m PaletteModel) Update(msg tea.Msg) (PaletteModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
		paletteWidth := 60
		if msg.Width < 70 {
			paletteWidth = msg.Width - 10
		}
		paletteHeight := 15
		if msg.Height < 20 {
			paletteHeight = msg.Height - 5
		}
		
		m.list.SetSize(paletteWidth, paletteHeight)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if item, ok := m.list.SelectedItem().(paletteItem); ok {
				return m, func() tea.Msg {
					return PaletteSelectMsg{Action: item.action}
				}
			}
		case "esc", "ctrl+k":
			return m, func() tea.Msg { return ClosePaletteMsg{} }
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View implements tea.Model
func (m PaletteModel) View() string {
	// Create overlay background
	paletteView := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colors.Border).
		Padding(1, 2).
		Render(m.list.View())

	// Center the palette
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		paletteView,
	)
}

// PaletteSelectMsg is sent when a palette action is selected
type PaletteSelectMsg struct {
	Action PaletteAction
}

// ClosePaletteMsg is sent to close the palette
type ClosePaletteMsg struct{}

// OpenPaletteMsg is sent to open the palette
type OpenPaletteMsg struct{}
