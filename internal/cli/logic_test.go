package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/i-am-fran/markdown-do/internal/core"
)

func newTestTodoFile(t *testing.T, content string) *core.TodoFile {
	t.Helper()
	path := filepath.Join(t.TempDir(), "TODO.md")
	return core.NewTodoFile(path, content)
}

func TestLoadDefaultTodoFile(t *testing.T) {
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("EvalSymlinks failed: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}
	defer os.Chdir(origWd)

	todoFile, filePath, err := LoadDefaultTodoFile()
	if err != nil {
		t.Fatalf("LoadDefaultTodoFile failed: %v", err)
	}
	if filePath != filepath.Join(dir, "TODO.md") {
		t.Errorf("expected default path %q, got %q", filepath.Join(dir, "TODO.md"), filePath)
	}
	if len(todoFile.GetTasks()) != 0 {
		t.Errorf("expected no tasks for a nonexistent file, got %d", len(todoFile.GetTasks()))
	}
}

func TestPerformAddTaskCreatesFileOnFirstUse(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "TODO.md")
	todoFile := core.NewTodoFile(filePath, "")

	updated, task, err := PerformAddTask(todoFile, filePath, "Buy milk")
	if err != nil {
		t.Fatalf("PerformAddTask failed: %v", err)
	}
	if task.Text != "Buy milk" {
		t.Errorf("expected task text %q, got %q", "Buy milk", task.Text)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("expected file to be created on disk: %v", err)
	}
	if len(updated.GetTasks()) != 1 {
		t.Errorf("expected 1 task, got %d", len(updated.GetTasks()))
	}
	if !strings.Contains(string(content), "# TODO") {
		t.Errorf("expected new file to contain default template, got: %q", content)
	}
}

func TestPerformAddTaskRejectsShortText(t *testing.T) {
	todoFile := newTestTodoFile(t, "# TODO\n\n")
	if _, _, err := PerformAddTask(todoFile, todoFile.FilePath, "ab"); err == nil {
		t.Error("expected error for too-short task text, got nil")
	}
}

func TestPerformAddNote(t *testing.T) {
	todoFile := newTestTodoFile(t, "# TODO\n\n- [ ] Buy milk\n")
	updated, err := PerformAddNote(todoFile, todoFile.FilePath, "Remember to review the PR")
	if err != nil {
		t.Fatalf("PerformAddNote failed: %v", err)
	}
	found := false
	for _, s := range updated.GetSectionNames() {
		if s == "Notes" {
			found = true
		}
	}
	if !found {
		t.Error("expected a Notes section to be created")
	}
}

func TestPerformToggleTask(t *testing.T) {
	todoFile := newTestTodoFile(t, "# TODO\n\n- [ ] Buy milk\n")
	task, err := PerformToggleTask(todoFile, 1)
	if err != nil {
		t.Fatalf("PerformToggleTask failed: %v", err)
	}
	if task.Status != core.TaskCompleted {
		t.Errorf("expected task to be completed, got %v", task.Status)
	}
}

func TestPerformToggleTaskNotFound(t *testing.T) {
	todoFile := newTestTodoFile(t, "# TODO\n\n- [ ] Buy milk\n")
	if _, err := PerformToggleTask(todoFile, 99); err == nil {
		t.Error("expected error for nonexistent task, got nil")
	}
}

func TestPerformDeleteTask(t *testing.T) {
	todoFile := newTestTodoFile(t, "# TODO\n\n- [ ] Buy milk\n")
	deleted, err := PerformDeleteTask(todoFile, 1)
	if err != nil {
		t.Fatalf("PerformDeleteTask failed: %v", err)
	}
	if deleted.Text != "Buy milk" {
		t.Errorf("expected deleted task text %q, got %q", "Buy milk", deleted.Text)
	}
	if len(todoFile.GetTasks()) != 0 {
		t.Errorf("expected task to be removed, got %d tasks", len(todoFile.GetTasks()))
	}
}

func TestPerformDeleteTaskNotFound(t *testing.T) {
	todoFile := newTestTodoFile(t, "# TODO\n\n- [ ] Buy milk\n")
	if _, err := PerformDeleteTask(todoFile, 99); err == nil {
		t.Error("expected error for nonexistent task, got nil")
	}
}

