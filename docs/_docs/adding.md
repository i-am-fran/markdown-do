---
title: Adding Tasks
slug: adding
order: 2
headings:
  - id: cmd-add
    label: add
  - id: cmd-notes
    label: notes
  - id: sections
    label: Sections
---

Adding is the default action — if what you type doesn't look like one of the commands below, it becomes a new task.

<aside class="callout callout--tip not-prose" markdown="1">
**Tip:** A word like `complete` or `list` is only read as a command when what follows it matches that command's shape (e.g. `complete` needs a task-ID-like next token); otherwise it falls through to an implicit add, so `mdd Complete the tax return` still just adds a task titled "Complete the tax return". Use `mdd add "..."` to force add if your task text is ever misread as a command.
</aside>

### add {#cmd-add}

`mdd <text>` and `mdd add <text>` do the same thing; quotes are only needed if your shell would otherwise split the text into multiple arguments. Task text must be at least 3 characters. New tasks land in the inbox at the top of the file unless you file them into a section with `@Section` (see Sections below) — and if the file already has an active ID sequence from `mdd tag`, the new task automatically picks up the next ID in it.

```bash
mdd Buy groceries
mdd "Fix login bug @bb"
mdd add "-1, needs more thought first"   # forces add; a leading "-" is otherwise read as a flag
```

### notes {#cmd-notes}

`mdd notes <text>` adds a line to the file-wide `## Notes` section (created automatically if it doesn't exist yet) — for context that isn't itself a task, like a reminder or a decision you want on record. This is different from `annotate`, which attaches a note to one specific task.

```bash
mdd notes "API rate limit is 100 req/min"
mdd notes Deploy window is 2-4am UTC
```

### Sections {#sections}

End a task with `@Section` to file it there — the section is created automatically if it doesn't exist yet. Tasks added without a section go to the inbox at the top of the file.

```bash
mdd "New feature @Features"    # Adds to ## Features (creates if needed)
mdd "Another task"             # Adds to the inbox (top of file)
```

Built-in section aliases, case-insensitive: `@ff` → Features, `@bb` → Bugs, `@ii` → Ideas, `@ww` → Warnings. Add your own with `mdd config set alias.<name> <Section>`.
