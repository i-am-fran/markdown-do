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
	// Simulates an older config.json written before enableInProgress/
	// confirmDestructive existed: those keys are simply absent from the file.
	withTempConfig(t, `{"showCompleted":true,"editor":"vim"}`)

	s := GetSettings()
	if s.EnableInProgress {
		t.Error("expected EnableInProgress to default to false when absent from config.json")
	}
	if s.ConfirmDestructive {
		t.Error("expected ConfirmDestructive to default to false when absent from config.json")
	}
	if s.Editor != EditorVim {
		t.Errorf("expected loaded editor to still apply, got %v", s.Editor)
	}
}

func TestGetSettingsRespectsExplicitFalse(t *testing.T) {
	withTempConfig(t, `{"showCompleted":false,"enableInProgress":true}`)

	s := GetSettings()
	if s.ShowCompleted {
		t.Error("expected explicit showCompleted:false to be respected")
	}
	if !s.EnableInProgress {
		t.Error("expected explicit enableInProgress:true to be respected")
	}
}
