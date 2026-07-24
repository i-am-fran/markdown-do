package views

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/i-am-fran/markdowndo/internal/config"
	"github.com/i-am-fran/markdowndo/internal/tui/colors"
)

type settingsItem struct {
	title  string
	action string
}

func (i settingsItem) Title() string       { return i.title }
func (i settingsItem) Description() string { return "" }
func (i settingsItem) FilterValue() string { return i.title }

// SettingsModel is the settings view model
type SettingsModel struct {
	list       list.Model
	editorList list.Model
	settings   config.Settings
	viewMode   string // "main", "editor", "confirmReset"
	width      int
	height     int
}

// NewSettingsModel creates a new settings model
func NewSettingsModel(width, height int) SettingsModel {
	settings := config.GetSettings()

	items := buildSettingsItems(settings)

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.SetSpacing(0)

	l := list.New(items, delegate, width, height-6)
	l.Title = "Settings"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(colors.Accent).
		Bold(true).
		MarginBottom(1).
		MarginTop(1)
	l.Styles.NoItems = lipgloss.NewStyle().
		Foreground(colors.Muted).
		Italic(true)

	// Editor list
	editorItems := buildEditorItems(settings)
	el := list.New(editorItems, delegate, width, height-6)
	el.Title = "Select editor:"
	el.SetShowStatusBar(false)
	el.SetFilteringEnabled(false)
	el.SetShowHelp(false)
	el.Styles.Title = lipgloss.NewStyle().
		Foreground(colors.Accent).
		Bold(true).
		MarginBottom(1).
		MarginTop(1)
	el.Styles.NoItems = lipgloss.NewStyle().
		Foreground(colors.Muted).
		Italic(true)

	return SettingsModel{
		list:       l,
		editorList: el,
		settings:   settings,
		viewMode:   "main",
		width:      width,
		height:     height,
	}
}

func buildSettingsItems(settings config.Settings) []list.Item {
	editorLabel := "System default"
	for _, opt := range config.EditorOptions() {
		if opt.Value == settings.Editor {
			editorLabel = opt.Label
			break
		}
	}

	fullscreenLabel := "off"
	if settings.Fullscreen {
		fullscreenLabel = "on"
	}

	showCompletedLabel := "off"
	if settings.ShowCompleted {
		showCompletedLabel = "on"
	}

	enableTUILabel := "off"
	if settings.EnableTUI {
		enableTUILabel = "on"
	}

	enableInProgressLabel := "off"
	if settings.EnableInProgress {
		enableInProgressLabel = "on"
	}

	confirmDestructiveLabel := "off"
	if settings.ConfirmDestructive {
		confirmDestructiveLabel = "on"
	}

	return []list.Item{
		settingsItem{title: fmt.Sprintf("Editor: %s", editorLabel), action: "editor"},
		settingsItem{title: fmt.Sprintf("Fullscreen mode: %s", fullscreenLabel), action: "fullscreen"},
		settingsItem{title: fmt.Sprintf("Show completed tasks: %s", showCompletedLabel), action: "showCompleted"},
		settingsItem{title: fmt.Sprintf("Enable TUI: %s", enableTUILabel), action: "enableTUI"},
		settingsItem{title: fmt.Sprintf("Enable in-progress status: %s", enableInProgressLabel), action: "enableInProgress"},
		settingsItem{title: fmt.Sprintf("Confirm destructive actions: %s", confirmDestructiveLabel), action: "confirmDestructive"},
		settingsItem{title: "Reset to defaults", action: "reset"},
		settingsItem{title: "Back", action: "back"},
	}
}

func buildEditorItems(settings config.Settings) []list.Item {
	var items []list.Item
	for _, opt := range config.EditorOptions() {
		label := opt.Label
		if opt.Value == settings.Editor {
			label += " (current)"
		}
		items = append(items, settingsItem{title: label, action: string(opt.Value)})
	}
	return items
}

