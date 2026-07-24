package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/i-am-fran/markdowndo/internal/cli"
	"github.com/i-am-fran/markdowndo/internal/tui"
)

func main() {
	// No arguments at all -> TUI
	if len(os.Args) == 1 {
		p := tea.NewProgram(tui.New(), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	args := cli.ParseArgs(os.Args[1:])

	// If nothing but flags were passed (no command), fall back to help.
	if args.Command == "" {
		cli.ShowHelp()
		return
	}

	if err := handleCLI(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// usageError prints a usage message for a malformed command and exits(1).
func usageError(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", a...)
	os.Exit(1)
}

func handleCLI(args cli.ParsedArgs) error {
	switch args.Command {
	case "help":
		cli.ShowHelp()
		return nil

	case "version":
		cli.ShowVersion()
		return nil

	case "list":
		return cli.ListTasks(args.Recursive)

	case "find":
		if len(args.Args) == 0 {
			usageError("find requires a keyword, e.g. mdd find bug")
		}
		return cli.FindTasks(args.Text, args.Recursive)

	case "toggle":
		if len(args.Args) != 1 {
			usageError("toggle takes a single task id, e.g. mdd toggle 2")
		}
		return cli.ToggleTask(args.Args[0])

	case "complete":
		return cli.CompleteTask(args.Args)

	case "edit":
		if len(args.Args) < 2 {
			usageError("edit requires an id and new text, e.g. mdd edit 1 Fix the bug")
		}
		return cli.EditTask(args.Args[0], strings.Join(args.Args[1:], " "))

	case "annotate":
		if len(args.Args) < 2 {
			usageError("annotate requires an id and note text, e.g. mdd annotate 1 Needs review")
		}
		return cli.AddTaskNote(args.Args[0], strings.Join(args.Args[1:], " "))

	case "remove":
		if len(args.Args) != 1 {
			usageError("remove takes a single task id, e.g. mdd remove 3")
		}
		return cli.DeleteTask(args.Args[0])

	case "clear":
		return cli.DeleteCompletedTasks()

	case "notes":
		if len(args.Args) == 0 {
			usageError("notes requires text, e.g. mdd notes check pantry")
		}
		return cli.AddNote(args.Text)

	case "open":
		return cli.OpenInEditor()

	case "lint":
		return cli.LintFile()

	case "tag":
		if len(args.Args) != 1 {
			usageError("tag requires a single 3-letter prefix, e.g. mdd tag ABC")
		}
		return cli.SetTaskIDs(args.Args[0])

	case "untag":
		return cli.RemoveTaskIDs()

	case "add":
		if args.Text == "" {
			usageError("add requires task text, e.g. mdd add Buy groceries")
		}
		return cli.AddTask(args.Text)

	case "unknown":
		usageError("unrecognized flag %q — run 'mdd help' for the current commands, or 'mdd add \"...\"' to add a task starting with a dash", args.Args[0])
		return nil

	default:
		cli.ShowHelp()
		return nil
	}
}