func TestPerformMoveTask(t *testing.T) {
	todoFile := newTestTodoFile(t, "# TODO\n\n- [ ] Buy milk\n")
	task, err := PerformMoveTask(todoFile, 1, stringPtr("Errands"))
	if err != nil {
		t.Fatalf("PerformMoveTask failed: %v", err)
	}
	if task.Section == nil || *task.Section != "Errands" {
		t.Errorf("expected task moved to section %q, got %v", "Errands", task.Section)
	}

	found := false
	for _, s := range todoFile.GetSectionNames() {
		if s == "Errands" {
			found = true
		}
	}
	if !found {
		t.Error("expected new section \"Errands\" to be created")
	}
}

func TestPerformMoveTaskToInbox(t *testing.T) {
	todoFile := newTestTodoFile(t, "# TODO\n\n## Errands\n\n- [ ] Buy milk\n")
	task, err := PerformMoveTask(todoFile, 1, nil)
	if err != nil {
		t.Fatalf("PerformMoveTask failed: %v", err)
	}
	if task.Section != nil {
		t.Errorf("expected task moved to inbox (nil section), got %v", *task.Section)
	}
}

func TestPerformMoveTaskNotFound(t *testing.T) {
	todoFile := newTestTodoFile(t, "# TODO\n\n- [ ] Buy milk\n")
	if _, err := PerformMoveTask(todoFile, 99, nil); err == nil {
		t.Error("expected error for nonexistent task, got nil")
	}
}

func TestPerformRestore(t *testing.T) {
	todoFile := newTestTodoFile(t, "# TODO\n\n- [ ] Buy milk\n")
	snapshot := todoFile.Snapshot()

	if _, err := PerformDeleteTask(todoFile, 1); err != nil {
		t.Fatalf("PerformDeleteTask failed: %v", err)
	}
	if len(todoFile.GetTasks()) != 0 {
		t.Fatalf("expected task deleted before restore, got %d tasks", len(todoFile.GetTasks()))
	}

	if err := PerformRestore(todoFile, snapshot); err != nil {
		t.Fatalf("PerformRestore failed: %v", err)
	}
	tasks := todoFile.GetTasks()
	if len(tasks) != 1 || tasks[0].Text != "Buy milk" {
		t.Errorf("expected restored task %q, got %+v", "Buy milk", tasks)
	}

	content, err := os.ReadFile(todoFile.FilePath)
	if err != nil {
		t.Fatalf("expected restored state to be saved: %v", err)
	}
	if !strings.Contains(string(content), "Buy milk") {
		t.Errorf("expected saved file to contain restored task, got: %q", content)
	}
}

func TestPerformUndoTogglesLastChange(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "TODO.md")
	todoFile := core.NewTodoFile(path, "# TODO\n\n- [ ] Buy milk\n")
	if err := todoFile.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if _, err := PerformDeleteTask(todoFile, 1); err != nil {
		t.Fatalf("PerformDeleteTask failed: %v", err)
	}
	if len(todoFile.GetTasks()) != 0 {
		t.Fatalf("expected task deleted before undo, got %d tasks", len(todoFile.GetTasks()))
	}

	if err := PerformUndo(todoFile, path); err != nil {
		t.Fatalf("PerformUndo failed: %v", err)
	}
	tasks := todoFile.GetTasks()
	if len(tasks) != 1 || tasks[0].Text != "Buy milk" {
		t.Fatalf("expected undo to restore the deleted task, got %+v", tasks)
	}

	// Undo again: because the restore was written through Save(), the
	// pre-undo (deleted) state became the new backup — this re-applies
	// the delete.
	if err := PerformUndo(todoFile, path); err != nil {
		t.Fatalf("second PerformUndo failed: %v", err)
	}
	if len(todoFile.GetTasks()) != 0 {
		t.Errorf("expected second undo to re-apply the delete, got %d tasks", len(todoFile.GetTasks()))
	}
}

func TestPerformUndoNothingToUndo(t *testing.T) {
	todoFile := newTestTodoFile(t, "# TODO\n\n- [ ] Buy milk\n")
	if err := PerformUndo(todoFile, todoFile.FilePath); err != ErrNothingToUndo {
		t.Errorf("expected ErrNothingToUndo, got %v", err)
	}
}

func stringPtr(s string) *string { return &s }

