---
title: All features
slug: all-features
order: 8
headings:
  - id: features-file
    label: The file
  - id: features-adding
    label: Adding
  - id: features-viewing
    label: Viewing and finding
  - id: features-managing
    label: Managing
  - id: features-ids
    label: Task IDs
  - id: features-utilities
    label: Utilities
  - id: features-settings
    label: Settings
---

### The file {#features-file}

- Plain markdown checkboxes: `- [ ]`, `- [x]`, and `- [/]` for in progress
- `## Sections`, with an inbox at the top for tasks without one
- Uses `TODO.md` or `todo.md` in the current folder, creates `TODO.md` if there's neither
- Open tasks sorted above completed ones in each section on every save
- Notes under a task as indented bullets, kept with it through every change
- A file-wide `## Notes` section, always kept at the end
- Nested checklist items keep their indentation
- Atomic saves with a lock file, so a crash or a second `mdd` can't corrupt the file

### Adding {#features-adding}

- `mdd <text>` adds a task, no quotes needed
- `mdd add <text>` forces an add when the text looks like a command
- `@Section` files a task under a section, creating it if needed
- Built-in shortcuts: `@ff` Features, `@bb` Bugs, `@ii` Ideas, `@ww` Warnings
- Your own shortcuts with `mdd config set alias.<name> <Section>`
- `\@` for a literal `@`, `\\` for a literal backslash
- `--path <dir>` adds to another folder's TODO file
- New tasks pick up the next ID once the file is tagged
- Task text must be at least 3 characters
- A mistyped single-word command (`mdd lis`) suggests the real one instead of adding a task
- `mdd notes <text>` adds a line to `## Notes`

### Viewing and finding {#features-viewing}

- `mdd list` groups tasks by section, inbox first, with notes inline
- `-a`/`--active` for open tasks only, `--done` for completed only
- `-r` lists every `TODO.md` in subfolders, grouped by folder
- Numbers from a recursive list keep working in the next command
- `mdd find <keyword>`, case-insensitive, with `-r` for subfolders

### Managing {#features-managing}

- `toggle` flips pending and completed, or cycles through in progress
- `complete` takes several IDs at once and skips any that don't exist
- `edit` replaces a task's text
- `annotate` adds a note under one task
- `remove` deletes one task, `clear` deletes every completed task
- Optional y/n prompt before deleting, skipped with `-y`
- `archive` moves completed tasks to `## Archive`, adding "(from Section)"
- `undo` reverts the last change; run it again to redo

### Task IDs {#features-ids}

- `mdd tag ABC` gives every task an ID like `ABC-001`
- Any three-letter prefix, normalised to capitals
- Every command that takes a task number also takes its ID
- `mdd untag` removes them all

### Utilities {#features-utilities}

- `lint` removes empty sections, empty tasks and extra blank lines
- `lint` fixes heading spacing, `[]`, `[X]` and malformed checkboxes
- `lint` turns checkboxes under `## Notes` into plain bullets
- `open` opens the file in your editor
- Tab completion for bash and zsh: commands and section names
- `update` installs the latest release after checking its signature
- Short aliases for every command: `-l`, `-f`, `-t`, `-c`, `-e`, `-an`, `-d`, `-n`, `-o`, `-dc`, `-lint`, `-id`
- `-h` and `-v` work anywhere in the command
- A banner with the most-used commands when you run `mdd` on its own

### Settings {#features-settings}

- `mdd config list`, `get`, `set` and `edit`
- Show or hide completed tasks
- Editor: system (`$EDITOR`), vim, nano or the default app
- In-progress state for toggling
- Confirm before deleting
- A broken `config.json` prints a warning and falls back to defaults
