---
title: Configuration
slug: configuration
order: 7
headings:
  - id: settings
    label: Settings
  - id: file-discovery
    label: File Discovery
---

### Settings {#settings}

Settings are stored in `~/.config/markdowndo/config.json`, editable via `mdd config get/set` or `mdd config edit`.

| Setting | Default | Description |
| --- | --- | --- |
| `showCompleted` | `true` | Show completed tasks in lists. |
| `editor` | `"system"` | Editor used by `mdd open` — `system`, `vim`, `nano`, or `default-app`. |
| `enableInProgress` | `false` | Enables the `- [/]` in-progress checkbox state and the 3-way toggle cycle. |
| `confirmDestructive` | `false` | Require a y/n prompt before `remove`/`clear` (skip with `-y`/`--yes`). |
| `sectionAliases` | `{}` | Custom `@alias` shortcuts layered on top of the built-in ones. |

### File Discovery {#file-discovery}

Markdown-do looks for these files, in order:

1. `TODO.md` in the current directory
2. `todo.md` in the current directory
3. Creates `TODO.md` if none exists

For recursive operations (`list -r`, `find -r`), it searches all subdirectories.
