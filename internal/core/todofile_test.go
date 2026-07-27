package core

import (
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/i-am-fran/markdown-do/internal/fsutil"
)

func TestSetTaskIDsAssignsSequentialTags(t *testing.T) {
	content := "# TODO\n\n- [ ] Buy milk\n- [x] Walk dog\n- [ ] Call mom\n"
	tf := NewTodoFile("TODO.md", content)

	if err := tf.SetTaskIDs("abc"); err != nil {
		t.Fatalf("SetTaskIDs failed: %v", err)
	}

	want := []string{"ABC-003", "ABC-002", "ABC-001"}
	tasks := tf.GetTasks()
	if len(tasks) != len(want) {
		t.Fatalf("expected %d tasks, got %d", len(want), len(tasks))
	}
	for i, task := range tasks {
		if task.StableID != want[i] {
			t.Errorf("task %d: expected StableID %q, got %q", i, want[i], task.StableID)
		}
	}
}

func TestSetTaskIDsRejectsInvalidPrefix(t *testing.T) {
	tf := NewTodoFile("TODO.md", "# TODO\n\n- [ ] Buy milk\n")
	if err := tf.SetTaskIDs("abcd"); err == nil {
		t.Error("expected error for 4-letter prefix, got nil")
	}
	if err := tf.SetTaskIDs("a1"); err == nil {
		t.Error("expected error for non-alphabetic prefix, got nil")
	}
}

func TestSetTaskIDsOverwritesExisting(t *testing.T) {
	tf := NewTodoFile("TODO.md", "# TODO\n\n- [ ] [XYZ01] Buy milk\n")
	if err := tf.SetTaskIDs("abc"); err != nil {
		t.Fatalf("SetTaskIDs failed: %v", err)
	}
	task := tf.GetTasks()[0]
	if task.StableID != "ABC-001" {
		t.Errorf("expected overwritten StableID %q, got %q", "ABC-001", task.StableID)
	}
}

func TestRemoveTaskIDs(t *testing.T) {
	tf := NewTodoFile("TODO.md", "# TODO\n\n- [ ] Buy milk\n- [x] Walk dog\n")
	if err := tf.SetTaskIDs("abc"); err != nil {
		t.Fatalf("SetTaskIDs failed: %v", err)
	}

	if changed := tf.RemoveTaskIDs(); !changed {
		t.Fatal("expected RemoveTaskIDs to report a change")
	}
	for _, task := range tf.GetTasks() {
		if task.StableID != "" {
			t.Errorf("expected empty StableID after removal, got %q", task.StableID)
		}
	}

	if changed := tf.RemoveTaskIDs(); changed {
		t.Error("expected RemoveTaskIDs to be a no-op when no IDs remain")
	}
}

func TestAddTaskContinuesActiveSequence(t *testing.T) {
	tf := NewTodoFile("TODO.md", "# TODO\n\n- [ ] Buy milk\n- [x] Walk dog\n")
	if err := tf.SetTaskIDs("abc"); err != nil {
		t.Fatalf("SetTaskIDs failed: %v", err)
	}

	task, err := tf.AddTask("Call mom")
	if err != nil {
		t.Fatalf("AddTask failed: %v", err)
	}
	if task.StableID != "ABC-003" {
		t.Errorf("expected new task to continue sequence with %q, got %q", "ABC-003", task.StableID)
	}
}

func TestAddTaskContinuesSequencePastThreeDigits(t *testing.T) {
	tf := NewTodoFile("TODO.md", "# TODO\n\n- [ ] [ABC-999] Buy milk\n")

	task, err := tf.AddTask("Call mom")
	if err != nil {
		t.Fatalf("AddTask failed: %v", err)
	}
	if task.StableID != "ABC-1000" {
		t.Errorf("expected sequence to grow past 3 digits instead of truncating, got %q", task.StableID)
	}
}

