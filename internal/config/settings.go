package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

var (
	configDir  string
	configFile string
	cached     *Settings
)

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	configDir = filepath.Join(homeDir, ".config", "markdowndo")
	configFile = filepath.Join(configDir, "config.json")
}

func ensureConfigDir() error {
	return os.MkdirAll(configDir, 0755)
}

// ConfigFilePath returns the path to the settings file
// (~/.config/markdowndo/config.json), creating it with the current settings
// first if it doesn't exist yet.
func ConfigFilePath() (string, error) {
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		if err := SaveSettings(GetSettings()); err != nil {
			return "", err
		}
	}
	return configFile, nil
}

// GetSettings returns the current settings (cached)
func GetSettings() Settings {
	if cached != nil {
		return *cached
	}

	settings := DefaultSettings()

	data, err := os.ReadFile(configFile)
	if err == nil {
		var loaded Settings
		var present map[string]interface{}
		if json.Unmarshal(data, &loaded) == nil && json.Unmarshal(data, &present) == nil {
			// Only override a default when the key is actually present in the
			// file — a bool field silently missing from an older config.json
			// must not be read as an explicit `false`.
			if loaded.Theme != "" {
				settings.Theme = loaded.Theme
			}
			if _, ok := present["fullscreen"]; ok {
				settings.Fullscreen = loaded.Fullscreen
			}
			if _, ok := present["showCompleted"]; ok {
				settings.ShowCompleted = loaded.ShowCompleted
			}
			if loaded.Editor != "" {
				settings.Editor = loaded.Editor
			}
			if _, ok := present["showStatusBar"]; ok {
				settings.ShowStatusBar = loaded.ShowStatusBar
			}
			if _, ok := present["enableAnimations"]; ok {
				settings.EnableAnimations = loaded.EnableAnimations
			}
			if _, ok := present["enableTUI"]; ok {
				settings.EnableTUI = loaded.EnableTUI
			}
			if _, ok := present["enableInProgress"]; ok {
				settings.EnableInProgress = loaded.EnableInProgress
			}
			if _, ok := present["confirmDestructive"]; ok {
				settings.ConfirmDestructive = loaded.ConfirmDestructive
			}
			if loaded.SectionAliases != nil {
				settings.SectionAliases = loaded.SectionAliases
			}
		}
	}

	cached = &settings
	return settings
}

// SaveSettings saves settings to disk
func SaveSettings(settings Settings) error {
	if err := ensureConfigDir(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return err
	}

	cached = &settings
	return nil
}

// UpdateSettings updates specific settings
func UpdateSettings(updates map[string]interface{}) error {
	current := GetSettings()

	if v, ok := updates["fullscreen"].(bool); ok {
		current.Fullscreen = v
	}
	if v, ok := updates["showCompleted"].(bool); ok {
		current.ShowCompleted = v
	}
	if v, ok := updates["editor"].(EditorOption); ok {
		current.Editor = v
	}
	if v, ok := updates["theme"].(string); ok {
		current.Theme = v
	}
	if v, ok := updates["showStatusBar"].(bool); ok {
		current.ShowStatusBar = v
	}
	if v, ok := updates["enableAnimations"].(bool); ok {
		current.EnableAnimations = v
	}
	if v, ok := updates["enableTUI"].(bool); ok {
		current.EnableTUI = v
	}
	if v, ok := updates["enableInProgress"].(bool); ok {
		current.EnableInProgress = v
	}
	if v, ok := updates["confirmDestructive"].(bool); ok {
		current.ConfirmDestructive = v
	}
	if v, ok := updates["sectionAlias"].([2]string); ok {
		if current.SectionAliases == nil {
			current.SectionAliases = map[string]string{}
		}
		current.SectionAliases[v[0]] = v[1]
	}

	return SaveSettings(current)
}

// ResetSettings resets all settings to defaults
func ResetSettings() error {
	return SaveSettings(DefaultSettings())
}

// ClearCache clears the settings cache (useful for testing)
func ClearCache() {
	cached = nil
}
