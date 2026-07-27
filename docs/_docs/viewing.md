---
title: Viewing & Finding
slug: viewing
order: 3
headings:
  - id: cmd-list
    label: list
  - id: cmd-find
    label: find
---

Read-only commands for seeing what's in a `TODO.md`.

### list {#cmd-list}

`mdd list` prints every task grouped by section in file order, with the inbox shown first. Add `-r`/`--recursive` to walk every subdirectory instead of just the current one — in that mode, task numbers become temporary global IDs cached to disk, so a follow-up like `mdd complete 14` still resolves to the right task even though it lives in a different file than task 1.

`-l` is a short-hand alias for `list` (e.g. `mdd -l -r`).

| Flag | Description |
| --- | --- |
| `-r`, `--recursive` | List tasks from every `TODO.md`/`todo.md` found in subdirectories, grouped by directory |

```bash
mdd list
mdd list -r
```

### find {#cmd-find}

`mdd find <keyword>` searches every task's text for a case-insensitive substring match, in the current file by default. Add `-r` to search every `TODO.md` in subdirectories too.

`-f` is a short-hand alias for `find` (e.g. `mdd -f bug -r`).

| Flag | Description |
| --- | --- |
| `-r`, `--recursive` | Search every `TODO.md`/`todo.md` found in subdirectories |

```bash
mdd find bug
mdd find auth -r
```