// Init implements tea.Model
func (m SettingsModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m SettingsModel) Update(msg tea.Msg) (SettingsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-6)
		m.editorList.SetSize(msg.Width, msg.Height-6)
		return m, nil

	case tea.KeyMsg:
		if m.viewMode == "confirmReset" {
			switch msg.String() {
			case "y", "Y":
				config.ResetSettings()
				m.settings = config.GetSettings()
				m.list.SetItems(buildSettingsItems(m.settings))
				m.editorList.SetItems(buildEditorItems(m.settings))
				m.viewMode = "main"
				return m, func() tea.Msg { return SettingsChangedMsg{} }
			case "n", "N", "esc":
				m.viewMode = "main"
				return m, nil
			}
			return m, nil
		}

		if m.viewMode == "editor" {
			switch msg.String() {
			case "enter":
				if item, ok := m.editorList.SelectedItem().(settingsItem); ok {
					config.UpdateSettings(map[string]interface{}{
						"editor": config.EditorOption(item.action),
					})
					m.settings = config.GetSettings()
					m.list.SetItems(buildSettingsItems(m.settings))
					m.editorList.SetItems(buildEditorItems(m.settings))
					m.viewMode = "main"
					return m, func() tea.Msg { return SettingsChangedMsg{} }
				}
			case "esc":
				m.viewMode = "main"
				return m, nil
			}

			var cmd tea.Cmd
			m.editorList, cmd = m.editorList.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "enter":
			if item, ok := m.list.SelectedItem().(settingsItem); ok {
				switch item.action {
				case "editor":
					m.viewMode = "editor"
					return m, nil
				case "fullscreen":
					config.UpdateSettings(map[string]interface{}{
						"fullscreen": !m.settings.Fullscreen,
					})
					m.settings = config.GetSettings()
					m.list.SetItems(buildSettingsItems(m.settings))
					return m, func() tea.Msg { return SettingsChangedMsg{} }
				case "showCompleted":
					config.UpdateSettings(map[string]interface{}{
						"showCompleted": !m.settings.ShowCompleted,
					})
					m.settings = config.GetSettings()
					m.list.SetItems(buildSettingsItems(m.settings))
					return m, func() tea.Msg { return SettingsChangedMsg{} }
				case "enableTUI":
					config.UpdateSettings(map[string]interface{}{
						"enableTUI": !m.settings.EnableTUI,
					})
					m.settings = config.GetSettings()
					m.list.SetItems(buildSettingsItems(m.settings))
					return m, func() tea.Msg { return SettingsChangedMsg{} }
				case "enableInProgress":
					config.UpdateSettings(map[string]interface{}{
						"enableInProgress": !m.settings.EnableInProgress,
					})
					m.settings = config.GetSettings()
					m.list.SetItems(buildSettingsItems(m.settings))
					return m, func() tea.Msg { return SettingsChangedMsg{} }
				case "confirmDestructive":
					config.UpdateSettings(map[string]interface{}{
						"confirmDestructive": !m.settings.ConfirmDestructive,
					})
					m.settings = config.GetSettings()
					m.list.SetItems(buildSettingsItems(m.settings))
					return m, func() tea.Msg { return SettingsChangedMsg{} }
				case "reset":
					m.viewMode = "confirmReset"
					return m, nil
				case "back":
					return m, func() tea.Msg { return BackMsg{} }
				}
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
func (m SettingsModel) View() string {
	if m.viewMode == "confirmReset" {
		s := lipgloss.NewStyle().Bold(true).Render("Reset all settings to defaults?") + "\n\n"
		s += lipgloss.NewStyle().Foreground(colors.Hint).Render("y yes  n no")
		return s
	}

	if m.viewMode == "editor" {
		hint := "enter select  esc back"
		return m.editorList.View() + "\n\n" + lipgloss.NewStyle().Foreground(colors.Hint).Render(hint)
	}

	hint := "enter toggle/select  esc back"
	return m.list.View() + "\n\n" + lipgloss.NewStyle().Foreground(colors.Hint).Render(hint)
}

// SettingsChangedMsg is sent when settings change
type SettingsChangedMsg struct{}
