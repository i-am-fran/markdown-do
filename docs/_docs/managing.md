---
title: Managing Tasks
slug: managing
order: 4
headings:
  - id: cmd-toggle
    label: toggle
  - id: cmd-complete
    label: complete
  - id: cmd-edit
    label: edit
  - id: cmd-annotate
    label: annotate
  - id: cmd-remove
    label: remove
  - id: cmd-clear
    label: clear
  - id: cmd-archive
    label: archive
  - id: cmd-undo
    label: undo
---

Commands that change a task's status or content. All of them take an `<id>` — a task's position number from `mdd list`, or a stable ID tag once you've run `mdd tag` (see Task IDs below).

### toggle {#cmd-toggle}

`mdd toggle <id>` cycles a task's status. By default that's a simple flip: pending <-> completed. With the `enableInProgress` setting on, it becomes a 3-way cycle: pending -> in-progress -> completed -> pending.

`-t` is a short-hand alias for `toggle`.

```bash
mdd toggle 2
```

### complete {#cmd-complete}

`mdd complete <id> [id2 id3 ...]` marks one or more tasks complete directly, skipping the toggle cycle. Each ID is resolved independently, so `mdd complete 1 2 3` still completes #2 and #3 even if #1 doesn't exist — you get a warning about the one that failed instead of the whole command aborting.

`-c` is a short-hand alias for `complete`, and takes multiple IDs the same way (e.g. `mdd -c 1 2 3`).

```bash
mdd complete 1
mdd complete 1 2 3
```

### edit {#cmd-edit}

`mdd edit <id> <text>` replaces a task's text outright — it's a full overwrite, not an append. Multi-word text needs no quotes.

`-e` is a short-hand alias for `edit`.

```bash
mdd edit 1 Fix the login bug instead
```

### annotate {#cmd-annotate}

`mdd annotate <id> <text>` adds an indented note line under one specific task, shown inline wherever that task is listed. Not the same as `notes`, which is file-wide rather than attached to a task. Notes travel with the task through edit, move, complete, and delete.

`-an` is a short-hand alias for `annotate`.

```bash
mdd annotate 1 Needs a second reviewer
```

### remove {#cmd-remove}

`mdd remove <id>` deletes one task. Prompts for a y/n confirmation first if the `confirmDestructive` setting is on; add `-y`/`--yes` to skip it, which is also handy when calling `mdd` from a script.

`-d` is a short-hand alias for `remove`.

| Flag | Description |
| --- | --- |
| `-y`, `--yes` | Skip the confirmation prompt (only matters when `confirmDestructive` is on) |

```bash
mdd remove 3
mdd remove 3 -y
```

### clear {#cmd-clear}

`mdd clear` deletes every completed task in the file at once. Same confirmation behavior as `remove`.

`-dc` is a short-hand alias for `clear`.

| Flag | Description |
| --- | --- |
| `-y`, `--yes` | Skip the confirmation prompt (only matters when `confirmDestructive` is on) |

```bash
mdd clear
mdd clear -y
```

### archive {#cmd-archive}

`mdd archive` moves every completed task into a `## Archive` section instead of deleting it, appending "(from <Section>)" to each one so you can still tell where it came from. The Archive section is always kept last in the file, after `## Notes`. Unlike `remove`/`clear`, this isn't gated by `confirmDestructive` — nothing is actually deleted.

```bash
mdd archive
```

### undo {#cmd-undo}

`mdd undo` reverts the last change made to the TODO file. Because the restore is written through the normal save path, running `mdd undo` again re-applies the change it just undid — it's a toggle, not a multi-step history.

```bash
mdd undo
```
