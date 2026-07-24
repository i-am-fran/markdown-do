package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/fatih/color"
	"github.com/i-am-fran/markdowndo/internal/config"
	"github.com/i-am-fran/markdowndo/internal/core"
)

const Version = "1.1.0"

var (
	green         = color.New(color.FgGreen).SprintFunc()
	yellow        = color.New(color.FgYellow).SprintFunc()
	red           = color.New(color.FgRed).SprintFunc()
	cyan          = color.New(color.FgCyan).SprintFunc()
	magenta       = color.New(color.FgMagenta).SprintFunc()
	bold          = color.New(color.Bold).SprintFunc()
	strikethrough = color.New(color.CrossedOut).SprintFunc()
)

func formatTaskLine(task *core.Task, showFile string, displayID int) string {
	var checkbox, text string

	id := displayID
	if id == 0 {
		id = task.ID
	}

	taskText := task.Text
	if task.StableID != "" {
		taskText = color.New(color.FgHiBlack).Sprint("["+task.StableID+"]") + " " + taskText
	}

	switch task.Status {
	case core.TaskCompleted:
		checkbox = green("[x]")
		text = strikethrough(taskText)
	default:
		checkbox = "[ ]"
		text = taskText
	}

	idStr := fmt.Sprintf("%d.", id)
	file := ""
	if showFile != "" {
		file = fmt.Sprintf(" (%s)", showFile)
	}

	line := fmt.Sprintf("  %s %s %s%s", idStr, checkbox, text, file)
	for _, note := range task.Notes {
		line += "\n" + color.New(color.FgHiBlack).Sprintf("      - %s", note)
	}
	return line
}

// loadTaskFile loads the task and file for a given ID (positional number or
// stable ID string), checking cache if needed
func loadTaskFile(idStr string) (*core.TodoFile, *core.Task, int, string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, nil, 0, "", err
	}

	filePath, err := core.FindDefaultTodoFile(cwd)
	if err != nil {
		return nil, nil, 0, "", err
	}

	todoFile, err := core.Load(filePath)
	if err != nil {
		return nil, nil, 0, "", err
	}

	id, isNumeric := 0, true
	if n, err := strconv.Atoi(idStr); err == nil {
		id = n
	} else {
		isNumeric = false
	}

	if !isNumeric {
		task := todoFile.GetTaskByStableID(idStr)
		if task == nil {
			return nil, nil, 0, "", fmt.Errorf("task %s not found", idStr)
		}
		return todoFile, task, task.ID, filePath, nil
	}

	task := todoFile.GetTask(id)
	localID := id

	if task == nil {
		// Task not found in local file, try to use cache
		cache, err := core.LoadCache()
		if err == nil && cache != nil {
			if cachedTask, exists := cache.Tasks[id]; exists {
				// Load the file from cache
				todoFile, err = core.Load(cachedTask.FilePath)
				if err != nil {
					return nil, nil, 0, "", fmt.Errorf("error loading cached file %s: %v", cachedTask.FilePath, err)
				}

				// Get the task using the local ID
				task = todoFile.GetTask(cachedTask.LocalID)
				if task == nil {
					return nil, nil, 0, "", fmt.Errorf("task %d not found in cached file", id)
				}

				// Use the cached file path and local ID
				filePath = cachedTask.FilePath
				localID = cachedTask.LocalID
			} else {
				return nil, nil, 0, "", fmt.Errorf("task %d not found", id)
			}
		} else {
			return nil, nil, 0, "", fmt.Errorf("task %d not found", id)
		}
	}

	return todoFile, task, localID, filePath, nil
}

// clearCacheWithWarning clears the cache and logs a warning if it fails
func clearCacheWithWarning() {
	if err := core.ClearCache(); err != nil {
		fmt.Fprintf(os.Stderr, yellow("Warning: Could not clear task cache: %v\n"), err)
	}
}

