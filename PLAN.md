# MarkdownDO UI/UX Enhancement Plan

Implement 7 features using the Charmbracelet ecosystem to enhance the TUI.

## New Dependencies

```go
github.com/charmbracelet/glamour   // Markdown rendering
github.com/charmbracelet/harmonica // Spring animations
// bubbles/help, bubbles/paginator, bubbles/progress already available
```

## Features Overview

| # | Feature | Complexity | New Files |
|---|---------|------------|-----------|
| 1 | Smooth List Transitions | Medium | `internal/tui/animation/spring.go` |
| 2 | Rich Markdown Preview | Low | `internal/tui/markdown/renderer.go` |
| 3 | Fuzzy Search + Live Preview | Medium | Modify `views/search.go` |
| 4 | Contextual Help Overlay | Low | `internal/tui/help/keymap.go` |
| 5 | Status Bar | Low | `internal/tui/views/statusbar.go` |
| 6 | Pagination Indicator | Low | Modify `views/tasklist.go` |
| 7 | Undo + Toast Notifications | High | `internal/tui/undo/stack.go`, `views/toast.go` |
| 8 | Quick Actions Palette | Medium | `internal/tui/views/palette.go` |

---

## Implementation Order

### Phase 1: Foundation (Low Risk)

#### Feature 4: Help Overlay
- Extend `keys.go` to implement `help.KeyMap` interface
- Add `helpModel help.Model` and `showHelp bool` to app.go Model
- Handle `?` key globally to toggle
- Create `internal/tui/help/keymap.go` with per-view keymaps

#### Feature 5: Status Bar
- Create `internal/tui/views/statusbar.go`
- Show: `12/20 tasks | ~/project/TODO.md`
- Add `statusBarModel` to app.go, update on file changes
- Reserve bottom line in View() calculations

### Phase 2: Core Enhancements

#### Feature 6: Pagination
- Add `paginator.Model` to TaskListModel
- Use dot-style indicators: `● ○ ○ ○`
- Sync with list cursor, handle pgup/pgdown
- Only show when items exceed one page

#### Feature 3: Fuzzy Search
- Enable list filtering: `l.SetFilteringEnabled(true)`
- Use `github.com/sahilm/fuzzy` (already indirect dep)
- Highlight matched characters in results
- Add preview panel showing task context

### Phase 3: Visual Polish

#### Feature 2: Markdown Preview
- Create `internal/tui/markdown/renderer.go` wrapping glamour
- Render task text in TaskActionsModel view
- Support **bold**, *italic*, `code`, [links]
- Cache rendered output per task

#### Feature 1: Smooth Animations
- Create `internal/tui/animation/spring.go` with harmonica helpers
- Add animation state map to TaskListModel
- Animate on toggle (flash effect) and delete (collapse)
- Use `tea.Tick` at 60fps during active animations

### Phase 4: Advanced Features

#### Feature 7: Undo System
- Create `internal/tui/undo/stack.go` - stores snapshots
- Add `Snapshot()` and `Restore()` to TodoFile
- Create `internal/tui/views/toast.go` with progress countdown
- Snapshot before: delete, toggle, edit, move
- Handle `u` key globally when undo available (5s timeout)

#### Feature 8: Command Palette
- Create `internal/tui/views/palette.go`
- Modal overlay with fuzzy-filtered command list
- Commands: Add Task, List, Find, Settings, Lint, Quit, etc.
- Handle `ctrl+k` globally to toggle
- Execute selected command via existing message types

---

## Key File Modifications

### `internal/tui/app.go`
```go
type Model struct {
    // Existing fields...

    // New fields
    helpModel      help.Model
    showHelp       bool
    statusBarModel views.StatusBarModel
    undoStack      *undo.Stack
    toastModel     views.ToastModel
    paletteModel   views.PaletteModel
    showPalette    bool
}
```

New global key handlers:
- `?` - Toggle help overlay
- `u` - Undo (when available)
- `ctrl+k` - Toggle command palette

### `internal/tui/keys.go`
Implement `help.KeyMap` interface:
```go
func (k KeyMap) ShortHelp() []key.Binding
func (k KeyMap) FullHelp() [][]key.Binding
```

### `internal/core/todofile.go`
Add for undo:
```go
func (tf *TodoFile) Snapshot() []string
func (tf *TodoFile) Restore(snapshot []string)
```

### `internal/tui/views/tasklist.go`
Add:
- `paginator paginator.Model`
- `animations map[int]*AnimationState` (for harmonica)

### `internal/tui/views/search.go`
- Enable fuzzy filtering on list
- Add preview panel
- Highlight matched characters

---

## New Files

```
internal/tui/
├── animation/
│   └── spring.go          # Harmonica spring helpers
├── help/
│   └── keymap.go          # Per-view keybindings for help
├── markdown/
│   └── renderer.go        # Glamour wrapper
├── undo/
│   └── stack.go           # Undo action stack
└── views/
    ├── statusbar.go       # Status bar component
    ├── toast.go           # Toast with countdown
    └── palette.go         # Command palette
```

---

## Message Types Summary

```go
// Help
type ToggleHelpMsg struct{}

// Undo/Toast
type ToastTickMsg struct{}
type UndoExecuteMsg struct{}

// Palette
type OpenPaletteMsg struct{}
type ClosePaletteMsg struct{}
type PaletteSelectMsg struct{ Action func() tea.Msg }

// Animation
type AnimationTickMsg struct{ Time time.Time }
```

---

## Settings Integration

Add to `config.Settings`:
```go
ShowStatusBar    bool  // Default: true
EnableAnimations bool  // Default: true
```

---

## Verification Plan

1. **Build**: `make build` succeeds
2. **Run**: `make run` or `./build/mdd`
3. **Test each feature**:
   - Press `?` - help overlay appears/disappears
   - Status bar shows at bottom with task count
   - Navigate long list - pagination dots appear
   - Search with typos - fuzzy matches work
   - Delete task - toast with undo countdown appears
   - Press `u` within 5s - task restored
   - Press `ctrl+k` - command palette opens
   - Toggle task - subtle animation plays
4. **Run tests**: `make test`
5. **Lint**: `make lint`