func TestPerformEditTask(t *testing.T) {
	todoFile := newTestTodoFile(t, "# TODO\n\n- [ ] Buy milk\n")
	task, err := PerformEditTask(todoFile, 1, "Buy oat milk")
	if err != nil {
		t.Fatalf("PerformEditTask failed: %v", err)
	}
	if task.Text != "Buy oat milk" {
		t.Errorf("expected updated text %q, got %q", "Buy oat milk", task.Text)
	}
}

func TestPerformEditTaskNotFound(t *testing.T) {
	todoFile := newTestTodoFile(t, "# TODO\n\n- [ ] Buy milk\n")
	if _, err := PerformEditTask(todoFile, 99, "text"); err == nil {
		t.Error("expected error for nonexistent task, got nil")
	}
}

func TestPerformCompleteTask(t *testing.T) {
	todoFile := newTestTodoFile(t, "# TODO\n\n- [ ] Buy milk\n")
	task, err := PerformCompleteTask(todoFile, 1)
	if err != nil {
		t.Fatalf("PerformCompleteTask failed: %v", err)
	}
	if task.Status != core.TaskCompleted {
		t.Errorf("expected task to be completed, got %v", task.Status)
	}
}

func TestPerformCompleteTasks(t *testing.T) {
	todoFile := newTestTodoFile(t, "# TODO\n\n- [ ] Buy milk\n- [ ] Walk dog\n- [ ] Call mom\n")
	if _, err := PerformSetTaskIDs(todoFile, "abc"); err != nil {
		t.Fatalf("PerformSetTaskIDs failed: %v", err)
	}

	completed, failed, err := PerformCompleteTasks(todoFile, []string{"1", "ABC-002", "99"})
	if err != nil {
		t.Fatalf("PerformCompleteTasks failed: %v", err)
	}
	if len(completed) != 2 {
		t.Errorf("expected 2 completed tasks, got %d", len(completed))
	}
	if len(failed) != 1 || failed[0] != "99" {
		t.Errorf("expected failed IDs [99], got %v", failed)
	}
}

func TestPerformCompleteTasksAllFail(t *testing.T) {
	todoFile := newTestTodoFile(t, "# TODO\n\n- [ ] Buy milk\n")
	completed, failed, err := PerformCompleteTasks(todoFile, []string{"99"})
	if err != nil {
		t.Fatalf("PerformCompleteTasks failed: %v", err)
	}
	if len(completed) != 0 {
		t.Errorf("expected no completed tasks, got %d", len(completed))
	}
	if len(failed) != 1 {
		t.Errorf("expected 1 failed ID, got %d", len(failed))
	}
}

func TestPerformDeleteCompletedTasks(t *testing.T) {
	todoFile := newTestTodoFile(t, "# TODO\n\n- [ ] Buy milk\n- [x] Walk dog\n- [x] Call mom\n")
	count, err := PerformDeleteCompletedTasks(todoFile)
	if err != nil {
		t.Fatalf("PerformDeleteCompletedTasks failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 deleted tasks, got %d", count)
	}
	if len(todoFile.GetTasks()) != 1 {
		t.Errorf("expected 1 remaining task, got %d", len(todoFile.GetTasks()))
	}
}

func TestPerformDeleteCompletedTasksNoneCompleted(t *testing.T) {
	todoFile := newTestTodoFile(t, "# TODO\n\n- [ ] Buy milk\n")
	count, err := PerformDeleteCompletedTasks(todoFile)
	if err != nil {
		t.Fatalf("PerformDeleteCompletedTasks failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 deleted tasks, got %d", count)
	}
}

func TestPerformArchiveCompletedTasks(t *testing.T) {
	todoFile := newTestTodoFile(t, "# TODO\n\n- [ ] Buy milk\n\n## Bugs\n\n- [x] Fix login bug\n")
	count, err := PerformArchiveCompletedTasks(todoFile)
	if err != nil {
		t.Fatalf("PerformArchiveCompletedTasks failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 archived task, got %d", count)
	}
	archived := todoFile.GetTasksBySection(stringPtr("Archive"))
	if len(archived) != 1 || archived[0].Text != "Fix login bug (from Bugs)" {
		t.Errorf("expected task archived with origin noted, got %+v", archived)
	}
}

func TestPerformArchiveCompletedTasksNoneCompleted(t *testing.T) {
	todoFile := newTestTodoFile(t, "# TODO\n\n- [ ] Buy milk\n")
	count, err := PerformArchiveCompletedTasks(todoFile)
	if err != nil {
		t.Fatalf("PerformArchiveCompletedTasks failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 archived tasks, got %d", count)
	}
}

