package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/i-am-fran/markdown-do/internal/config"
	"github.com/i-am-fran/markdown-do/internal/core"
)

const Version = "3.2.2"

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
	case core.TaskInProgress:
		checkbox = cyan("[/]")
		text = taskText
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

// matchesStatusFilter reports whether status passes the given list filter
// ("active", "done", or "" for no filter).
func matchesStatusFilter(status core.TaskStatus, filter string) bool {
	switch filter {
	case "active":
		return status != core.TaskCompleted
	case "done":
		return status == core.TaskCompleted
	default:
		return true
	}
}

// ListTasks lists all tasks, optionally filtered by status ("active" or
// "done"; "" shows everything).
func ListTasks(recursive bool, statusFilter string) error {
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

				var tasks []core.Task
				for _, t := range todoFile.GetTasks() {
					if matchesStatusFilter(t.Status, statusFilter) {
						tasks = append(tasks, t)
					}
				}
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

		printedAny := false
		for _, group := range groups {
			var tasks []core.Task
			for _, t := range group.Tasks {
				if matchesStatusFilter(t.Status, statusFilter) {
					tasks = append(tasks, t)
				}
			}
			if len(tasks) == 0 {
				continue
			}

			if group.Section == nil {
				fmt.Println(bold(cyan("Tasks:")))
			} else {
				fmt.Println()
				fmt.Println(bold(magenta("## " + *group.Section)))
			}

			for _, task := range tasks {
				fmt.Println(formatTaskLine(&task, "", 0))
			}
			printedAny = true
		}

		if !printedAny {
			fmt.Println("No tasks found")
		}
	}

	return nil
}

// AddTask adds a new task. If path is non-empty, the task is added to the
// default TODO file in that directory instead of the current one.
func AddTask(text string, path string) error {
	var todoFile *core.TodoFile
	var filePath string
	var err error

	if path != "" {
		if !filepath.IsAbs(path) {
			cwd, cwdErr := os.Getwd()
			if cwdErr != nil {
				return cwdErr
			}
			path = filepath.Join(cwd, path)
		}
		todoFile, filePath, err = LoadTodoFileFromDir(path)
	} else {
		todoFile, filePath, err = LoadDefaultTodoFile()
	}
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

// dieOnErr prints err's own message and exits(1) — used where err.Error()
// is already the most useful message (load failures, validation errors).
func dieOnErr(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, red(err.Error()))
		os.Exit(1)
	}
}

// dieIfTaskNotFound prints a generic "Task %s not found" message and
// exits(1) — used where the underlying error is a low-information
// "task %d not found" that reads oddly against a string/stable ID.
func dieIfTaskNotFound(idStr string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, red("Task %s not found\n"), idStr)
		os.Exit(1)
	}
}

// mustLoadTaskFile collapses the load-task-file-or-die pattern shared by
// every single-ID CLI command.
func mustLoadTaskFile(idStr string) (*core.TodoFile, *core.Task, int) {
	todoFile, task, localID, _, err := loadTaskFile(idStr)
	dieOnErr(err)
	return todoFile, task, localID
}

// confirmPrompt shows msg and reads a y/n answer from stdin. It returns true
// immediately, without prompting, when confirmation isn't required (the
// confirmDestructive setting is off) or the caller already passed -y/--yes.
func confirmPrompt(msg string, yes bool) bool {
	if yes || !config.GetSettings().ConfirmDestructive {
		return true
	}

	fmt.Printf("%s [y/N] ", msg)
	var answer string
	fmt.Scanln(&answer)
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes"
}

// DeleteTask deletes a task by ID
func DeleteTask(idStr string, yes bool) error {
	todoFile, task, localID := mustLoadTaskFile(idStr)

	if !confirmPrompt(fmt.Sprintf("Delete task: %s?", task.Text), yes) {
		fmt.Println("Cancelled")
		return nil
	}

	deleted, err := PerformDeleteTask(todoFile, localID)
	dieOnErr(err)

	fmt.Print(yellow("Deleted: "))
	fmt.Println(deleted.Text)
	return nil
}

