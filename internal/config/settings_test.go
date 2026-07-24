package config

import (
	"os"
	"path/filepath"
	"testing"
)

// withTempConfig points configFile at a fresh temp file for the duration of
// the test and clears the in-memory cache before and after.
func withTempConfig(t *testing.T, content string) {
	t.Helper()
	dir := t.TempDir()
	origDir, origFile := configDir, configFile
	configDir = dir
	configFile = filepath.Join(dir, "config.json")
	if content != "" {
		if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	ClearCache()
	t.Cleanup(func() {
		configDir, configFile = origDir, origFile
		ClearCache()
	})
}

func TestGetSettingsFillsMissingKeysWithDefaults(t *testing.T) {
	// Simulates an older config.json written before enableTUI/enableAnimations/
	// showStatusBar existed: those keys are simply absent from the file.
	withTempConfig(t, `{"theme":"default","fullscreen":true,"showCompleted":true,"editor":"vim"}`)

	s := GetSettings()
	if !s.EnableTUI {
		t.Error("expected EnableTUI to default to true when absent from config.json")
	}
	if !s.EnableAnimations {
		t.Error("expected EnableAnimations to default to true when absent from config.json")
	}
	if !s.ShowStatusBar {
		t.Error("expected ShowStatusBar to default to true when absent from config.json")
	}
	if s.Editor != EditorVim {
		t.Errorf("expected loaded editor to still apply, got %v", s.Editor)
	}
}

func TestGetSettingsRespectsExplicitFalse(t *testing.T) {
	withTempConfig(t, `{"enableTUI":false,"enableAnimations":false,"showStatusBar":false}`)

	s := GetSettings()
	if s.EnableTUI {
		t.Error("expected explicit EnableTUI:false to be respected")
	}
	if s.EnableAnimations {
		t.Error("expected explicit EnableAnimations:false to be respected")
	}
	if s.ShowStatusBar {
		t.Error("expected explicit ShowStatusBar:false to be respected")
	}
}
