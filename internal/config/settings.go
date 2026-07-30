package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/i-am-fran/markdown-do/v3/internal/fsutil"
)

var (
	configDir  string
	configFile string
	cached     *Settings
)

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not determine home directory (%v); storing mdd config under the current directory\n", err)
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
		unmarshalErr := json.Unmarshal(data, &loaded)
		presentErr := json.Unmarshal(data, &present)
		if unmarshalErr != nil || presentErr != nil {
			firstErr := unmarshalErr
			if firstErr == nil {
				firstErr = presentErr
			}
			fmt.Fprintf(os.Stderr, "Warning: could not parse %s (%v) — using default settings\n", configFile, firstErr)
		} else {
			// Only override a default when the key is actually present in the
			// file — a bool field silently missing from an older config.json
			// must not be read as an explicit `false`.
			if _, ok := present["showCompleted"]; ok {
				settings.ShowCompleted = loaded.ShowCompleted
			}
			if loaded.Editor != "" {
				settings.Editor = loaded.Editor
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

	lock, err := fsutil.AcquireLock(configFile)
	if err != nil {
		return fmt.Errorf("could not lock %s for writing: %w", configFile, err)
	}
	defer func() { _ = lock.Unlock() }()

	if err := fsutil.AtomicWriteFile(configFile, data, 0644); err != nil {
		return err
	}

	cached = &settings
	return nil
}

// UpdateSettings updates specific settings
func UpdateSettings(updates map[string]interface{}) error {
	current := GetSettings()

	if v, ok := updates["showCompleted"].(bool); ok {
		current.ShowCompleted = v
	}
	if v, ok := updates["editor"].(EditorOption); ok {
		current.Editor = v
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

// SetConfigDirForTesting overrides the directory used for config.json and
// returns a function that restores the previous value. Intended for tests
// only, so package-level tests never read or write the real
// ~/.config/markdowndo on the machine running `go test`.
func SetConfigDirForTesting(dir string) (restore func()) {
	origDir, origFile := configDir, configFile
	configDir = dir
	configFile = filepath.Join(dir, "config.json")
	cached = nil
	return func() {
		configDir, configFile = origDir, origFile
		cached = nil
	}
}
