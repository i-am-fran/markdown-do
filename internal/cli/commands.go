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

const Version = "1.0.0"

var (
	green     = color.New(color.FgGreen).SprintFunc()
	yellow    = color.New(color.FgYellow).SprintFunc()
	red       = color.New(color.FgRed).SprintFunc()
	cyan      = color.New(color.FgCyan).SprintFunc()
	magenta   = color.New(color.FgMagenta).SprintFunc()
	bold      = color.New(color.Bold).SprintFunc()
	dim       = color.New(color.Faint).SprintFunc()
	strikethrough = color.New(color.CrossedOut, color.Faint).SprintFunc()
)

func formatTaskLine(task *core.Task, showFile string, displayID int) string {
	var checkbox, text string

	id := displayID
	if id == 0 {
		id = task.ID
	}

	switch task.Status {
	case core.TaskCompleted:
		checkbox = green("[x]")
		text = strikethrough(task.Text)
	default:
		checkbox = dim("[ ]")
		text = task.Text
	}

	idStr := dim(fmt.Sprintf("%d.", id))
	file := ""
	if showFile != "" {
		file = dim(fmt.Sprintf(" (%s)", showFile))
	}

	return fmt.Sprintf("  %s %s %s%s", idStr, checkbox, text, file)
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
			fmt.Println(dim("No TODO files found"))
			return nil
		}

		globalTaskID := 1

		for dir, files := range filesByDir {
			fmt.Println(bold(cyan(dir)))

			for _, file := range files {
				todoFile, err := core.Load(file.Path)
				if err != nil {
					continue
				}

				tasks := todoFile.GetTasks()
				if len(tasks) == 0 {
					fmt.Println(dim("  (no tasks)"))
				} else {
					for _, task := range tasks {
						fmt.Println(formatTaskLine(&task, "", globalTaskID))
						globalTaskID++
					}
				}
			}
			fmt.Println()
		}
	} else {
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
			fmt.Println(dim("No tasks found"))
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
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	filePath, err := core.FindDefaultTodoFile(cwd)
	if err != nil {
		return err
	}

	todoFile, err := core.Load(filePath)
	if err != nil {
		return err
	}

	// Create file if it doesn't exist
	if len(todoFile.GetTasks()) == 0 && todoFile.Serialize() == "" {
		todoFile, err = core.Create(filePath)
		if err != nil {
			return err
		}
	}

	task, err := todoFile.AddTask(text)
	if err != nil {
		return err
	}

	if err := todoFile.Save(); err != nil {
		return err
	}

	fmt.Print(green("Added: "))
	fmt.Println(formatTaskLine(task, "", 0))
	return nil
}

