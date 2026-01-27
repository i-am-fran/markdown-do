---
title: MarkdownDO
description: An opinionated CLI/TUI tool to manage TODO.md files
show_downloads: true
---
## Intro

MarkdownDO (or markdown-do) is a fast, minimal, opinionated CLI/TUI tool for managing `TODO.md` files using standard markdown syntax.

## Features

- **Plain markdown** - Tasks are standard `- [ ]` checkboxes, readable anywhere
- **Three task states** - Pending `[ ]`, in progress `[/]`, and completed `[x]`
- **Sections** - Organize tasks with `## Headers`
- **Inbox workflow** - New tasks go to the top unless you specify a section
- **Auto-sorting** - Tasks reorder by status on save (pending → in progress → done)
- **Fast CLI** - Add, toggle, and manage tasks without leaving your terminal
- **Interactive TUI** - Full keyboard navigation with hotkeys
- **Recursive search** - Find tasks across all TODO files in subdirectories
- **Lint & fix** - Auto-fix common formatting issues

## Examples

```bash
# Add your first task
mdd Buy milk

# List tasks
mdd -l

# Toggle task #1 (pending → in progress → done)
mdd -t 1

# Open interactive mode
mdd
```