// ListTasks lists all tasks
func ListTasks(recursive bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if recursive {
		filesByDir, err := core.FindTodoFilesInSubdirs(cwd)
		if err != nil {
			return err
		}

		if len(filesByDir) == 0 {
			fmt.Println("No TODO files found")
			return nil
		}

		globalTaskID := 1
		cache := &core.Cache{
			Tasks: make(map[int]core.TaskCache),
		}

		for dir, files := range filesByDir {
			fmt.Println(bold(cyan(dir)))

			for _, file := range files {
				todoFile, err := core.Load(file.Path)
				if err != nil {
					continue
				}

				tasks := todoFile.GetTasks()
				if len(tasks) == 0 {
					fmt.Println("  (no tasks)")
				} else {
					for _, task := range tasks {
						fmt.Println(formatTaskLine(&task, "", globalTaskID))

						// Store task in cache
						cache.Tasks[globalTaskID] = core.TaskCache{
							GlobalID: globalTaskID,
							FilePath: file.Path,
							LocalID:  task.ID,
							TaskText: task.Text,
						}

						globalTaskID++
					}
				}
			}
			fmt.Println()
		}

		// Save cache to disk
		if err := core.SaveCache(cache); err != nil {
			// Don't fail if cache can't be saved, just log a warning
			fmt.Fprintf(os.Stderr, yellow("Warning: Could not save task cache: %v\n"), err)
		}
	} else {
		// Clear cache when listing non-recursively
		clearCacheWithWarning()
		filePath, err := core.FindDefaultTodoFile(cwd)
		if err != nil {
			return err
		}

		todoFile, err := core.Load(filePath)
		if err != nil {
			return err
		}

		groups := todoFile.GetTasksGroupedBySectionOrdered()
		if len(groups) == 0 {
			fmt.Println("No tasks found")
			return nil
		}

		for _, group := range groups {
			if group.Section == nil {
				fmt.Println(bold(cyan("Tasks:")))
			} else {
				fmt.Println()
				fmt.Println(bold(magenta("## " + *group.Section)))
			}

			for _, task := range group.Tasks {
				fmt.Println(formatTaskLine(&task, "", 0))
			}
		}
	}

	return nil
}

// AddTask adds a new task
func AddTask(text string) error {
	todoFile, filePath, err := LoadDefaultTodoFile()
	if err != nil {
		return err
	}

	_, task, err := PerformAddTask(todoFile, filePath, text)
	if err != nil {
		return err
	}

	fmt.Print(green("Added: "))
	fmt.Println(task.Text)
	return nil
}

// DeleteTask deletes a task by ID
func DeleteTask(idStr string) error {
	todoFile, _, localID, _, err := loadTaskFile(idStr)
	if err != nil {
		fmt.Fprintln(os.Stderr, red(err.Error()))
		os.Exit(1)
	}

	deleted, err := PerformDeleteTask(todoFile, localID)
	if err != nil {
		fmt.Fprintln(os.Stderr, red(err.Error()))
		os.Exit(1)
	}

	fmt.Print(yellow("Deleted: "))
	fmt.Println(deleted.Text)
	return nil
}

// EditTask edits a task's text
func EditTask(idStr, newText string) error {
	todoFile, _, localID, _, err := loadTaskFile(idStr)
	if err != nil {
		fmt.Fprintln(os.Stderr, red(err.Error()))
		os.Exit(1)
	}

	updated, err := PerformEditTask(todoFile, localID, newText)
	if err != nil {
		fmt.Fprintf(os.Stderr, red("Task %s not found\n"), idStr)
		os.Exit(1)
	}

	fmt.Print(green("Updated: "))
	fmt.Println(formatTaskLine(updated, "", 0))
	return nil
}

// AddTaskNote adds a note to a task
func AddTaskNote(idStr, noteText string) error {
	todoFile, _, localID, _, err := loadTaskFile(idStr)
	if err != nil {
		fmt.Fprintln(os.Stderr, red(err.Error()))
		os.Exit(1)
	}

	updated, err := PerformAddTaskNote(todoFile, localID, noteText)
	if err != nil {
		fmt.Fprintln(os.Stderr, red(err.Error()))
		os.Exit(1)
	}

	fmt.Print(green("Added note to: "))
	fmt.Println(formatTaskLine(updated, "", 0))
	return nil
}

// ToggleTask toggles a task's status
func ToggleTask(idStr string) error {
	todoFile, task, localID, _, err := loadTaskFile(idStr)
	if err != nil {
		fmt.Fprintln(os.Stderr, red(err.Error()))
		os.Exit(1)
	}

	prevStatus := task.Status

	updated, err := PerformToggleTask(todoFile, localID)
	if err != nil {
		fmt.Fprintf(os.Stderr, red("Task %s not found\n"), idStr)
		os.Exit(1)
	}

	action := "Completed"
	if prevStatus == core.TaskCompleted {
		action = "Reopened"
	}

	fmt.Print(green(action + ": "))
	fmt.Println(formatTaskLine(updated, "", 0))
	return nil
}

// CompleteTask marks a task as completed
func CompleteTask(idStr string) error {
	todoFile, _, localID, _, err := loadTaskFile(idStr)
	if err != nil {
		fmt.Fprintln(os.Stderr, red(err.Error()))
		os.Exit(1)
	}

	updated, err := PerformCompleteTask(todoFile, localID)
	if err != nil {
		fmt.Fprintf(os.Stderr, red("Task %s not found\n"), idStr)
		os.Exit(1)
	}

	fmt.Print(green("Completed: "))
	fmt.Println(formatTaskLine(updated, "", 0))
	return nil
}