// DeleteTask deletes a task by ID
func DeleteTask(idStr string) error {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Fprintln(os.Stderr, red("Invalid task ID"))
		os.Exit(1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	filePath, err := core.FindDefaultTodoFile(cwd)
	if err != nil {
		return err
	}

	todoFile, err := core.Load(filePath)
	if err != nil {
		return err
	}

	task := todoFile.GetTask(id)
	if task == nil {
		fmt.Fprintf(os.Stderr, red("Task %d not found\n"), id)
		os.Exit(1)
	}

	todoFile.DeleteTask(id)
	if err := todoFile.Save(); err != nil {
		return err
	}

	fmt.Print(yellow("Deleted: "))
	fmt.Println(task.Text)
	return nil
}

// EditTask edits a task's text
func EditTask(idStr, newText string) error {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Fprintln(os.Stderr, red("Invalid task ID"))
		os.Exit(1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	filePath, err := core.FindDefaultTodoFile(cwd)
	if err != nil {
		return err
	}

	todoFile, err := core.Load(filePath)
	if err != nil {
		return err
	}

	task := todoFile.GetTask(id)
	if task == nil {
		fmt.Fprintf(os.Stderr, red("Task %d not found\n"), id)
		os.Exit(1)
	}

	taskStatus := task.Status

	if !todoFile.UpdateTask(id, newText) {
		fmt.Fprintf(os.Stderr, red("Task %d not found\n"), id)
		os.Exit(1)
	}

	if err := todoFile.Save(); err != nil {
		return err
	}

	fmt.Print(green("Updated: "))
	fmt.Println(formatTaskLine(&core.Task{
		ID:     id,
		Text:   newText,
		Status: taskStatus,
	}, "", 0))
	return nil
}

// ToggleTask toggles a task's status
func ToggleTask(idStr string) error {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Fprintln(os.Stderr, red("Invalid task ID"))
		os.Exit(1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	filePath, err := core.FindDefaultTodoFile(cwd)
	if err != nil {
		return err
	}

	todoFile, err := core.Load(filePath)
	if err != nil {
		return err
	}

	task := todoFile.GetTask(id)
	if task == nil {
		fmt.Fprintf(os.Stderr, red("Task %d not found\n"), id)
		os.Exit(1)
	}

	taskText := task.Text
	prevStatus := task.Status

	if !todoFile.ToggleTask(id) {
		fmt.Fprintf(os.Stderr, red("Task %d not found\n"), id)
		os.Exit(1)
	}

	if err := todoFile.Save(); err != nil {
		return err
	}

	newStatus := core.TaskCompleted
	if prevStatus == core.TaskCompleted {
		newStatus = core.TaskPending
	}

	action := "Completed"
	if newStatus == core.TaskPending {
		action = "Reopened"
	}

	fmt.Print(green(action + ": "))
	fmt.Println(formatTaskLine(&core.Task{
		ID:     id,
		Text:   taskText,
		Status: newStatus,
	}, "", 0))
	return nil
}

// CompleteTask marks a task as completed
func CompleteTask(idStr string) error {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Fprintln(os.Stderr, red("Invalid task ID"))
		os.Exit(1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	filePath, err := core.FindDefaultTodoFile(cwd)
	if err != nil {
		return err
	}

	todoFile, err := core.Load(filePath)
	if err != nil {
		return err
	}

	task := todoFile.GetTask(id)
	if task == nil {
		fmt.Fprintf(os.Stderr, red("Task %d not found\n"), id)
		os.Exit(1)
	}

	taskText := task.Text

	if !todoFile.SetTaskStatus(id, core.TaskCompleted) {
		fmt.Fprintf(os.Stderr, red("Task %d not found\n"), id)
		os.Exit(1)
	}

	if err := todoFile.Save(); err != nil {
		return err
	}

	fmt.Print(green("Completed: "))
	fmt.Println(formatTaskLine(&core.Task{
		ID:     id,
		Text:   taskText,
		Status: core.TaskCompleted,
	}, "", 0))
	return nil
}

// AddNote adds a note to the Notes section
func AddNote(text string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	filePath, err := core.FindDefaultTodoFile(cwd)
	if err != nil {
		return err
	}

	todoFile, err := core.Load(filePath)
	if err != nil {
		return err
	}

	// Create file if it doesn't exist
	if len(todoFile.GetTasks()) == 0 && todoFile.Serialize() == "" {
		todoFile, err = core.Create(filePath)
		if err != nil {
			return err
		}
	}

	todoFile.AddNote(text)
	if err := todoFile.Save(); err != nil {
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
		fmt.Println(dim("No TODO files found"))
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
				fmt.Println(dim(file.RelativePath))
			}
			for _, task := range matches {
				fmt.Println(formatTaskLine(&task, "", 0))
			}
		}
	}

	if !found {
		fmt.Println(dim("No matching tasks found"))
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

	fmt.Println(dim(fmt.Sprintf("Opening %s...", filePath)))

	cmd := exec.Command(command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// DeleteCompletedTasks deletes all completed tasks
func DeleteCompletedTasks() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	filePath, err := core.FindDefaultTodoFile(cwd)
	if err != nil {
		return err
	}

	todoFile, err := core.Load(filePath)
	if err != nil {
		return err
	}

	count := todoFile.DeleteCompletedTasks()
	if count == 0 {
		fmt.Println(dim("No completed tasks to delete"))
		return nil
	}

	if err := todoFile.Save(); err != nil {
		return err
	}

	suffix := "s"
	if count == 1 {
		suffix = ""
	}
	fmt.Println(green(fmt.Sprintf("Deleted %d completed task%s", count, suffix)))
	return nil
}

// LintFile lints and fixes the TODO file
func LintFile() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	filePath, err := core.FindDefaultTodoFile(cwd)
	if err != nil {
		return err
	}

	todoFile, err := core.Load(filePath)
	if err != nil {
		return err
	}

	fmt.Println(dim(fmt.Sprintf("Linting %s...", filePath)))
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

	result := todoFile.Lint()

	if len(result.Issues) == 0 {
		fmt.Println(green("✓ No issues found"))
		fmt.Println()
		suffix := "s"
		if len(tasks) == 1 {
			suffix = ""
		}
		fmt.Println(dim(fmt.Sprintf("Checked %d task%s (%d pending, %d completed)", len(tasks), suffix, pendingCount, completedCount)))
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
		if err := todoFile.Save(); err != nil {
			return err
		}
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
	fmt.Println(dim(fmt.Sprintf("Checked %d task%s (%d pending, %d completed)", len(tasks), suffix, pendingCount, completedCount)))
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
  mdd                    Open interactive TUI
  mdd <task text>        Add a new task (quotes optional)
  mdd -l                 List tasks
  mdd -ls                List tasks recursively
  mdd -t <id>            Toggle task status
  mdd -c <id>            Complete task by ID
  mdd -e <id> <text>     Edit task text
  mdd -d <id>            Delete task by ID
  mdd -dc                Delete all completed tasks
  mdd -f <keyword>       Find tasks by keyword
  mdd -fs <keyword>      Find tasks recursively
  mdd -n <text>          Add a note to ## Notes section
  mdd -o                 Open TODO file in editor
  mdd -lint              Lint and fix TODO file formatting
  mdd -v, --version      Show version
  mdd -h, --help         Show this help

%s
  mdd Buy groceries      Add new task (no quotes needed)
  mdd -l                 Show all tasks
  mdd -t 2               Toggle task #2
  mdd -c 1               Complete task #1
  mdd -e 1 Fix the bug   Edit task #1
  mdd -d 3               Delete task #3
  mdd -f bug             Find tasks containing "bug"
`, bold(cyan("mdd")), bold("Usage:"), bold("Examples:"))
}