// EditTask edits a task's text
func EditTask(idStr, newText string) error {
	todoFile, _, localID := mustLoadTaskFile(idStr)

	updated, err := PerformEditTask(todoFile, localID, newText)
	dieIfTaskNotFound(idStr, err)

	fmt.Print(green("Updated: "))
	fmt.Println(formatTaskLine(updated, "", 0))
	return nil
}

// AddTaskNote adds a note to a task
func AddTaskNote(idStr, noteText string) error {
	todoFile, _, localID := mustLoadTaskFile(idStr)

	updated, err := PerformAddTaskNote(todoFile, localID, noteText)
	dieOnErr(err)

	fmt.Print(green("Added note to: "))
	fmt.Println(formatTaskLine(updated, "", 0))
	return nil
}

// ToggleTask toggles a task's status
func ToggleTask(idStr string) error {
	todoFile, _, localID := mustLoadTaskFile(idStr)

	updated, err := PerformToggleTask(todoFile, localID)
	dieIfTaskNotFound(idStr, err)

	action := "Reopened"
	switch updated.Status {
	case core.TaskCompleted:
		action = "Completed"
	case core.TaskInProgress:
		action = "Started"
	}

	fmt.Print(green(action + ": "))
	fmt.Println(formatTaskLine(updated, "", 0))
	return nil
}

// CompleteTask marks one or more tasks as completed, by position or stable
// ID. IDs that resolve against the local default TODO file are completed
// together through PerformCompleteTasks, which saves once for the whole
// batch — completing one local ID must not reorder the file (and shift
// everyone else's positional IDs) before the rest of the batch is resolved.
// IDs only found via the recursive-listing task cache (i.e. belonging to a
// different file) fall back to the old one-file-load-and-save-per-ID path,
// since those files don't share position numbering with each other.
func CompleteTask(idsStr []string) error {
	todoFile, _, err := LoadDefaultTodoFile()
	if err != nil {
		return err
	}

	var localIDs, remoteIDs []string
	for _, idStr := range idsStr {
		if n, convErr := strconv.Atoi(idStr); convErr == nil {
			if todoFile.GetTask(n) != nil {
				localIDs = append(localIDs, idStr)
			} else {
				remoteIDs = append(remoteIDs, idStr)
			}
		} else if todoFile.GetTaskByStableID(idStr) != nil {
			localIDs = append(localIDs, idStr)
		} else {
			remoteIDs = append(remoteIDs, idStr)
		}
	}

	var completed []core.Task
	var failedIDs []string

	if len(localIDs) > 0 {
		localCompleted, localFailed, err := PerformCompleteTasks(todoFile, localIDs)
		if err != nil {
			return err
		}
		completed = append(completed, localCompleted...)
		failedIDs = append(failedIDs, localFailed...)
	}

	for _, idStr := range remoteIDs {
		remoteTodoFile, _, localID, _, err := loadTaskFile(idStr)
		if err != nil {
			failedIDs = append(failedIDs, idStr)
			continue
		}

		updated, err := PerformCompleteTask(remoteTodoFile, localID)
		if err != nil {
			failedIDs = append(failedIDs, idStr)
			continue
		}

		completed = append(completed, *updated)
	}

	if len(idsStr) == 1 {
		dieIfTaskNotFound(idsStr[0], errIfEmpty(completed))
		fmt.Print(green("Completed: "))
		fmt.Println(formatTaskLine(&completed[0], "", 0))
		return nil
	}

	if len(completed) == 0 {
		fmt.Fprintln(os.Stderr, red("No tasks were completed"))
		if len(failedIDs) > 0 {
			fmt.Fprintf(os.Stderr, red("Failed to complete tasks: %v\n"), failedIDs)
		}
		os.Exit(1)
	}

	fmt.Println(green(fmt.Sprintf("Completed %d task(s):", len(completed))))
	for _, task := range completed {
		fmt.Println(formatTaskLine(&task, "", 0))
	}

	if len(failedIDs) > 0 {
		fmt.Fprintf(os.Stderr, yellow("\nWarning: Failed to complete some tasks: %v\n"), failedIDs)
	}

	return nil
}

