# MarkdownDO

A fast, minimal, opinionated CLI/TUI tool for managing TODO.md files using standard markdown syntax.

**Current Version: 1.1.0** | [Changelog](CHANGELOG.md) | [Upgrade Guide](UPGRADE.md)

## Features

- **Plain markdown** - Tasks are standard `- [ ]` checkboxes, readable anywhere
- **Two task states** - Pending `[ ]` and completed `[x]`
- **Sections** - Organize tasks with `## Headers`
- **Inbox workflow** - New tasks go to the top unless you specify a section
- **Auto-sorting** - Tasks reorder by status on save (pending -> completed)
- **Fast CLI** - Add, toggle, and manage tasks without leaving your terminal
- **Interactive TUI** - Full keyboard navigation with hotkeys
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
go install github.com/i-am-fran/markdowndo/cmd/mdd@latest
```

## Quick Start

```bash
# Add your first task
mdd Buy milk

# List tasks
mdd -l

# Toggle task #1 (pending <-> completed)
mdd -t 1

# Open interactive mode
mdd
```

## CLI Reference

### Adding Tasks

```bash
mdd Buy groceries              # Add task (quotes optional)
mdd "Fix bug in auth"          # Add task with quotes
mdd "New feature @Features"    # Add to specific section
mdd -n "Meeting notes"         # Add note to ## Notes section
```

### Viewing Tasks

```bash
mdd -l                         # List tasks in current directory
mdd -ls                        # List tasks recursively (subdirectories)
mdd -f bug                     # Find tasks containing "bug"
mdd -fs auth                   # Find recursively
```

### Managing Tasks

```bash
mdd -t 1                       # Toggle task #1 status
mdd -c 1                       # Mark task #1 as complete
mdd -e 1 "Updated text"        # Edit task #1
mdd -d 1                       # Delete task #1
mdd -dc                        # Delete all completed tasks
```

### File Operations

```bash
mdd -o                         # Open TODO.md in editor
mdd -lint                      # Lint and auto-fix formatting
mdd -h                         # Show help
```

## Interactive Mode (TUI)

Run `mdd` without arguments to enter interactive mode.

### Main Menu

- **List tasks** - View and manage tasks with keyboard shortcuts
- **Add new task** - Quick task entry
- **Open in editor** - Edit TODO.md directly
- **List subfolders** - Browse TODO files in subdirectories
- **Settings** - Configure preferences

### Task List Hotkeys

| Key | Action |
| --- | ------ |
| `up` `down` | Navigate |
| `Enter` | Select task |
| `c` | Toggle complete |
| `d` | Delete task |
| `e` | Edit task |
| `m` | Move to section |
| `Esc` | Go back |
| `Esc` `Esc` | Quit |

## Task Format

MarkdownDO uses standard markdown checkbox syntax:

```markdown
# TODO

- [ ] Pending task
- [x] Completed task

## Features

- [ ] Add dark mode @Features
- [x] Initial setup done
```

### Status Cycle

Toggle (`mdd -t N` or `c` in TUI) cycles through:

```text
pending [ ] <-> completed [x]
```

Use `mdd -c N` to mark a task as complete directly.

### Sections

Organize tasks under `## Header` sections:

```bash
mdd "New task @Backend"        # Adds to ## Backend (creates if needed)
mdd "Another task"             # Adds to inbox (top of file)
```

Section aliases for quick add: `@ff` -> Features, `@bb` -> Bugs, `@ii` -> Ideas, `@ww` -> Warnings

In the TUI, use `m` to move tasks between sections.

### Auto-Sorting

On save, tasks within each section are automatically reordered:

1. Pending tasks
2. Completed tasks

This keeps your active work visible at the top.

## Configuration

Settings are stored in `~/.config/markdowndo/config.json`.

| Setting | Default | Description |
| ------- | ------- | ----------- |
| `fullscreen` | `false` | Use alternate screen buffer for TUI |
| `showCompleted` | `true` | Show completed tasks in lists |
| `editor` | `"system"` | Editor for `-o` command |
| `theme` | `"default"` | TUI color theme |

### Editor Options

- `system` - Uses `$EDITOR` or `$VISUAL` environment variable
- `vim` - Open in Vim
- `nano` - Open in Nano
- `default-app` - Open with system default application

## File Discovery

MarkdownDO looks for these files (in order):

1. `TODO.md` in current directory
2. `todo.md` in current directory
3. Creates `TODO.md` if none exists

For recursive operations (`-ls`, `-fs`), it searches all subdirectories.

## Examples

### Daily Workflow

```bash
# Morning: check your tasks
mdd -l

# Add tasks as they come up
mdd "Review PR #42"
mdd "Deploy to staging @DevOps"

# Start working on something
mdd -t 1

# Done with a task
mdd -c 1

# End of day: clean up
mdd -dc
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

## Why MarkdownDO?

- **No lock-in** - Your tasks are plain markdown, version-controlled with your code
- **Fast** - Single binary, instant startup, no daemon or runtime dependencies
- **Keyboard-first** - CLI for quick actions, TUI for exploration
- **Minimal** - Does one thing well: manage TODO.md files

## License

MIT
