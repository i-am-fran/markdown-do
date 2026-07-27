package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/i-am-fran/markdown-do/internal/core"
)

// chdirToTempTodoFile creates a TODO.md with the given content in a fresh
// temp dir, chdirs into it for the duration of the test, and returns its
// path.
func chdirToTempTodoFile(t *testing.T, content string) string {
	t.Helper()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("EvalSymlinks failed: %v", err)
	}
	path := filepath.Join(dir, "TODO.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write TODO.md: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}
	t.Cleanup(func() { os.Chdir(origWd) })
	return path
}

// A multi-ID `mdd complete 1 2` must complete the tasks originally at
// positions 1 and 2, not whatever ends up at those positions after Save()
// reorders the file mid-command (MDD-034).
func TestCompleteTaskMultiIDCompletesOriginalPositions(t *testing.T) {
	path := chdirToTempTodoFile(t, "# TODO\n\n- [ ] Task A\n- [ ] Task B\n- [ ] Task C\n")

	if err := CompleteTask([]string{"1", "2"}); err != nil {
		t.Fatalf("CompleteTask failed: %v", err)
	}

	todoFile, err := core.Load(path)
	if err != nil {
		t.Fatalf("failed to reload TODO.md: %v", err)
	}

	statusOf := func(text string) core.TaskStatus {
		for _, task := range todoFile.GetTasks() {
			if task.Text == text {
				return task.Status
			}
		}
		t.Fatalf("task %q not found after CompleteTask", text)
		return ""
	}

	if statusOf("Task A") != core.TaskCompleted {
		t.Errorf("expected Task A completed, got %v", statusOf("Task A"))
	}
	if statusOf("Task B") != core.TaskCompleted {
		t.Errorf("expected Task B completed, got %v", statusOf("Task B"))
	}
	if statusOf("Task C") != core.TaskPending {
		t.Errorf("expected Task C still pending, got %v", statusOf("Task C"))
	}
}

// `mdd add --path <dir>` must add the task to <dir>'s TODO.md, not the cwd's
// (MDD-038).
func TestAddTaskWithPathTargetsOtherDirectory(t *testing.T) {
	cwdPath := chdirToTempTodoFile(t, "# TODO\n\n")

	otherDir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("EvalSymlinks failed: %v", err)
	}

	if err := AddTask("Buy milk", otherDir); err != nil {
		t.Fatalf("AddTask failed: %v", err)
	}

	otherTodoFile, err := core.Load(filepath.Join(otherDir, "TODO.md"))
	if err != nil {
		t.Fatalf("failed to load TODO.md from target dir: %v", err)
	}
	tasks := otherTodoFile.GetTasks()
	if len(tasks) != 1 || tasks[0].Text != "Buy milk" {
		t.Fatalf("expected 1 task %q in target dir, got %v", "Buy milk", tasks)
	}

	cwdTodoFile, err := core.Load(cwdPath)
	if err != nil {
		t.Fatalf("failed to reload cwd TODO.md: %v", err)
	}
	if len(cwdTodoFile.GetTasks()) != 0 {
		t.Errorf("expected cwd TODO.md untouched, got %v", cwdTodoFile.GetTasks())
	}
}

func TestAddTaskWithPathRejectsMissingDirectory(t *testing.T) {
	chdirToTempTodoFile(t, "# TODO\n\n")

	err := AddTask("Buy milk", filepath.Join(t.TempDir(), "does-not-exist"))
	if err == nil {
		t.Fatal("expected error for nonexistent --path directory, got nil")
	}
}