// errIfEmpty returns a non-nil error when tasks is empty, for reuse with
// dieIfTaskNotFound's error-presence check.
func errIfEmpty(tasks []core.Task) error {
	if len(tasks) == 0 {
		return fmt.Errorf("no tasks completed")
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

	return openPathInEditor(filePath)
}

// ConfigEdit opens the settings file (config.json) directly in the
// configured editor.
func ConfigEdit() error {
	filePath, err := config.ConfigFilePath()
	if err != nil {
		return err
	}

	return openPathInEditor(filePath)
}

// openPathInEditor opens filePath in the user's configured editor.
func openPathInEditor(filePath string) error {
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
func DeleteCompletedTasks(yes bool) error {
	todoFile, _, err := LoadDefaultTodoFile()
	if err != nil {
		return err
	}

	pendingCount := len(todoFile.GetCompletedTasks())
	if pendingCount == 0 {
		fmt.Println("No completed tasks to delete")
		return nil
	}

	suffix := "s"
	if pendingCount == 1 {
		suffix = ""
	}
	if !confirmPrompt(fmt.Sprintf("Delete %d completed task%s?", pendingCount, suffix), yes) {
		fmt.Println("Cancelled")
		return nil
	}

	count, err := PerformDeleteCompletedTasks(todoFile)
	if err != nil {
		return err
	}

	fmt.Println(green(fmt.Sprintf("Deleted %d completed task%s", count, suffix)))
	return nil
}

// ArchiveCompletedTasks moves all completed tasks into the Archive section
func ArchiveCompletedTasks() error {
	todoFile, _, err := LoadDefaultTodoFile()
	if err != nil {
		return err
	}

	count, err := PerformArchiveCompletedTasks(todoFile)
	if err != nil {
		return err
	}

	if count == 0 {
		fmt.Println("No completed tasks to archive")
		return nil
	}

	suffix := "s"
	if count == 1 {
		suffix = ""
	}
	fmt.Println(green(fmt.Sprintf("Archived %d completed task%s", count, suffix)))
	return nil
}

// UndoLastChange reverts the last change made to the TODO file. Because the
// restore is written through the normal save path, running it again
// re-applies the change it just undid (toggle behavior).
func UndoLastChange() error {
	todoFile, filePath, err := LoadDefaultTodoFile()
	if err != nil {
		return err
	}

	if err := PerformUndo(todoFile, filePath); err != nil {
		if err == ErrNothingToUndo {
			fmt.Println("Nothing to undo")
			return nil
		}
		return err
	}

	fmt.Println(green("Undid last change (run \"mdd undo\" again to redo it)"))
	return nil
}

// latestReleaseTag fetches the tag name of the latest GitHub release for
// this repo. "go install pkg@latest" can't be used for this: this repo's
// release tags are bare semver (e.g. "3.2.0"), and Go's module tooling only
// recognizes "v"-prefixed tags (vX.Y.Z) as valid versions, so "@latest"
// silently resolves to the one legacy "v1.0.0" tag instead of the real
// latest release.
func latestReleaseTag() (string, error) {
	resp, err := http.Get("https://api.github.com/repos/i-am-fran/markdown-do/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned %s", resp.Status)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}
	if release.TagName == "" {
		return "", errors.New("GitHub API response had no tag_name")
	}
	return release.TagName, nil
}

// UpdateBinary installs the latest release of mdd via "go install", pinned
// to the exact tag returned by the GitHub API (see latestReleaseTag).
func UpdateBinary() error {
	if _, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("go is not installed or not on PATH — download a binary from " +
			"https://github.com/i-am-fran/markdown-do/releases or install Go and retry")
	}

	tag, err := latestReleaseTag()
	if err != nil {
		return fmt.Errorf("could not determine the latest release: %w", err)
	}

	target := fmt.Sprintf("github.com/i-am-fran/markdown-do/cmd/mdd@%s", tag)
	fmt.Println("Running: go install " + target)

	cmd := exec.Command("go", "install", target)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	fmt.Println(green(fmt.Sprintf("Updated to %s. This process is still v%s — run \"mdd version\" to confirm the new one.", tag, Version)))
	return nil
}