// CompleteTasks completes multiple tasks at once
func CompleteTasks(idsStr []string) error {
	todoFile, _, err := LoadDefaultTodoFile()
	if err != nil {
		return err
	}

	completedTasks, failedIDs, err := PerformCompleteTasks(todoFile, idsStr)
	if err != nil {
		return err
	}

	if len(completedTasks) == 0 {
		fmt.Fprintln(os.Stderr, red("No tasks were completed"))
		if len(failedIDs) > 0 {
			fmt.Fprintf(os.Stderr, red("Failed to complete tasks: %v\n"), failedIDs)
		}
		os.Exit(1)
	}

	fmt.Println(green(fmt.Sprintf("Completed %d task(s):", len(completedTasks))))
	for _, task := range completedTasks {
		fmt.Println(formatTaskLine(&task, "", 0))
	}

	if len(failedIDs) > 0 {
		fmt.Fprintf(os.Stderr, yellow("\nWarning: Failed to complete some tasks: %v\n"), failedIDs)
	}

	return nil
}

// AddNote adds a note to the Notes section
func AddNote(text string) error {
	todoFile, filePath, err := LoadDefaultTodoFile()
	if err != nil {
		return err
	}

	if _, err := PerformAddNote(todoFile, filePath, text); err != nil {
		return err
	}

	fmt.Print(green("Added note: "))
	fmt.Println(text)
	return nil
}

// FindTasks finds tasks by keyword
func FindTasks(keyword string, recursive bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	files, err := core.FindTodoFiles(cwd, recursive)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		fmt.Println("No TODO files found")
		return nil
	}

	found := false
	fmt.Println(bold(cyan(fmt.Sprintf("Searching for: \"%s\"", keyword))))
	fmt.Println()

	for _, file := range files {
		todoFile, err := core.Load(file.Path)
		if err != nil {
			continue
		}

		matches := todoFile.FindTasks(keyword)
		if len(matches) > 0 {
			found = true
			if recursive {
				fmt.Println(file.RelativePath)
			}
			for _, task := range matches {
				fmt.Println(formatTaskLine(&task, "", 0))
			}
		}
	}

	if !found {
		fmt.Println("No matching tasks found")
	}

	return nil
}

