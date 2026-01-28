package views

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/i-am-fran/markdowndo/internal/core"
)

type folderItem struct {
	dir      string
	filePath string
	count    int
}

func (i folderItem) Title() string {
	suffix := "s"
	if i.count == 1 {
		suffix = ""
	}
	return fmt.Sprintf("%s (%d file%s)", i.dir, i.count, suffix)
}
func (i folderItem) Description() string { return "" }
func (i folderItem) FilterValue() string { return i.dir }

// SubfoldersModel is the subfolders view model
type SubfoldersModel struct {
	list         list.Model
	taskList     TaskListModel
	folders      []folderItem
	selectedFile string
	todoFile     *core.TodoFile
	viewMode     string // "folders" or "tasks"
	loading      bool
	width        int
	height       int
}

// NewSubfoldersModel creates a new subfolders model
func NewSubfoldersModel(width, height int) SubfoldersModel {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.SetSpacing(0)

	l := list.New([]list.Item{}, delegate, width, height-6)
	l.Title = "Select a folder:"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.Styles.Title = lipgloss.NewStyle().Bold(true).MarginBottom(1)

	return SubfoldersModel{
		list:     l,
		viewMode: "folders",
		loading:  true,
		width:    width,
		height:   height,
	}
}

// Init implements tea.Model
func (m SubfoldersModel) Init() tea.Cmd {
	return m.loadFolders()
}

func (m SubfoldersModel) loadFolders() tea.Cmd {
	return func() tea.Msg {
		cwd, _ := os.Getwd()
		filesByDir, err := core.FindTodoFilesInSubdirs(cwd)
		if err != nil {
			return subfoldersLoadedMsg{folders: nil}
		}

		var folders []folderItem
		for dir, files := range filesByDir {
			if len(files) > 0 {
				folders = append(folders, folderItem{
					dir:      dir,
					filePath: files[0].Path,
					count:    len(files),
				})
			}
		}

		return subfoldersLoadedMsg{folders: folders}
	}
}

// Update implements tea.Model
func (m SubfoldersModel) Update(msg tea.Msg) (SubfoldersModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-6)
		return m, nil

	case subfoldersLoadedMsg:
		m.loading = false
		m.folders = msg.folders

		var items []list.Item
		for _, folder := range m.folders {
			items = append(items, folder)
		}
		items = append(items, backItem{})
		m.list.SetItems(items)
		return m, nil

	case tea.KeyMsg:
		if m.viewMode == "tasks" {
			switch msg.String() {
			case "esc":
				m.viewMode = "folders"
				m.todoFile = nil
				return m, nil
			}

			var cmd tea.Cmd
			m.taskList, cmd = m.taskList.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "enter":
			if item, ok := m.list.SelectedItem().(folderItem); ok {
				todoFile, err := core.Load(item.filePath)
				if err == nil {
					m.selectedFile = item.filePath
					m.todoFile = todoFile
					m.taskList = NewTaskListModel(todoFile, m.width, m.height)
					m.viewMode = "tasks"
				}
				return m, nil
			}
			if _, ok := m.list.SelectedItem().(backItem); ok {
				return m, func() tea.Msg { return BackMsg{} }
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
func (m SubfoldersModel) View() string {
	if m.loading {
		return "Loading subfolders..."
	}

	if m.viewMode == "tasks" && m.todoFile != nil {
		return m.taskList.View()
	}

	if len(m.folders) == 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#B8860B", Dark: "#FFD700"}).Render("No TODO files found in subfolders") + "\n\n" +
			lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#5C5C5C"}).Render("esc back")
	}

	hint := "↑↓ navigate • enter select • esc back"
	return m.list.View() + "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#5C5C5C"}).Render(hint)
}

type subfoldersLoadedMsg struct {
	folders []folderItem
}