func TestSetTaskIDsAcceptsOldTwoDigitFormat(t *testing.T) {
	tf := NewTodoFile("TODO.md", "# TODO\n\n- [ ] [XYZ29] Buy milk\n")

	if tf.GetTaskByStableID("XYZ29") == nil {
		t.Error("expected old-format (no separator) StableID to still parse and be found")
	}

	task, err := tf.AddTask("Call mom")
	if err != nil {
		t.Fatalf("AddTask failed: %v", err)
	}
	if task.StableID != "XYZ-030" {
		t.Errorf("expected new task tagged from an old-format sequence to use the new 3-digit dashed format, got %q", task.StableID)
	}
}

func TestAddTaskNoIDWhenSequenceInactive(t *testing.T) {
	tf := NewTodoFile("TODO.md", "# TODO\n\n- [ ] Buy milk\n")
	task, err := tf.AddTask("Call mom")
	if err != nil {
		t.Fatalf("AddTask failed: %v", err)
	}
	if task.StableID != "" {
		t.Errorf("expected no StableID when no ID scheme is active, got %q", task.StableID)
	}
}

func TestAddTaskEscapedAtSignStaysInInbox(t *testing.T) {
	tf := NewTodoFile("TODO.md", "# TODO\n\n")
	task, err := tf.AddTask(`Meet Bob \@bb`)
	if err != nil {
		t.Fatalf("AddTask failed: %v", err)
	}
	if task.Section != nil {
		t.Errorf("expected escaped @ to not be read as a section tag, got section %q", *task.Section)
	}
	if task.Text != "Meet Bob @bb" {
		t.Errorf("expected literal text %q, got %q", "Meet Bob @bb", task.Text)
	}
}

func TestGetTaskByStableIDCaseInsensitive(t *testing.T) {
	tf := NewTodoFile("TODO.md", "# TODO\n\n- [ ] Buy milk\n")
	if err := tf.SetTaskIDs("abc"); err != nil {
		t.Fatalf("SetTaskIDs failed: %v", err)
	}

	if tf.GetTaskByStableID("abc-001") == nil {
		t.Error("expected case-insensitive lookup to find task")
	}
	if tf.GetTaskByStableID("ABC-001") == nil {
		t.Error("expected uppercase lookup to find task")
	}
	if tf.GetTaskByStableID("nonexistent") != nil {
		t.Error("expected nil for unknown StableID")
	}
}

func TestLoadDoesNotAutoAssignIDs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "TODO.md")

	tf := NewTodoFile(path, "# TODO\n\n- [ ] Buy milk\n")
	if err := tf.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	reloaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	for _, task := range reloaded.GetTasks() {
		if task.StableID != "" {
			t.Errorf("expected Load to leave tasks untagged, got StableID %q", task.StableID)
		}
	}
}

func TestParseAttachesNotesToTask(t *testing.T) {
	content := "# TODO\n\n- [ ] Some task\n  - a note about the task\n  - another note\n"
	tf := NewTodoFile("TODO.md", content)
	tasks := tf.GetTasks()
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	want := []string{"a note about the task", "another note"}
	if len(tasks[0].Notes) != len(want) {
		t.Fatalf("expected %d notes, got %d: %v", len(want), len(tasks[0].Notes), tasks[0].Notes)
	}
	for i := range want {
		if tasks[0].Notes[i] != want[i] {
			t.Errorf("note %d: expected %q, got %q", i, want[i], tasks[0].Notes[i])
		}
	}
}

func TestParseTaskWithoutNotesHasEmptyNotes(t *testing.T) {
	tf := NewTodoFile("TODO.md", "# TODO\n\n- [ ] Some task\n")
	tasks := tf.GetTasks()
	if len(tasks[0].Notes) != 0 {
		t.Errorf("expected no notes, got %v", tasks[0].Notes)
	}
}

func TestParseNotesStopAtBlankLine(t *testing.T) {
	content := "# TODO\n\n- [ ] Some task\n  - a note\n\n- [ ] Another task\n"
	tf := NewTodoFile("TODO.md", content)
	tasks := tf.GetTasks()
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
	if len(tasks[0].Notes) != 1 || tasks[0].Notes[0] != "a note" {
		t.Errorf("expected task 0 to have 1 note, got %v", tasks[0].Notes)
	}
	if len(tasks[1].Notes) != 0 {
		t.Errorf("expected task 1 to have no notes, got %v", tasks[1].Notes)
	}
}

