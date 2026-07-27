---
title: Task IDs
slug: task-ids
order: 5
headings:
  - id: cmd-tag
    label: tag
  - id: cmd-untag
    label: untag
---

A task's position number shifts every time the file is re-sorted or a task above it is removed — fine for a quick `mdd complete 1` right after `mdd list`, less fine if you want to refer back to the same task tomorrow, or hand a task reference to an LLM across a session boundary. Tag the file once and you get a stable ID that survives all of that.

### tag {#cmd-tag}

`mdd tag <PREFIX>` tags every task in the file with sequential IDs: `PREFIX-001`, `PREFIX-002`, and so on. `PREFIX` must be exactly 3 letters (case-insensitive — `abc` is normalized to `ABC`). Once a sequence is active, any task you add afterward automatically picks up the next ID in it.

```bash
mdd tag ABC             # Tags all tasks: [ABC-001], [ABC-002], ...
mdd complete ABC-001     # Reference a task by its stable ID
```

### untag {#cmd-untag}

`mdd untag` strips every ID tag from the file. Printing "No task IDs to remove" and exiting cleanly if there weren't any to begin with.

```bash
mdd untag
```

Any command above that takes an `<id>` — `toggle`, `complete`, `edit`, `annotate`, `remove` — accepts either a task's position number or its stable ID tag, once IDs have been assigned.
