package config

import (
	"io"
	"os"
	"strings"
	"testing"
)

// withTempConfig points configFile at a fresh temp file for the duration of
// the test and clears the in-memory cache before and after.
func withTempConfig(t *testing.T, content string) {
	t.Helper()
	restore := SetConfigDirForTesting(t.TempDir())
	if content != "" {
		if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(restore)
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

// TestGetSettingsWarnsOnMalformedJSON pins the fix for a corrupt
// config.json being silently swallowed into defaults with no diagnostic —
// a hand-corrupted or partially-written file used to look indistinguishable
// from "no config file at all".
func TestGetSettingsWarnsOnMalformedJSON(t *testing.T) {
	withTempConfig(t, `{not valid json`)

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	origStderr := os.Stderr
	os.Stderr = w
	s := GetSettings()
	w.Close()
	os.Stderr = origStderr

	captured, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read captured stderr: %v", err)
	}

	if !strings.Contains(string(captured), "could not parse") {
		t.Errorf("expected a warning about the malformed config, got: %q", captured)
	}
	want := DefaultSettings()
	if s.ShowCompleted != want.ShowCompleted || s.Editor != want.Editor ||
		s.EnableInProgress != want.EnableInProgress || s.ConfirmDestructive != want.ConfirmDestructive {
		t.Errorf("expected defaults returned for malformed config, got %+v", s)
	}
}