func TestParseNotesStopAtNewTask(t *testing.T) {
	content := "# TODO\n\n- [ ] Task one\n  - note for one\n- [ ] Task two\n"
	tf := NewTodoFile("TODO.md", content)
	tasks := tf.GetTasks()
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
	if len(tasks[0].Notes) != 1 || tasks[0].Notes[0] != "note for one" {
		t.Errorf("expected task 0 to have 1 note, got %v", tasks[0].Notes)
	}
	if len(tasks[1].Notes) != 0 {
		t.Errorf("expected task 1 to have no notes, got %v", tasks[1].Notes)
	}
}

func TestParseNotesStopAtNonIndentedLine(t *testing.T) {
	content := "# TODO\n\n- [ ] Some task\n  - a note\nsome stray plain text\n"
	tf := NewTodoFile("TODO.md", content)
	tasks := tf.GetTasks()
	if len(tasks[0].Notes) != 1 || tasks[0].Notes[0] != "a note" {
		t.Errorf("expected 1 note, got %v", tasks[0].Notes)
	}
}

func TestParseIndentedChecklistLineIsNewTaskNotNote(t *testing.T) {
	content := "# TODO\n\n- [ ] Parent task\n  - [ ] Child checklist item\n"
	tf := NewTodoFile("TODO.md", content)
	tasks := tf.GetTasks()
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks (indented checkbox is a sibling task), got %d", len(tasks))
	}
	if len(tasks[0].Notes) != 0 {
		t.Errorf("expected parent task to have no notes, got %v", tasks[0].Notes)
	}
	if tasks[1].Text != "Child checklist item" {
		t.Errorf("expected second task text %q, got %q", "Child checklist item", tasks[1].Text)
	}
	if tasks[1].Indent != "  " {
		t.Errorf("expected second task Indent %q, got %q", "  ", tasks[1].Indent)
	}
}

// TestSaveRoundTripsIndentedChecklistItem pins the fix for indentation being
// silently discarded: ParseTaskLine used to throw away the leading
// whitespace it captured, and FormatTask always emitted a top-level "- ",
// so any mutation of an indented/nested checklist item flattened it.
func TestSaveRoundTripsIndentedChecklistItem(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "TODO.md")
	tf := NewTodoFile(path, "# TODO\n\n- [ ] Parent task\n  - [ ] Child checklist item\n")

	if err := tf.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	reloaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if !strings.Contains(reloaded.Serialize(), "  - [ ] Child checklist item") {
		t.Errorf("expected indentation preserved through save/reload, got:\n%s", reloaded.Serialize())
	}
}

func TestAddTaskNoteOnIndentedTaskNestsUnderIt(t *testing.T) {
	content := "# TODO\n\n- [ ] Parent task\n  - [ ] Child checklist item\n"
	tf := NewTodoFile("TODO.md", content)
	child := tf.GetTasks()[1]

	if ok := tf.AddTaskNote(child.ID, "a note"); !ok {
		t.Fatal("expected AddTaskNote to succeed")
	}

	if !strings.Contains(tf.Serialize(), "    - a note") {
		t.Errorf("expected the note nested two levels deep under the indented task, got:\n%s", tf.Serialize())
	}
}

func TestParseConsecutiveTasksEachWithNotes(t *testing.T) {
	content := "# TODO\n\n- [ ] Task one\n  - note one-a\n- [ ] Task two\n  - note two-a\n  - note two-b\n"
	tf := NewTodoFile("TODO.md", content)
	tasks := tf.GetTasks()
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
	if len(tasks[0].Notes) != 1 || tasks[0].Notes[0] != "note one-a" {
		t.Errorf("expected task 0 notes [note one-a], got %v", tasks[0].Notes)
	}
	if len(tasks[1].Notes) != 2 || tasks[1].Notes[0] != "note two-a" || tasks[1].Notes[1] != "note two-b" {
		t.Errorf("expected task 1 notes [note two-a note two-b], got %v", tasks[1].Notes)
	}
}

