# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [3.2.2] - 2026-07-27

### Added
- A `.github/workflows/release.yml` pipeline now builds all five platform
  binaries (`make build-all`) and a `checksums.txt`, and attaches them to
  the GitHub release whenever one is published. `mdd update` now actually
  downloads and swaps in the matching release binary for the running
  process, verifying its sha256 checksum first, instead of shelling out to
  `go install` (MDD32).

### Changed
- `mdd` no longer writes `.mdd-cache` / `.mdd-undo` into the project
  directory it's operating on. Both now live under
  `~/.config/markdowndo/` (keyed by a hash of the relevant absolute path)
  alongside `config.json` (MDD35).

### Fixed
- `internal/cli` package tests read the real `~/.config/markdowndo/config.json`
  with no isolation, so `TestPerformToggleTask` failed whenever
  `enableInProgress` was set to `true` on the machine running `go test`.
  Tests now point settings at a scratch directory for the whole package
  run, the same way cache/undo state was already isolated (MDD36).

## [3.2.1] - 2026-07-25

### Fixed
- `mdd update` (and the README/docs "Go install" instructions) used
  `go install .../cmd/mdd@latest`, which silently installed the ancient
  `v1.0.0` tag instead of the real latest release — this repo's release
  tags are bare semver (`3.2.0`) and Go's module tooling only recognizes
  `v`-prefixed tags as valid versions. `mdd update` now looks up the
  latest release via the GitHub API and installs that exact tag instead
  of relying on `@latest`.

## [3.2.0] - 2026-07-25

### Added
- `mdd undo` reverts the last change made through `mdd`; running it again
  re-applies the change it just undid (MDD21).
- `mdd update` installs the latest release via `go install` (MDD30).

### Fixed
- The Go module path in `go.mod` was `github.com/i-am-fran/markdowndo`
  (no dash), which doesn't match this repo's actual GitHub name
  (`i-am-fran/markdown-do`) — `go install`/`go get` against the documented
  path has never actually resolved. Corrected the module path and every
  internal import to match the real repo.

### Changed
- Renamed the project everywhere from "MarkdownDO" to "Markdown-do"
  (MDD31).

## [3.1.0] - 2026-07-25

### Added
- `mdd completion bash` / `mdd completion zsh` print a shell completion
  script (`source <(mdd completion zsh)` in your rc file) that tab-completes
  command verbs (`mdd ar` → `mdd archive`) and section tags, including
  built-in/custom aliases and any section already in `TODO.md`
  (`@Bu` → `@Bugs`) (MDD16).

## [3.0.1] - 2026-07-25

### Changed
- `mdd tag` now generates IDs with a separator and 3-digit padding from the
  start, e.g. `ABC-001` instead of `ABC01` (MDD29, MDD28). This keeps the
  format visually aligned past 99 items instead of drifting from 2 to 3
  digits mid-sequence; numbers still grow past 999 without truncating.
  IDs already tagged in the old `ABC01` format continue to be read and
  referenced correctly — only newly generated IDs use the new format.

## [3.0.0] - 2026-07-25

### Removed
- **Breaking:** the TUI is gone. `internal/tui` (~4,800 lines) is deleted,
  along with its dependencies (`bubbletea`, `bubbles`, `lipgloss`, `glamour`,
  `sahilm/fuzzy`). A bare `mdd` invocation now always shows help instead of
  launching an interactive menu.
- **Breaking:** the TUI-only settings — `theme`, `fullscreen`,
  `showStatusBar`, `enableAnimations`, `enableTUI` — are removed from
  `config.json` handling and from `mdd config list/get/set`. Existing config
  files with those keys still load fine; the keys are just ignored.

## [2.2.0] - 2026-07-25

### Added
- `mdd archive` moves every completed task into a `## Archive` section
  (created automatically, always kept last, ahead of `## Notes`), appending
  "(from `<Section>`)" to record where each task came from. Unlike `mdd
  clear`, this doesn't delete anything and isn't gated by
  `confirmDestructive`.

## [2.1.0] - 2026-07-24

### Added
- `mdd config list` / `mdd config get <key>` / `mdd config set <key> <value>`
  to view and change settings from the CLI, without opening the TUI.
- Optional third task state, `- [/]` (in-progress), gated by the new
  `enableInProgress` setting. When on, toggling cycles pending -> in-progress
  -> completed -> pending instead of the plain pending/completed flip;
  `- [/]` always parses and displays correctly regardless of the setting.
- `confirmDestructive` setting: prompts for y/n before `mdd clear` / `mdd
  remove`; add `-y`/`--yes` to skip the prompt.
- `enableTUI` setting: when off, a bare `mdd` invocation shows help instead
  of launching the TUI.
- `sectionAliases` setting: add custom `@alias` shortcuts alongside the
  built-in `ff`/`bb`/`ii`/`ww`, via `mdd config set alias.<name> <Section>`.
- `mdd config edit` (and an "Edit config file" item in the TUI Settings menu)
  opens `config.json` directly in the configured editor.

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
- Initial release of Markdown-do
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
