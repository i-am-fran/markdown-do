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
  - id: escaping
    label: Escaping & Quoting
---

Adding is the default action — if what you type doesn't look like one of the commands below, it becomes a new task.

<aside class="callout callout--tip not-prose" markdown="1">
**Tip:** A word like `complete` or `list` is only read as a command when what follows it matches that command's shape (e.g. `complete` needs a task-ID-like next token); otherwise it falls through to an implicit add, so `mdd Complete the tax return` still just adds a task titled "Complete the tax return". Use `mdd add "..."` to force add if your task text is ever misread as a command.
</aside>

<aside class="callout callout--warning not-prose" markdown="1">
**Note:** If that single word is itself a near-miss of a known command — an exact match but wrong case or missing its required argument (e.g. `mdd Complete`, bare `mdd complete`), or a likely typo (e.g. `mdd lis`, `mdd toggel`) — mdd refuses the implicit add and suggests the real command instead of silently creating a task titled after your typo. This only applies when the whole input is a single word, so multi-word task text is never affected. Use `mdd add "..."` to force add in that case too.
</aside>

### add {#cmd-add}

`mdd <text>` and `mdd add <text>` do the same thing; quotes are only needed if your shell would otherwise split the text into multiple arguments. Task text must be at least 3 characters. New tasks land in the inbox at the top of the file unless you file them into a section with `@Section` (see Sections below) — and if the file already has an active ID sequence from `mdd tag`, the new task automatically picks up the next ID in it.

```bash
mdd Buy groceries
mdd "Fix login bug @bb"
mdd add "-1, needs more thought first"   # forces add; a leading "-" is otherwise read as a flag
```

Add `--path <dir>` to add to the TODO file in `<dir>` instead of the current directory:

```bash
mdd add "Renew SSL cert" --path ~/projects/infra
```

### notes {#cmd-notes}

`mdd notes <text>` adds a line to the file-wide `## Notes` section (created automatically if it doesn't exist yet) — for context that isn't itself a task, like a reminder or a decision you want on record. This is different from `annotate`, which attaches a note to one specific task.

`-n` is a short-hand alias for `notes`.

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

### Escaping & Quoting {#escaping}

Prefix a character with `\` to keep it literal instead of special. This only applies to `add`'s text — `edit`/`annotate` text is always stored literally.

```bash
mdd 'Meet Bob \@bb'   # adds a task ending in a literal "@bb" instead of filing it under Bugs
mdd 'Cost is \\$10'   # \\ for a literal backslash
```

Some shells (zsh in particular) treat an unquoted trailing `?` or `*` as a glob pattern and will error with "no matches found" before mdd ever runs — quote task text containing those characters, e.g. `mdd "Ping the vendor?"`.