func TestSerializeRoundTripsTaskNotes(t *testing.T) {
	content := "# TODO\n\n- [ ] Some task\n  - a note about the task\n  - another note\n"
	tf := NewTodoFile("TODO.md", content)
	if got := tf.Serialize(); got != content {
		t.Errorf("round trip mismatch:\ngot:  %q\nwant: %q", got, content)
	}
}

func TestDeleteTaskRemovesItsNotes(t *testing.T) {
	content := "# TODO\n\n- [ ] Task one\n  - note one\n  - note two\n- [ ] Task two\n"
	tf := NewTodoFile("TODO.md", content)
	if !tf.DeleteTask(1) {
		t.Fatal("expected DeleteTask to succeed")
	}
	tasks := tf.GetTasks()
	if len(tasks) != 1 || tasks[0].Text != "Task two" {
		t.Fatalf("expected only Task two to remain, got %+v", tasks)
	}
	if strings.Contains(tf.Serialize(), "note one") || strings.Contains(tf.Serialize(), "note two") {
		t.Errorf("expected notes removed along with task, got: %q", tf.Serialize())
	}
}

func TestMoveTaskCarriesNotesToNewSection(t *testing.T) {
	content := "# TODO\n\n- [ ] Task one\n  - note one\n"
	tf := NewTodoFile("TODO.md", content)
	section := "Errands"
	if !tf.MoveTask(1, &section) {
		t.Fatal("expected MoveTask to succeed")
	}
	task := tf.GetTask(1)
	if task == nil || task.Section == nil || *task.Section != "Errands" {
		t.Fatalf("expected task moved to Errands, got %+v", task)
	}
	if len(task.Notes) != 1 || task.Notes[0] != "note one" {
		t.Errorf("expected note to travel with task, got %v", task.Notes)
	}
}

func TestMoveTaskIntoNotesSectionKeepsNoteBullets(t *testing.T) {
	content := "# TODO\n\n- [ ] Task one\n  - note one\n"
	tf := NewTodoFile("TODO.md", content)
	if !tf.MoveTask(1, stringPtrCore("Notes")) {
		t.Fatal("expected MoveTask to succeed")
	}
	serialized := tf.Serialize()
	if !strings.Contains(serialized, "- Task one") {
		t.Errorf("expected task converted to plain list item, got: %q", serialized)
	}
	if !strings.Contains(serialized, "note one") {
		t.Errorf("expected note bullet preserved, got: %q", serialized)
	}
}

func TestUpdateTaskPreservesNotes(t *testing.T) {
	content := "# TODO\n\n- [ ] Task one\n  - note one\n"
	tf := NewTodoFile("TODO.md", content)
	if !tf.UpdateTask(1, "Task one edited") {
		t.Fatal("expected UpdateTask to succeed")
	}
	task := tf.GetTask(1)
	if len(task.Notes) != 1 || task.Notes[0] != "note one" {
		t.Errorf("expected note preserved after update, got %v", task.Notes)
	}
}

func TestToggleTaskPreservesNotes(t *testing.T) {
	content := "# TODO\n\n- [ ] Task one\n  - note one\n"
	tf := NewTodoFile("TODO.md", content)
	if !tf.ToggleTask(1, false) {
		t.Fatal("expected ToggleTask to succeed")
	}
	task := tf.GetTask(1)
	if len(task.Notes) != 1 || task.Notes[0] != "note one" {
		t.Errorf("expected note preserved after toggle, got %v", task.Notes)
	}
}

func TestToggleTaskBinaryFlip(t *testing.T) {
	tf := NewTodoFile("TODO.md", "# TODO\n\n- [ ] Task one\n")
	tf.ToggleTask(1, false)
	if got := tf.GetTask(1).Status; got != TaskCompleted {
		t.Errorf("expected TaskCompleted, got %v", got)
	}
	tf.ToggleTask(1, false)
	if got := tf.GetTask(1).Status; got != TaskPending {
		t.Errorf("expected TaskPending, got %v", got)
	}
}

