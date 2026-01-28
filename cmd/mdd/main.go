package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/i-am-fran/markdowndo/internal/cli"
	"github.com/i-am-fran/markdowndo/internal/tui"
)

func main() {
	// Parse command line arguments
	args := cli.ParseArgs(os.Args[1:])

	// If no arguments or flags, start TUI
	if len(args.Flags) == 0 && args.Text == "" {
		p := tea.NewProgram(tui.New(), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Handle CLI commands
	if err := handleCLI(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleCLI(args cli.ParsedArgs) error {
	// Help flag
	if args.HasFlag("h") || args.HasFlag("help") {
		cli.ShowHelp()
		return nil
	}

	// Version flag
	if args.HasFlag("v") || args.HasFlag("version") {
		cli.ShowVersion()
		return nil
	}

	// List tasks
	if args.HasFlag("l") {
		return cli.ListTasks(false)
	}

	// List tasks recursively
	if args.HasFlag("ls") {
		return cli.ListTasks(true)
	}

	// Find tasks
	if args.HasFlag("f") {
		return cli.FindTasks(args.Value, false)
	}

	// Find tasks recursively
	if args.HasFlag("fs") {
		return cli.FindTasks(args.Value, true)
	}

	// Toggle task
	if args.HasFlag("t") {
		return cli.ToggleTask(args.Value)
	}

	// Complete task
	if args.HasFlag("c") {
		return cli.CompleteTask(args.Value)
	}

	// Complete multiple tasks
	if args.HasFlag("cm") {
		if len(args.Values) == 0 {
			fmt.Fprintln(os.Stderr, "Error: -cm requires at least one task ID")
			os.Exit(1)
		}
		return cli.CompleteTasks(args.Values)
	}

	// Edit task
	if args.HasFlag("e") {
		return cli.EditTask(args.Value, args.Text)
	}

	// Delete task
	if args.HasFlag("d") {
		return cli.DeleteTask(args.Value)
	}

	// Delete completed tasks
	if args.HasFlag("dc") {
		return cli.DeleteCompletedTasks()
	}

	// Add note
	if args.HasFlag("n") {
		return cli.AddNote(args.Value)
	}

	// Open in editor
	if args.HasFlag("o") {
		return cli.OpenInEditor()
	}

	// Lint
	if args.HasFlag("lint") {
		return cli.LintFile()
	}

	// If we get here, treat the text as a task to add
	if args.Text != "" {
		return cli.AddTask(args.Text)
	}

	// No valid command, show help
	cli.ShowHelp()
	return nil
}
