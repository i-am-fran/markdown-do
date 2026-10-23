---
title: Getting Started
slug: getting-started
order: 1
headings:
  - id: installation
    label: Installation
  - id: quick-start
    label: First steps
  - id: task-format
    label: Task format
---

Markdown-do (`mdd`) manages the `TODO.md` file in your project from the terminal. It's a single binary with no server, database or account. Tasks are plain markdown checkboxes, so the file stays readable in any editor even if you never run `mdd` again.

![A terminal recording touring mdd: add, list, find, toggle, complete, edit, annotate, notes, lint, clear, archive, tag and untag, undo, remove, recursive list and find, shell completion, config and open]({{ '/assets/full-demo.gif' | relative_url }}){: .doc-section__media}

### Installation {#installation}

With Go installed:

```bash
go install github.com/i-am-fran/markdown-do/v3/cmd/mdd@latest
```

Without Go:

1. Download the binary for your system from the [latest release](https://github.com/i-am-fran/markdown-do/releases/latest){:target="_blank" rel="noopener noreferrer"}: `mdd-darwin-arm64` or `mdd-darwin-amd64` for a Mac, `mdd-linux-amd64` or `mdd-linux-arm64` for Linux, `mdd-windows-amd64.exe` for Windows.
2. Rename it to `mdd`, make it executable (`chmod +x mdd`) and move it to a folder on your `PATH`.

From source:

```bash
git clone https://github.com/i-am-fran/markdown-do.git
cd markdown-do
make install    # or make build, which puts the binary in ./build/mdd
```

To update later, run `mdd update` (see [update](#cmd-update)).

<div class="actions not-prose">
  <a class="button button--primary" href="https://github.com/i-am-fran/markdown-do/releases/latest">Latest release{% include arrow-icon.html %}</a>
</div>

### First steps {#quick-start}

Run `mdd` in the folder of a project. It uses the `TODO.md` there (or `todo.md`), and creates `TODO.md` the first time you add a task.

```bash
mdd Reply to Sam about the invoice     # add a task to the inbox
mdd "Fix the date parser @bb"          # add a task under ## Bugs
mdd list                               # number and list every task
mdd complete 1                         # tick off task 1
```

`mdd` on its own prints a short banner with the most-used commands. `mdd help` prints all of them.

### Task format {#task-format}

Tasks are standard markdown checkboxes, plus a third state for work in progress:

```markdown
## Features

- [ ] Pending task
- [/] In-progress task
- [x] Completed task
  - A note about the task
  - Another note
```

Indented plain bullets under a task (no checkbox) are its notes. You add them with `mdd annotate`, and they move with the task when it's edited, completed, archived or removed.

`mdd` always reads `[/]` correctly, but toggling only goes through it when the `enableInProgress` setting is on.

On every save, tasks in each section are reordered: open tasks first, completed ones after them. Completed tasks stay in the file until you [clear](#cmd-clear) or [archive](#cmd-archive) them.