// OpenInEditor opens the TODO file in the configured editor
func OpenInEditor() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	filePath, err := core.FindDefaultTodoFile(cwd)
	if err != nil {
		return err
	}

	settings := config.GetSettings()

	var command string
	var args []string

	switch settings.Editor {
	case config.EditorVim:
		command = "vim"
		args = []string{filePath}
	case config.EditorNano:
		command = "nano"
		args = []string{filePath}
	case config.EditorDefaultApp:
		command = "open"
		args = []string{filePath}
	default:
		// System default
		command = os.Getenv("EDITOR")
		if command == "" {
			command = os.Getenv("VISUAL")
		}
		if command == "" {
			command = "vim"
		}
		args = []string{filePath}
	}

	fmt.Println(fmt.Sprintf("Opening %s...", filePath))

	cmd := exec.Command(command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// DeleteCompletedTasks deletes all completed tasks
func DeleteCompletedTasks() error {
	todoFile, _, err := LoadDefaultTodoFile()
	if err != nil {
		return err
	}

	count, err := PerformDeleteCompletedTasks(todoFile)
	if err != nil {
		return err
	}

	if count == 0 {
		fmt.Println("No completed tasks to delete")
		return nil
	}

	suffix := "s"
	if count == 1 {
		suffix = ""
	}
	fmt.Println(green(fmt.Sprintf("Deleted %d completed task%s", count, suffix)))
	return nil
}

// SetTaskIDs tags every task in the file with sequential IDs like ABC01, ABC02, ...
func SetTaskIDs(prefix string) error {
	todoFile, _, err := LoadDefaultTodoFile()
	if err != nil {
		return err
	}

	count, err := PerformSetTaskIDs(todoFile, prefix)
	if err != nil {
		fmt.Fprintln(os.Stderr, red(err.Error()))
		os.Exit(1)
	}

	suffix := "s"
	if count == 1 {
		suffix = ""
	}
	fmt.Println(green(fmt.Sprintf("Tagged %d task%s with IDs", count, suffix)))
	return nil
}

// RemoveTaskIDs strips ID tags from every task in the file
func RemoveTaskIDs() error {
	todoFile, _, err := LoadDefaultTodoFile()
	if err != nil {
		return err
	}

	changed, err := PerformRemoveTaskIDs(todoFile)
	if err != nil {
		return err
	}

	if !changed {
		fmt.Println("No task IDs to remove")
		return nil
	}

	fmt.Println(yellow("Removed task IDs"))
	return nil
}

// LintFile lints and fixes the TODO file
func LintFile() error {
	todoFile, filePath, err := LoadDefaultTodoFile()
	if err != nil {
		return err
	}

	fmt.Println(fmt.Sprintf("Linting %s...", filePath))
	fmt.Println()

	tasks := todoFile.GetTasks()
	pendingCount := 0
	completedCount := 0
	for _, t := range tasks {
		if t.Status == core.TaskPending {
			pendingCount++
		} else {
			completedCount++
		}
	}

	result, err := PerformLint(todoFile)
	if err != nil {
		return err
	}

	if len(result.Issues) == 0 {
		fmt.Println(green("✓ No issues found"))
		fmt.Println()
		suffix := "s"
		if len(tasks) == 1 {
			suffix = ""
		}
		fmt.Println(fmt.Sprintf("Checked %d task%s (%d pending, %d completed)", len(tasks), suffix, pendingCount, completedCount))
		return nil
	}

	fmt.Println(bold(cyan("Lint results:")))
	fmt.Println()

	for _, issue := range result.Issues {
		status := yellow("found")
		if issue.Fixed {
			status = green("fixed")
		}
		fmt.Printf("  Line %d: %s [%s]\n", issue.Line, issue.Issue, status)
	}

	fmt.Println()

	if result.FixedCount > 0 {
		suffix := "s"
		if result.FixedCount == 1 {
			suffix = ""
		}
		fmt.Println(green(fmt.Sprintf("✓ Fixed %d issue%s", result.FixedCount, suffix)))
	}

	suffix := "s"
	if len(tasks) == 1 {
		suffix = ""
	}
	fmt.Println(fmt.Sprintf("Checked %d task%s (%d pending, %d completed)", len(tasks), suffix, pendingCount, completedCount))
	return nil
}

// ShowVersion prints the version
func ShowVersion() {
	fmt.Printf("markdown-do v%s\n", Version)
}

// ShowHelp prints the help message
func ShowHelp() {
	fmt.Printf(`
%s - MarkdownDO: Manage TODO.md files

%s
  mdd              Open interactive TUI
  mdd <task text>  Add a new task (quotes optional)

%s
  mdd -l             List tasks
  mdd -ls            List tasks recursively (subdirectories)
  mdd -f <keyword>   Find tasks by keyword
  mdd -fs <keyword>  Find tasks by keyword, recursively

%s
  mdd -t <id>             Toggle task status (pending <-> completed)
  mdd -c <id>             Complete task by ID
  mdd -cm <id1> <id2>...  Complete multiple tasks by ID
  mdd -e <id> <text>      Edit task text
  mdd -an <id> <text>     Add a note to a task (shown inline wherever it's listed)
  mdd -d <id>             Delete task by ID
  mdd -dc                 Delete all completed tasks

%s
  mdd -n <text>      Add a note to the ## Notes section
  mdd -o             Open TODO file in editor
  mdd -lint          Lint and fix TODO file formatting
  mdd -id <PREFIX>   Tag every task with sequential IDs (PREFIX01, PREFIX02, ...)
  mdd -id -r         Remove all task ID tags
  mdd -v, --version  Show version
  mdd -h, --help     Show this help

%s
  End a task with "@Section" to file it there (created automatically
  if it doesn't exist yet). Built-in shortcuts, case-insensitive:

    @ff  -> Features        @ii  -> Ideas
    @bb  -> Bugs            @ww  -> Warnings

  Any other @name creates/uses a custom section, e.g. @Admin.

%s
  <id> above accepts either a task's position number, or its ID tag
  once IDs have been assigned with -id (e.g. "mdd -c ABC01").

%s
  mdd Buy groceries          Add a task to the inbox
  mdd "Fix login bug @bb"    Add a task to the Bugs section
  mdd "Dark mode @Features"  Add a task to the Features section
  mdd -l                     Show all tasks
  mdd -t 2                   Toggle task #2
  mdd -c 1                   Complete task #1
  mdd -cm 1 2 3              Complete tasks #1, #2, and #3
  mdd -e 1 Fix the bug       Edit task #1
  mdd -an 1 Needs a review   Add a note to task #1
  mdd -d 3                   Delete task #3
  mdd -f bug                 Find tasks containing "bug"
  mdd -id ABC                Tag all tasks: [ABC01], [ABC02], ...
  mdd -c ABC01               Complete the task tagged ABC01
  mdd -id -r                 Remove all ID tags
`, bold(cyan("mdd")), bold("Usage:"), bold("Viewing & Finding:"), bold("Managing Tasks:"), bold("Utilities:"), bold("Sections:"), bold("Task IDs:"), bold("Examples:"))
}
