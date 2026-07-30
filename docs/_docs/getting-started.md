---
title: Getting Started
slug: getting-started
order: 1
headings:
  - id: installation
    label: Installation
  - id: quick-start
    label: Quick Start
  - id: task-format
    label: Task Format
---

Markdown-do (`mdd`) is a single Go binary for managing `TODO.md` files with standard markdown checkbox syntax. No server, no lock-in: the file stays readable in any editor even if you never install `mdd` again.

![Terminal recording touring mdd: add, list, find, toggle, complete, edit, annotate, notes, lint, clear, archive, tag/untag, undo, remove, recursive list/find, shell completion, config, and open](assets/full-demo.gif){: .doc-section__media}

### Installation {#installation}

Download the latest binary for your platform from the [releases page](https://github.com/i-am-fran/markdown-do/releases){:target="_blank" rel="noopener noreferrer"}, or build from source:

```bash
git clone https://github.com/i-am-fran/markdown-do.git
cd markdown-do
make build
# Binary is at ./build/mdd

# Or install to GOPATH/bin:
make install
```

Or install directly with Go:

```bash
go install github.com/i-am-fran/markdown-do/v3/cmd/mdd@latest
```

### Quick Start {#quick-start}

```bash
# Add your first task
mdd Buy milk

# List tasks
mdd list

# Complete task #1
mdd complete 1
```

### Task Format {#task-format}

Tasks are standard markdown checkboxes, plus a non-standard third state:

```markdown
## Features

- [ ] Pending task
- [/] In-progress task (only reachable via toggle when enableInProgress is on)
- [x] Completed task
  - A note about the task
  - Another note
```

A task's trailing indented plain-bullet lines (no checkbox) are its notes — added with `mdd annotate <id> "text"`, they always travel with the task through edit, move, delete, and complete. On save, tasks within each section are automatically reordered: pending first, then completed, keeping active work visible at the top.
