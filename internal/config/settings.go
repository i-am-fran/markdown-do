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

// GetSettings returns the current settings (cached)
func GetSettings() Settings {
	if cached != nil {
		return *cached
	}

	settings := DefaultSettings()

	data, err := os.ReadFile(configFile)
	if err == nil {
		var loaded Settings
		if json.Unmarshal(data, &loaded) == nil {
			// Merge with defaults
			if loaded.Theme != "" {
				settings.Theme = loaded.Theme
			}
			settings.Fullscreen = loaded.Fullscreen
			settings.ShowCompleted = loaded.ShowCompleted
			if loaded.Editor != "" {
				settings.Editor = loaded.Editor
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
