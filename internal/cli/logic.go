package cli

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/i-am-fran/markdown-do/internal/config"
	"github.com/i-am-fran/markdown-do/internal/core"
)

// LoadDefaultTodoFile finds and loads the default TODO file for the current
// working directory. It does not create the file on disk.
func LoadDefaultTodoFile() (*core.TodoFile, string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, "", err
	}

	filePath, err := core.FindDefaultTodoFile(cwd)
	if err != nil {
		return nil, "", err
	}

	todoFile, err := core.Load(filePath)
	if err != nil {
		return nil, "", err
	}

	return todoFile, filePath, nil
}

// ensureTodoFileCreated creates filePath on disk with the default template
// the first time a totally-empty/nonexistent file is written to.
func ensureTodoFileCreated(todoFile *core.TodoFile, filePath string) (*core.TodoFile, error) {
	if len(todoFile.GetTasks()) == 0 && todoFile.Serialize() == "" {
		return core.Create(filePath)
	}
	return todoFile, nil
}

// PerformAddTask adds a task to todoFile and saves it, creating the file on
// disk first if it doesn't exist yet.
func PerformAddTask(todoFile *core.TodoFile, filePath, text string) (*core.TodoFile, *core.Task, error) {
	todoFile, err := ensureTodoFileCreated(todoFile, filePath)
	if err != nil {
		return nil, nil, err
	}

	task, err := todoFile.AddTask(text)
	if err != nil {
		return nil, nil, err
	}

	if err := todoFile.Save(); err != nil {
		return nil, nil, err
	}

	clearCacheWithWarning()
	return todoFile, task, nil
}

// PerformAddNote adds a note to todoFile and saves it, creating the file on
// disk first if it doesn't exist yet.
func PerformAddNote(todoFile *core.TodoFile, filePath, text string) (*core.TodoFile, error) {
	todoFile, err := ensureTodoFileCreated(todoFile, filePath)
	if err != nil {
		return nil, err
	}

	todoFile.AddNote(text)
	if err := todoFile.Save(); err != nil {
		return nil, err
	}

	return todoFile, nil
}

// PerformToggleTask toggles a task's status and saves.
func PerformToggleTask(todoFile *core.TodoFile, id int) (*core.Task, error) {
	if !todoFile.ToggleTask(id, config.GetSettings().EnableInProgress) {
		return nil, fmt.Errorf("task %d not found", id)
	}

	if err := todoFile.Save(); err != nil {
		return nil, err
	}

	clearCacheWithWarning()
	return todoFile.GetTask(id), nil
}

// PerformDeleteTask deletes a task and saves, returning the task as it was
// immediately before deletion.
func PerformDeleteTask(todoFile *core.TodoFile, id int) (*core.Task, error) {
	task := todoFile.GetTask(id)
	if task == nil {
		return nil, fmt.Errorf("task %d not found", id)
	}
	deleted := *task

	if !todoFile.DeleteTask(id) {
		return nil, fmt.Errorf("task %d not found", id)
	}

	if err := todoFile.Save(); err != nil {
		return nil, err
	}

	clearCacheWithWarning()
	return &deleted, nil
}

// PerformMoveTask moves a task to a different section (nil for the inbox)
// and saves.
func PerformMoveTask(todoFile *core.TodoFile, id int, section *string) (*core.Task, error) {
	if !todoFile.MoveTask(id, section) {
		return nil, fmt.Errorf("task %d not found", id)
	}

	if err := todoFile.Save(); err != nil {
		return nil, err
	}

	clearCacheWithWarning()
	return todoFile.GetTask(id), nil
}

// PerformRestore restores todoFile to a prior snapshot (as produced by
// core.TodoFile.Snapshot) and saves.
func PerformRestore(todoFile *core.TodoFile, snapshot []string) error {
	todoFile.Restore(snapshot)
	if err := todoFile.Save(); err != nil {
		return err
	}

	clearCacheWithWarning()
	return nil
}

// ErrNothingToUndo is returned by PerformUndo when no undo backup exists.
var ErrNothingToUndo = errors.New("nothing to undo")

// PerformUndo restores todoFile from its on-disk undo backup and saves.
// Because the restore is written through the normal Save() path, the
// pre-undo state becomes the new backup — running PerformUndo again
// re-applies the change it just undid.
func PerformUndo(todoFile *core.TodoFile, filePath string) error {
	backup, err := core.LoadUndoBackup(filePath)
	if err != nil {
		return ErrNothingToUndo
	}

	todoFile.Restore(strings.Split(backup, "\n"))
	if err := todoFile.Save(); err != nil {
		return err
	}

	clearCacheWithWarning()
	return nil
}