func TestToggleTaskInProgressCycle(t *testing.T) {
	tf := NewTodoFile("TODO.md", "# TODO\n\n- [ ] Task one\n")

	tf.ToggleTask(1, true)
	if got := tf.GetTask(1).Status; got != TaskInProgress {
		t.Errorf("expected TaskInProgress, got %v", got)
	}
	tf.ToggleTask(1, true)
	if got := tf.GetTask(1).Status; got != TaskCompleted {
		t.Errorf("expected TaskCompleted, got %v", got)
	}
	tf.ToggleTask(1, true)
	if got := tf.GetTask(1).Status; got != TaskPending {
		t.Errorf("expected TaskPending, got %v", got)
	}
}

func TestSetTaskStatusPreservesNotes(t *testing.T) {
	content := "# TODO\n\n- [ ] Task one\n  - note one\n"
	tf := NewTodoFile("TODO.md", content)
	if !tf.SetTaskStatus(1, TaskCompleted) {
		t.Fatal("expected SetTaskStatus to succeed")
	}
	task := tf.GetTask(1)
	if len(task.Notes) != 1 || task.Notes[0] != "note one" {
		t.Errorf("expected note preserved after status change, got %v", task.Notes)
	}
}

func TestReorderOnSaveKeepsNotesWithTheirTask(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "TODO.md")
	content := "# TODO\n\n- [x] Completed task\n  - note a\n  - note b\n- [ ] Pending task\n"
	tf := NewTodoFile(path, content)
	if err := tf.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	reloaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	tasks := reloaded.GetTasks()
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
	if tasks[0].Text != "Pending task" || len(tasks[0].Notes) != 0 {
		t.Errorf("expected pending task first with no notes, got %+v", tasks[0])
	}
	if tasks[1].Text != "Completed task" || len(tasks[1].Notes) != 2 {
		t.Errorf("expected completed task last with 2 notes, got %+v", tasks[1])
	}
}

func TestAddTaskNoteAppendsNoteLine(t *testing.T) {
	tf := NewTodoFile("TODO.md", "# TODO\n\n- [ ] Task one\n")
	if !tf.AddTaskNote(1, "a new note") {
		t.Fatal("expected AddTaskNote to succeed")
	}
	task := tf.GetTask(1)
	if len(task.Notes) != 1 || task.Notes[0] != "a new note" {
		t.Errorf("expected note appended, got %v", task.Notes)
	}
}

func TestAddTaskNoteOnTaskWithExistingNotes(t *testing.T) {
	tf := NewTodoFile("TODO.md", "# TODO\n\n- [ ] Task one\n  - first note\n")
	if !tf.AddTaskNote(1, "second note") {
		t.Fatal("expected AddTaskNote to succeed")
	}
	task := tf.GetTask(1)
	want := []string{"first note", "second note"}
	if len(task.Notes) != len(want) {
		t.Fatalf("expected %d notes, got %d: %v", len(want), len(task.Notes), task.Notes)
	}
	for i := range want {
		if task.Notes[i] != want[i] {
			t.Errorf("note %d: expected %q, got %q", i, want[i], task.Notes[i])
		}
	}
}

func TestAddTaskNoteUnknownIDReturnsFalse(t *testing.T) {
	tf := NewTodoFile("TODO.md", "# TODO\n\n- [ ] Task one\n")
	if tf.AddTaskNote(99, "a note") {
		t.Error("expected AddTaskNote to return false for unknown ID")
	}
}

func TestReorderSkipsGroupWithUnexpectedGapLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "TODO.md")
	content := "# TODO\n\n- [x] Completed task\n\n- [ ] Pending task\n"
	tf := NewTodoFile(path, content)
	if err := tf.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	reloaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	tasks := reloaded.GetTasks()
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
	if tasks[0].Text != "Completed task" || tasks[1].Text != "Pending task" {
		t.Errorf("expected original order preserved when group has an unexpected gap, got %+v", tasks)
	}
}

func TestDeleteCompletedTasksRemovesNotes(t *testing.T) {
	content := "# TODO\n\n- [ ] Task pending\n- [x] Task done\n  - note a\n"
	tf := NewTodoFile("TODO.md", content)
	count := tf.DeleteCompletedTasks()
	if count != 1 {
		t.Fatalf("expected 1 deleted task, got %d", count)
	}
	if strings.Contains(tf.Serialize(), "note a") {
		t.Errorf("expected note removed along with completed task, got: %q", tf.Serialize())
	}
}

