package cli

import (
	"os"
	"testing"

	"github.com/i-am-fran/markdown-do/internal/config"
	"github.com/i-am-fran/markdown-do/internal/core"
)

// TestMain points core's cache/undo state and config's settings.json at
// scratch directories for the whole package's test run, so tests that call
// Save() (directly or via Perform*/CompleteTask/etc.) or read settings
// (e.g. EnableInProgress) never touch the real ~/.config/markdowndo on the
// machine running `go test`.
func TestMain(m *testing.M) {
	stateDir, err := os.MkdirTemp("", "mdd-test-state-*")
	if err != nil {
		panic(err)
	}
	restoreState := core.SetStateDirForTesting(stateDir)

	configDir, err := os.MkdirTemp("", "mdd-test-config-*")
	if err != nil {
		panic(err)
	}
	restoreConfig := config.SetConfigDirForTesting(configDir)

	code := m.Run()

	restoreConfig()
	restoreState()
	_ = os.RemoveAll(stateDir)
	_ = os.RemoveAll(configDir)
	os.Exit(code)
}