// PerformEditTask updates a task's text and saves.
func PerformEditTask(todoFile *core.TodoFile, id int, newText string) (*core.Task, error) {
	if !todoFile.UpdateTask(id, newText) {
		return nil, fmt.Errorf("task %d not found", id)
	}

	if err := todoFile.Save(); err != nil {
		return nil, err
	}

	clearCacheWithWarning()
	return todoFile.GetTask(id), nil
}

// PerformAddTaskNote appends a note to a task's block and saves.
func PerformAddTaskNote(todoFile *core.TodoFile, id int, text string) (*core.Task, error) {
	if strings.TrimSpace(text) == "" {
		return nil, errors.New("note text must not be empty")
	}

	if !todoFile.AddTaskNote(id, text) {
		return nil, fmt.Errorf("task %d not found", id)
	}

	if err := todoFile.Save(); err != nil {
		return nil, err
	}

	clearCacheWithWarning()
	return todoFile.GetTask(id), nil
}

// PerformCompleteTask marks a task completed and saves.
func PerformCompleteTask(todoFile *core.TodoFile, id int) (*core.Task, error) {
	if !todoFile.SetTaskStatus(id, core.TaskCompleted) {
		return nil, fmt.Errorf("task %d not found", id)
	}

	if err := todoFile.Save(); err != nil {
		return nil, err
	}

	clearCacheWithWarning()
	return todoFile.GetTask(id), nil
}

// PerformCompleteTasks marks multiple tasks (by numeric ID or stable ID) as
// completed and saves if at least one succeeded.
func PerformCompleteTasks(todoFile *core.TodoFile, idsStr []string) (completed []core.Task, failedIDs []string, err error) {
	for _, idStr := range idsStr {
		var task *core.Task
		if id, convErr := strconv.Atoi(idStr); convErr == nil {
			task = todoFile.GetTask(id)
		} else {
			task = todoFile.GetTaskByStableID(idStr)
		}
		if task == nil {
			failedIDs = append(failedIDs, idStr)
			continue
		}

		id := task.ID
		if todoFile.SetTaskStatus(id, core.TaskCompleted) {
			completed = append(completed, *todoFile.GetTask(id))
		} else {
			failedIDs = append(failedIDs, idStr)
		}
	}

	if len(completed) == 0 {
		return completed, failedIDs, nil
	}

	if err := todoFile.Save(); err != nil {
		return nil, nil, err
	}

	clearCacheWithWarning()
	return completed, failedIDs, nil
}

// PerformDeleteCompletedTasks deletes all completed tasks and saves if any
// were removed.
func PerformDeleteCompletedTasks(todoFile *core.TodoFile) (int, error) {
	count := todoFile.DeleteCompletedTasks()
	if count == 0 {
		return 0, nil
	}

	if err := todoFile.Save(); err != nil {
		return 0, err
	}

	clearCacheWithWarning()
	return count, nil
}

// PerformArchiveCompletedTasks moves all completed tasks into the Archive
// section and saves if any were archived.
func PerformArchiveCompletedTasks(todoFile *core.TodoFile) (int, error) {
	count := todoFile.ArchiveCompletedTasks()
	if count == 0 {
		return 0, nil
	}

	if err := todoFile.Save(); err != nil {
		return 0, err
	}

	clearCacheWithWarning()
	return count, nil
}

// PerformSetTaskIDs tags every task in todoFile with sequential IDs and saves.
func PerformSetTaskIDs(todoFile *core.TodoFile, prefix string) (int, error) {
	if err := todoFile.SetTaskIDs(prefix); err != nil {
		return 0, err
	}

	if err := todoFile.Save(); err != nil {
		return 0, err
	}

	clearCacheWithWarning()
	return len(todoFile.GetTasks()), nil
}

// PerformRemoveTaskIDs strips ID tags from every task in todoFile and saves
// if anything changed.
func PerformRemoveTaskIDs(todoFile *core.TodoFile) (bool, error) {
	if !todoFile.RemoveTaskIDs() {
		return false, nil
	}

	if err := todoFile.Save(); err != nil {
		return false, err
	}

	clearCacheWithWarning()
	return true, nil
}

// PerformLint lints todoFile and saves if any issues were fixed.
func PerformLint(todoFile *core.TodoFile) (core.LintResult, error) {
	result := todoFile.Lint()
	if result.FixedCount > 0 {
		if err := todoFile.Save(); err != nil {
			return result, err
		}
	}
	return result, nil
}
