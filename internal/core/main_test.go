package core

import (
	"os"
	"testing"
)

// TestMain points stateDir at a scratch directory for the whole package's
// test run, so any test that calls Save() (which writes an undo backup)
// or the cache functions never reads or writes the real
// ~/.config/markdowndo on the machine running `go test`.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "mdd-test-state-*")
	if err != nil {
		panic(err)
	}
	restore := SetStateDirForTesting(dir)

	code := m.Run()

	restore()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}