// ConfigList prints every setting, one per line.
func ConfigList() error {
	s := config.GetSettings()
	fmt.Printf("showCompleted: %t\n", s.ShowCompleted)
	fmt.Printf("editor: %s\n", s.Editor)
	fmt.Printf("enableInProgress: %t\n", s.EnableInProgress)
	fmt.Printf("confirmDestructive: %t\n", s.ConfirmDestructive)
	for name, section := range s.SectionAliases {
		fmt.Printf("alias.%s: %s\n", name, section)
	}
	return nil
}

// ConfigGet prints a single setting's value.
func ConfigGet(key string) error {
	s := config.GetSettings()
	switch {
	case key == "showCompleted":
		fmt.Println(s.ShowCompleted)
	case key == "editor":
		fmt.Println(s.Editor)
	case key == "enableInProgress":
		fmt.Println(s.EnableInProgress)
	case key == "confirmDestructive":
		fmt.Println(s.ConfirmDestructive)
	case strings.HasPrefix(key, "alias."):
		name := strings.TrimPrefix(key, "alias.")
		v, ok := s.SectionAliases[name]
		if !ok {
			return fmt.Errorf("no alias set for %q", name)
		}
		fmt.Println(v)
	default:
		return fmt.Errorf("unknown setting %q", key)
	}
	return nil
}

// ConfigSet updates a single setting and persists it.
func ConfigSet(key, value string) error {
	switch {
	case key == "editor":
		valid := false
		for _, opt := range config.EditorOptions() {
			if string(opt.Value) == value {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid editor %q — options: system, vim, nano, default-app", value)
		}
		dieOnErr(config.UpdateSettings(map[string]interface{}{"editor": config.EditorOption(value)}))

	case key == "showCompleted" || key == "enableInProgress" || key == "confirmDestructive":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("%q must be true or false", key)
		}
		dieOnErr(config.UpdateSettings(map[string]interface{}{key: b}))

	case strings.HasPrefix(key, "alias."):
		name := strings.ToLower(strings.TrimPrefix(key, "alias."))
		if name == "" {
			return fmt.Errorf("alias name required, e.g. mdd config set alias.xx Features")
		}
		dieOnErr(config.UpdateSettings(map[string]interface{}{"sectionAlias": [2]string{name, value}}))

	default:
		return fmt.Errorf("unknown setting %q", key)
	}

	fmt.Printf("%s: %s\n", key, value)
	return nil
}