func TestPerformSetTaskIDs(t *testing.T) {
	todoFile := newTestTodoFile(t, "# TODO\n\n- [ ] Buy milk\n- [ ] Walk dog\n")
	count, err := PerformSetTaskIDs(todoFile, "abc")
	if err != nil {
		t.Fatalf("PerformSetTaskIDs failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 tagged tasks, got %d", count)
	}
}

func TestPerformSetTaskIDsInvalidPrefix(t *testing.T) {
	todoFile := newTestTodoFile(t, "# TODO\n\n- [ ] Buy milk\n")
	if _, err := PerformSetTaskIDs(todoFile, "abcd"); err == nil {
		t.Error("expected error for invalid prefix, got nil")
	}
}

func TestPerformRemoveTaskIDs(t *testing.T) {
	todoFile := newTestTodoFile(t, "# TODO\n\n- [ ] Buy milk\n")
	if _, err := PerformSetTaskIDs(todoFile, "abc"); err != nil {
		t.Fatalf("PerformSetTaskIDs failed: %v", err)
	}

	changed, err := PerformRemoveTaskIDs(todoFile)
	if err != nil {
		t.Fatalf("PerformRemoveTaskIDs failed: %v", err)
	}
	if !changed {
		t.Error("expected RemoveTaskIDs to report a change")
	}

	changed, err = PerformRemoveTaskIDs(todoFile)
	if err != nil {
		t.Fatalf("PerformRemoveTaskIDs failed: %v", err)
	}
	if changed {
		t.Error("expected no-op when no IDs remain")
	}
}

func TestPerformAddTaskNoteAppendsAndSaves(t *testing.T) {
	todoFile := newTestTodoFile(t, "# TODO\n\n- [ ] Buy milk\n")
	task, err := PerformAddTaskNote(todoFile, 1, "remember the receipt")
	if err != nil {
		t.Fatalf("PerformAddTaskNote failed: %v", err)
	}
	if len(task.Notes) != 1 || task.Notes[0] != "remember the receipt" {
		t.Errorf("expected note appended, got %v", task.Notes)
	}

	content, err := os.ReadFile(todoFile.FilePath)
	if err != nil {
		t.Fatalf("expected file saved: %v", err)
	}
	if !strings.Contains(string(content), "remember the receipt") {
		t.Errorf("expected saved content to include the note, got: %q", content)
	}
}

func TestPerformAddTaskNoteRejectsEmptyText(t *testing.T) {
	todoFile := newTestTodoFile(t, "# TODO\n\n- [ ] Buy milk\n")
	if _, err := PerformAddTaskNote(todoFile, 1, "   "); err == nil {
		t.Error("expected error for empty note text, got nil")
	}
}

func TestPerformAddTaskNoteUnknownIDReturnsError(t *testing.T) {
	todoFile := newTestTodoFile(t, "# TODO\n\n- [ ] Buy milk\n")
	if _, err := PerformAddTaskNote(todoFile, 99, "a note"); err == nil {
		t.Error("expected error for nonexistent task, got nil")
	}
}

func TestPerformLint(t *testing.T) {
	todoFile := newTestTodoFile(t, "# TODO\n\n- [] Buy milk\n")
	result, err := PerformLint(todoFile)
	if err != nil {
		t.Fatalf("PerformLint failed: %v", err)
	}
	if result.FixedCount == 0 {
		t.Error("expected lint to fix the empty-checkbox issue")
	}

	content, err := os.ReadFile(todoFile.FilePath)
	if err != nil {
		t.Fatalf("expected fixed file to be saved: %v", err)
	}
	if !strings.Contains(string(content), "- [ ] Buy milk") {
		t.Errorf("expected saved content to include the fix, got: %q", content)
	}
}

func TestPerformLintNoIssues(t *testing.T) {
	todoFile := newTestTodoFile(t, "# TODO\n\n- [ ] Buy milk\n")
	result, err := PerformLint(todoFile)
	if err != nil {
		t.Fatalf("PerformLint failed: %v", err)
	}
	if result.FixedCount != 0 || len(result.Issues) != 0 {
		t.Errorf("expected no issues, got %+v", result)
	}
}
