# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0] - 2026-07-24

### Changed
- **Breaking:** replaced the terse flag-based CLI (`-l`, `-c`, `-cm`, `-d`,
  `-e`, `-t`, `-n`, `-an`, `-dc`, `-lint`, `-id`, `-id -r`, `-o`, `-f`/`-fs`,
  `-ls`) with readable verb subcommands: `list`, `find`, `complete`,
  `toggle`, `edit`, `annotate`, `remove`, `clear`, `notes`, `open`, `lint`,
  `tag`, `untag`, `version`, `help`, plus a new explicit `add`. Recursive
  listing/finding is now a single `-r`/`--recursive` modifier instead of
  separate `-ls`/`-fs` flags, and `complete` now takes any number of task
  IDs directly (folding in the old `-cm`).
- A word matching a command name is only read as that command when what
  follows it fits the command's shape (e.g. `complete` needs a task ID
  next); otherwise it's treated as task text, so `mdd Complete the tax
  return` still just adds a task. Use `mdd add "..."` to force an add if
  needed.
- Renamed the per-task note flag `-an` to `mdd annotate <id> <text>`, and the
  global-notes flag `-n` to `mdd notes <text>`, removing a longstanding
  source of confusion between the two.
- Folds in the old `-d` → `-r` rename request: task deletion is now the verb
  `mdd remove <id>`.

## [1.1.0] - 2026-07-24

### Added
- Per-task notes: attach freeform notes to a task with `mdd -an <id> <text>`,
  expressed as an indented bullet list following the checklist item. Notes
  render inline wherever the task is listed and travel with it through
  edit, move, delete, and complete.

## [1.0.0] - 2026-02-06

### Added
- Initial release of MarkdownDO
- CLI for managing TODO.md files
- Interactive TUI with keyboard navigation
- Task management features (add, complete, delete, edit, move)
- Complete multiple tasks at once with `-cm` flag (e.g., `mdd -cm 1 2 3`)
- Global `Ctrl+N` shortcut in TUI to quickly create new tasks from any view
- Section support with `## Headers`
- Recursive search across subdirectories
- Lint and auto-fix functionality
- Configuration file support
- Multiple editor options
- Auto-sorting of tasks by status

[1.0.0]: https://github.com/i-am-fran/markdown-do/releases/tag/v1.0.0
