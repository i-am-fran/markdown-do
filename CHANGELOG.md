# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2026-01-28

### Added
- Global `Ctrl+N` keyboard shortcut in TUI to quickly create new tasks from any view
- New `-cm` CLI flag to complete multiple tasks at once (e.g., `mdd -cm 1 2 3`)
  - Handles invalid task IDs gracefully with warnings
  - Shows clear success/error feedback
- `UPGRADE.md` file with comprehensive upgrade instructions
- `CHANGELOG.md` to track version history

### Changed
- **CLI output improvement**: When adding tasks via CLI, output now shows clean text without checkbox and task number
  - Before: `Added:   1. [ ] Buy groceries`
  - After: `Added: Buy groceries`
- Menu reorganization: Moved "Add note" lower in the menu for better workflow hierarchy
- Help text updated to include new `-cm` flag with examples

### Removed
- Redundant "Back" button from task list navigation (ESC key already provides this functionality)
- Redundant "Quit" button from main menu (ESC ESC already quits the application)
- Dead code: Removed unused `isBack` field and `ActionQuit` constant

### Fixed
- Code cleanup and improved maintainability
- Proper error handling for multiple task completion

## [1.0.0] - 2026-01-XX

### Added
- Initial release of MarkdownDO
- CLI for managing TODO.md files
- Interactive TUI with keyboard navigation
- Task management features (add, complete, delete, edit, move)
- Section support with `## Headers`
- Recursive search across subdirectories
- Lint and auto-fix functionality
- Configuration file support
- Multiple editor options
- Auto-sorting of tasks by status

[1.1.0]: https://github.com/i-am-fran/markdown-do/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/i-am-fran/markdown-do/releases/tag/v1.0.0
