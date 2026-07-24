package config

// EditorOption represents the editor preference
type EditorOption string

const (
	EditorSystem     EditorOption = "system"
	EditorVim        EditorOption = "vim"
	EditorNano       EditorOption = "nano"
	EditorDefaultApp EditorOption = "default-app"
)

// Settings represents user configuration
type Settings struct {
	Theme              string            `json:"theme"`
	Fullscreen         bool              `json:"fullscreen"`
	ShowCompleted      bool              `json:"showCompleted"`
	Editor             EditorOption      `json:"editor"`
	ShowStatusBar      bool              `json:"showStatusBar"`
	EnableAnimations   bool              `json:"enableAnimations"`
	EnableTUI          bool              `json:"enableTUI"`
	EnableInProgress   bool              `json:"enableInProgress"`
	ConfirmDestructive bool              `json:"confirmDestructive"`
	SectionAliases     map[string]string `json:"sectionAliases"`
}

// DefaultSettings returns the default settings
func DefaultSettings() Settings {
	return Settings{
		Theme:              "default",
		Fullscreen:         true,
		ShowCompleted:      true,
		Editor:             EditorSystem,
		ShowStatusBar:      true,
		EnableAnimations:   true,
		EnableTUI:          true,
		EnableInProgress:   false,
		ConfirmDestructive: false,
		SectionAliases:     map[string]string{},
	}
}

// EditorOptionInfo contains display info for an editor option
type EditorOptionInfo struct {
	Value EditorOption
	Label string
	Hint  string
}

// EditorOptions returns all available editor options
func EditorOptions() []EditorOptionInfo {
	return []EditorOptionInfo{
		{Value: EditorSystem, Label: "System default", Hint: "Uses $EDITOR or $VISUAL"},
		{Value: EditorVim, Label: "Vim", Hint: "Open in Vim"},
		{Value: EditorNano, Label: "Nano", Hint: "Open in Nano"},
		{Value: EditorDefaultApp, Label: "Default app", Hint: "Open with system default (Preview, etc.)"},
	}
}