// SetTaskIDs tags every task in the file with sequential IDs like ABC-001, ABC-002, ...
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
	inProgressCount := 0
	completedCount := 0
	for _, t := range tasks {
		switch t.Status {
		case core.TaskCompleted:
			completedCount++
		case core.TaskInProgress:
			inProgressCount++
		default:
			pendingCount++
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
		if inProgressCount > 0 {
			fmt.Println(fmt.Sprintf("Checked %d task%s (%d pending, %d in progress, %d completed)", len(tasks), suffix, pendingCount, inProgressCount, completedCount))
		} else {
			fmt.Println(fmt.Sprintf("Checked %d task%s (%d pending, %d completed)", len(tasks), suffix, pendingCount, completedCount))
		}
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
%s - Markdown-do: Manage TODO.md files

%s
  mdd              Show this help
  mdd <task text>  Add a new task (quotes optional)
  mdd add <text>   Add a new task, explicitly (see "A note on task text" below)
  mdd add <text> --path <dir>  Add to the TODO file in <dir> instead of cwd

%s
  mdd list             List tasks
  mdd list -r          List tasks recursively (subdirectories)
  mdd list -a          List only active tasks (pending + in-progress)
  mdd list --done      List only completed tasks
  mdd find <keyword>   Find tasks by keyword
  mdd find <kw> -r     Find tasks by keyword, recursively

%s
  mdd toggle <id>             Toggle task status (pending <-> completed)
  mdd complete <id> [id2...]  Complete one or more tasks by ID
  mdd edit <id> <text>        Edit task text
  mdd annotate <id> <text>    Add a note to a task (shown inline wherever it's listed)
  mdd remove <id>             Delete task by ID
  mdd clear                   Delete all completed tasks
                              (both prompt first if confirmDestructive is on;
                               add -y/--yes to skip the prompt)
  mdd archive                 Move all completed tasks to the ## Archive
                              section, noting where each came from
  mdd undo                    Revert the last change (run again to redo it)

%s
  mdd notes <text>   Add a note to the ## Notes section
  mdd open           Open TODO file in editor
  mdd lint           Lint and fix TODO file formatting
  mdd tag <PREFIX>   Tag every task with sequential IDs (PREFIX01, PREFIX02, ...)
  mdd untag          Remove all task ID tags
  mdd config list    Show all settings
  mdd config get k   Show one setting, e.g. mdd config get editor
  mdd config set k v Change a setting, e.g. mdd config set editor vim
  mdd config edit    Open config.json directly in your editor
  mdd completion bash|zsh  Print a shell completion script (see below)
  mdd update         Install the latest release (via "go install")
  mdd version        Show version (also: -v, --version)
  mdd help           Show this help (also: -h, --help)

%s
  End a task with "@Section" to file it there (created automatically
  if it doesn't exist yet). Built-in shortcuts, case-insensitive:

    @ff  -> Features        @ii  -> Ideas
    @bb  -> Bugs            @ww  -> Warnings

  Any other @name creates/uses a custom section, e.g. @Admin.

%s
  Prefix a character with \ to keep it literal instead of special, e.g.
  "mdd 'Meet Bob \@bb'" adds a task ending in a literal @bb instead of
  filing it under Bugs. Use \\ for a literal backslash.

  Some shells treat unquoted ?/* as glob patterns and will error ("no
  matches found") instead of passing them to mdd — quote task text that
  contains them, e.g. mdd "Ping the vendor?".

%s
  <id> above accepts either a task's position number, or its ID tag
  once IDs have been assigned with tag (e.g. "mdd complete ABC-001").

%s
  Tab-complete commands (e.g. "mdd ar" -> "mdd archive") and section
  tags (e.g. "@Bu" -> "@Bugs"). Add one line to your shell rc file:

    source <(mdd completion bash)   # ~/.bashrc
    source <(mdd completion zsh)    # ~/.zshrc

%s
  A word like "complete" or "list" is only treated as a command when
  what follows it looks like that command's arguments (e.g. "complete"
  needs a task ID next). Otherwise it's added as a task, so
  "mdd Complete the tax return" still just adds a task. If your task
  text is ever misread as a command, use "mdd add" to force it:

    mdd add "Find a good plumber"

%s
  mdd Buy groceries            Add a task to the inbox
  mdd "Fix login bug @bb"      Add a task to the Bugs section
  mdd "Dark mode @Features"    Add a task to the Features section
  mdd list                     Show all tasks
  mdd toggle 2                 Toggle task #2
  mdd complete 1               Complete task #1
  mdd complete 1 2 3           Complete tasks #1, #2, and #3
  mdd edit 1 Fix the bug       Edit task #1
  mdd annotate 1 Needs review  Add a note to task #1
  mdd remove 3                 Delete task #3
  mdd archive                  Move all completed tasks to ## Archive
  mdd find bug                 Find tasks containing "bug"
  mdd tag ABC                  Tag all tasks: [ABC-001], [ABC-002], ...
  mdd complete ABC-001         Complete the task tagged ABC-001
  mdd untag                    Remove all ID tags
`, bold(cyan("mdd")), bold("Usage:"), bold("Viewing & Finding:"), bold("Managing Tasks:"), bold("Utilities:"), bold("Sections:"), bold("Escaping & quoting:"), bold("Task IDs:"), bold("Shell Completion:"), bold("A note on task text:"), bold("Examples:"))
}