func stringPtrCore(s string) *string { return &s }

func TestArchiveCompletedTasksMovesToArchiveSection(t *testing.T) {
	content := "# TODO\n\n- [ ] Pending task\n\n## Bugs\n\n- [x] Fix login bug\n- [ ] Another bug\n"
	tf := NewTodoFile("TODO.md", content)
	count := tf.ArchiveCompletedTasks()
	if count != 1 {
		t.Fatalf("expected 1 archived task, got %d", count)
	}

	archived := tf.GetTasksBySection(stringPtrCore("Archive"))
	if len(archived) != 1 {
		t.Fatalf("expected 1 task in Archive section, got %d", len(archived))
	}
	if archived[0].Text != "Fix login bug (from Bugs)" {
		t.Errorf("expected origin section noted in text, got %q", archived[0].Text)
	}
	if archived[0].Status != TaskCompleted {
		t.Errorf("expected archived task to stay completed, got %v", archived[0].Status)
	}

	bugs := tf.GetTasksBySection(stringPtrCore("Bugs"))
	if len(bugs) != 1 || bugs[0].Text != "Another bug" {
		t.Errorf("expected Bugs section to keep only the pending task, got %+v", bugs)
	}
}

func TestArchiveCompletedTasksNoneCompleted(t *testing.T) {
	tf := NewTodoFile("TODO.md", "# TODO\n\n- [ ] Buy milk\n")
	count := tf.ArchiveCompletedTasks()
	if count != 0 {
		t.Errorf("expected 0 archived tasks, got %d", count)
	}
	if strings.Contains(tf.Serialize(), "Archive") {
		t.Errorf("expected no Archive section created, got: %q", tf.Serialize())
	}
}

func TestArchiveCompletedTasksPreservesNotes(t *testing.T) {
	content := "# TODO\n\n## Bugs\n\n- [x] Fix login bug\n  - root caused by cache\n"
	tf := NewTodoFile("TODO.md", content)
	count := tf.ArchiveCompletedTasks()
	if count != 1 {
		t.Fatalf("expected 1 archived task, got %d", count)
	}
	if !strings.Contains(tf.Serialize(), "root caused by cache") {
		t.Errorf("expected note to travel with archived task, got: %q", tf.Serialize())
	}
}

func TestArchiveCompletedTasksNoOriginSectionOmitsSuffix(t *testing.T) {
	tf := NewTodoFile("TODO.md", "# TODO\n\n- [x] Done already\n")
	tf.ArchiveCompletedTasks()
	archived := tf.GetTasksBySection(stringPtrCore("Archive"))
	if len(archived) != 1 || archived[0].Text != "Done already" {
		t.Errorf("expected task text unchanged when there's no origin section, got %+v", archived)
	}
}

func TestArchiveCompletedTasksSkipsAlreadyArchived(t *testing.T) {
	content := "# TODO\n\n## Bugs\n\n- [x] Fix login bug\n"
	tf := NewTodoFile("TODO.md", content)
	tf.ArchiveCompletedTasks()

	count := tf.ArchiveCompletedTasks()
	if count != 0 {
		t.Errorf("expected 0 tasks archived on second pass, got %d", count)
	}
	if strings.Count(tf.Serialize(), "(from Bugs)") != 1 {
		t.Errorf("expected origin annotated exactly once, got: %q", tf.Serialize())
	}
}

func TestArchiveStaysBeforeNotesAfterSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "TODO.md")
	content := "# TODO\n\n## Bugs\n\n- [x] Fix login bug\n\n## Notes\n\n- remember to floss\n"
	tf := NewTodoFile(path, content)
	tf.ArchiveCompletedTasks()
	if err := tf.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if _, err := tf.AddTask("New bug @Bugs"); err != nil {
		t.Fatalf("AddTask failed: %v", err)
	}
	if err := tf.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	names := tf.GetSectionNames()
	archiveIdx, notesIdx := -1, -1
	for i, n := range names {
		if n == "Archive" {
			archiveIdx = i
		}
		if n == "Notes" {
			notesIdx = i
		}
	}
	if archiveIdx == -1 || notesIdx == -1 || archiveIdx >= notesIdx {
		t.Errorf("expected Archive before Notes at the end, got sections: %v", names)
	}
}

func TestSaveSkipsBackupOnFirstSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "TODO.md")
	tf := NewTodoFile(path, "# TODO\n\n- [ ] Buy milk\n")
	if err := tf.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if HasUndo(path) {
		t.Error("expected no undo backup after the very first save")
	}
}

func TestSaveBacksUpPreviousContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "TODO.md")
	tf := NewTodoFile(path, "# TODO\n\n- [ ] Buy milk\n")
	if err := tf.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if !tf.DeleteTask(1) {
		t.Fatal("expected DeleteTask to succeed")
	}
	if err := tf.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if !HasUndo(path) {
		t.Fatal("expected an undo backup after the second save")
	}
	backup, err := LoadUndoBackup(path)
	if err != nil {
		t.Fatalf("LoadUndoBackup failed: %v", err)
	}
	if !strings.Contains(backup, "Buy milk") {
		t.Errorf("expected backup to contain pre-delete content, got: %q", backup)
	}
}

func TestSaveFailsWhenLockHeld(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "TODO.md")
	tf := NewTodoFile(path, "# TODO\n\n- [ ] Buy milk\n")

	lock, err := fsutil.AcquireLock(path)
	if err != nil {
		t.Fatalf("AcquireLock failed: %v", err)
	}
	defer lock.Unlock()

	if err := tf.Save(); err == nil {
		t.Fatal("expected Save to fail while another process holds the lock")
	}
}

// TestConcurrentSavesDoNotCorruptFile stress-tests Save() from many
// goroutines at once. Without a lock + atomic write, concurrent
// read-modify-write cycles could interleave and leave the file truncated
// or with mixed-up content; every attempt here must either succeed cleanly
// or fail with an error, and the file on disk must always end up valid,
// parseable markdown afterward.
func TestConcurrentSavesDoNotCorruptFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "TODO.md")
	seed := NewTodoFile(path, "# TODO\n\n- [ ] Buy milk\n")
	if err := seed.Save(); err != nil {
		t.Fatalf("seed Save failed: %v", err)
	}

	const goroutines = 8
	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			tf, err := Load(path)
			if err != nil {
				return
			}
			tf.AddTask("Task from goroutine")
			tf.Save() // best-effort; contention errors are expected and fine
		}(i)
	}
	wg.Wait()

	final, err := Load(path)
	if err != nil {
		t.Fatalf("expected the final file to still be loadable, got: %v", err)
	}
	tasks := final.GetTasks()
	if len(tasks) == 0 {
		t.Fatal("expected at least the seed task to survive")
	}
	for _, task := range tasks {
		if strings.TrimSpace(task.Text) == "" {
			t.Errorf("expected no corrupted/empty tasks, got %+v", tasks)
		}
	}
}

func TestSetTaskIDsPersistsAcrossLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "TODO.md")

	tf := NewTodoFile(path, "# TODO\n\n- [ ] Buy milk\n- [ ] Walk dog\n")
	if err := tf.SetTaskIDs("abc"); err != nil {
		t.Fatalf("SetTaskIDs failed: %v", err)
	}
	if err := tf.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	reloaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	tasks := reloaded.GetTasks()
	if tasks[0].StableID != "ABC-002" || tasks[1].StableID != "ABC-001" {
		t.Errorf("expected tags to persist across Load, got %q, %q", tasks[0].StableID, tasks[1].StableID)
	}
}

func TestGetActiveTasksExcludesCompletedOnly(t *testing.T) {
	content := "# TODO\n\n- [ ] Pending task\n- [/] In-progress task\n- [x] Completed task\n"
	tf := NewTodoFile("TODO.md", content)

	active := tf.GetActiveTasks()
	if len(active) != 2 {
		t.Fatalf("expected 2 active tasks, got %d", len(active))
	}
	for _, task := range active {
		if task.Status == TaskCompleted {
			t.Errorf("expected no completed tasks in GetActiveTasks result, got %v", task)
		}
	}
}
