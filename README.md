# Markdown-do

A fast, minimal, opinionated CLI tool for managing TODO.md files using standard markdown syntax.

## Features

- **Plain markdown** - Tasks are standard `- [ ]` checkboxes, readable anywhere
- **Two task states** - Pending `[ ]` and completed `[x]`
- **Sections** - Organize tasks with `## Headers`
- **Inbox workflow** - New tasks go to the top unless you specify a section
- **Auto-sorting** - Tasks reorder by status on save (pending -> completed)
- **Fast CLI** - Add, toggle, and manage tasks without leaving your terminal
- **Recursive search** - Find tasks across all TODO files in subdirectories
- **Lint & fix** - Auto-fix common formatting issues

## Installation

### From releases

Download the latest binary for your platform from the [releases page](https://github.com/i-am-fran/markdown-do/releases).

### From source

```bash
git clone https://github.com/i-am-fran/markdown-do.git
cd markdown-do
make build
# Binary is at ./build/mdd

# Or install to GOPATH/bin:
make install
```

### Go install

```bash
go install github.com/i-am-fran/markdown-do/cmd/mdd@latest
```

## Quick Start

```bash
# Add your first task
mdd Buy milk

# List tasks
mdd list

# Complete task #1
mdd complete 1
```

## CLI Reference

### Adding Tasks

```bash
mdd Buy groceries              # Add task (quotes optional)
mdd "Fix bug in auth"          # Add task with quotes
mdd "New feature @Features"    # Add to specific section
mdd add "..."                  # Force add, if task text is ever misread as a command
mdd notes "Meeting notes"      # Add note to ## Notes section
```

### Viewing Tasks

```bash
mdd list                       # List tasks in current directory
mdd list -r                    # List tasks recursively (subdirectories)
mdd find bug                   # Find tasks containing "bug"
mdd find auth -r               # Find recursively
```

### Managing Tasks

```bash
mdd toggle 1                   # Toggle task #1 status
mdd complete 1                 # Mark task #1 as complete
mdd complete 1 2 3             # Mark tasks #1, #2, and #3 as complete
mdd edit 1 "Updated text"      # Edit task #1
mdd annotate 1 "Needs review"  # Add a note to task #1
mdd remove 1                   # Delete task #1
mdd clear                      # Delete all completed tasks
```

`remove` and `clear` prompt for confirmation first if the `confirmDestructive`
setting is on (skip with `-y`/`--yes`).

### Task IDs

Tag every task in a file with stable, sequential IDs so they're easy to
reference — useful when working with LLMs or across sessions where position
numbers can shift.

```bash
mdd tag ABC                    # Tag all tasks: [ABC-001], [ABC-002], ...
mdd complete ABC-001           # Reference a task by its stable ID
mdd untag                      # Remove all ID tags
```

Any command that takes `<id>` accepts either a task's position number or its
ID tag once IDs have been assigned.

### File Operations

```bash
mdd open                       # Open TODO.md in editor
mdd lint                       # Lint and auto-fix formatting
mdd config list                # Show all settings
mdd config get editor          # Show one setting
mdd config set editor vim      # Change a setting
mdd config edit                # Open config.json directly in your editor
mdd help                       # Show help (also -h, --help)
mdd version                    # Show version (also -v, --version)
```

A word like `complete` or `list` is only treated as a command when what
follows it matches that command's shape (e.g. `complete` needs a task ID
next); otherwise it's added as a task, so `mdd Complete the tax return`
still just adds a task. Use `mdd add "..."` to force add if your task text
is ever misread as a command.

Run `mdd` without arguments to show help.

## Task Format

Markdown-do uses standard markdown checkbox syntax, plus a non-standard third
state:

```markdown
# TODO

- [ ] Pending task
- [/] In-progress task (only reachable via toggle when enableInProgress is on)
- [x] Completed task
  - A note about the task
  - Another note
```

A task's trailing indented plain-bullet lines (no checkbox) are its notes —
added with `mdd annotate N "text"`, they always travel with the task through
edit/move/delete/complete.

### Status Cycle

Toggle (`mdd toggle N`) cycles through pending &lt;-&gt;
completed by default, or pending -&gt; in-progress -&gt; completed -&gt;
pending when the `enableInProgress` setting is on.

Use `mdd complete N` to mark a task as complete directly.

### Sections

Organize tasks under `## Header` sections:

```bash
mdd "New task @Backend"        # Adds to ## Backend (creates if needed)
mdd "Another task"             # Adds to inbox (top of file)
```

Section aliases for quick add: `@ff` -> Features, `@bb` -> Bugs, `@ii` -> Ideas, `@ww` -> Warnings (extend with custom ones via `mdd config set alias.<name> <Section>`)

### Auto-Sorting

On save, tasks within each section are automatically reordered:

1. Pending tasks
2. Completed tasks

This keeps your active work visible at the top.

## Configuration

Settings are stored in `~/.config/markdowndo/config.json`.

| Setting | Default | Description |
| ------- | ------- | ----------- |
| `showCompleted` | `true` | Show completed tasks in lists |
| `editor` | `"system"` | Editor for the `open` command |
| `enableInProgress` | `false` | Enables the `- [/]` in-progress checkbox state and 3-way toggle cycle |
| `confirmDestructive` | `false` | Require a y/n prompt before `remove`/`clear` (skip with `-y`/`--yes`) |
| `sectionAliases` | `{}` | Custom `@alias` shortcuts layered on top of the built-in ones |

### Editor Options

- `system` - Uses `$EDITOR` or `$VISUAL` environment variable
- `vim` - Open in Vim
- `nano` - Open in Nano
- `default-app` - Open with system default application

## File Discovery

Markdown-do looks for these files (in order):

1. `TODO.md` in current directory
2. `todo.md` in current directory
3. Creates `TODO.md` if none exists

For recursive operations (`list -r`, `find -r`), it searches all subdirectories.

## Examples

### Daily Workflow

```bash
# Morning: check your tasks
mdd list

# Add tasks as they come up
mdd "Review PR #42"
mdd "Deploy to staging @DevOps"

# Start working on something
mdd toggle 1

# Done with a task
mdd complete 1

# End of day: clean up
mdd clear
```

### Project Organization

```markdown
# TODO

- [ ] Write tests

## Backend

- [ ] Implement user auth
- [ ] Add rate limiting

## Frontend

- [ ] Design login page
- [x] Setup React Router

## Notes

- API rate limit is 100 req/min
- Deploy window: 2-4am UTC
```

## Building

```bash
make build       # Build binary to ./build/mdd
make install     # Install to GOPATH/bin
make test        # Run tests
make lint        # Lint code (requires golangci-lint)
make build-all   # Cross-compile for all platforms
```

## Why Markdown-do?

- **No lock-in** - Your tasks are plain markdown, version-controlled with your code
- **Fast** - Single binary, instant startup, no daemon or runtime dependencies
- **Minimal** - Does one thing well: manage TODO.md files

## License

MIT
