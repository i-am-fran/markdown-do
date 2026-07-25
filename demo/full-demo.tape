# Full feature tour of mdd (markdown-do).
#
# Run from inside demo/:
#   cd demo && vhs full-demo.tape
#
# Isolates $HOME so `mdd config` writes never touch your real
# ~/.config/markdowndo/config.json, and prepends $HOME/go/bin (where
# `make install` puts the freshly built binary) onto PATH so the recording
# always uses current source, not whatever stale `mdd` is on your real PATH.

Require mdd

Set Shell "bash"
Set FontSize 16
Set Width 1200
Set Height 700
Set Padding 20
Set Theme "Catppuccin Mocha"
Set TypingSpeed 35ms

Output full-demo.gif
Output full-demo.mp4

Hide
Type `export PATH="$HOME/go/bin:$PATH"`
Enter
Type `export HOME="$(mktemp -d)"`
Enter
Type `export PS1="$ "`
Enter
Type "clear"
Enter
Show

# --- Start: a real TODO.md, three checkbox states already in play ---
Type "mdd list"
Sleep 800ms
Enter
Sleep 2s

# --- Lint keeps formatting honest ---
Type "# mdd lint fixes formatting slips automatically"
Sleep 800ms
Enter
Sleep 600ms
Type "mdd lint"
Sleep 800ms
Enter
Sleep 2s

# --- Settings live in ~/.config/markdowndo/config.json ---
Type "mdd config list"
Sleep 800ms
Enter
Sleep 1200ms
Type "mdd config set enableInProgress true"
Sleep 800ms
Enter
Sleep 500ms
Type "mdd config set confirmDestructive true"
Sleep 800ms
Enter
Sleep 500ms
Type `mdd config set alias.admin Admin`
Sleep 800ms
Enter
Sleep 1200ms

# --- Adding tasks: implicit, section shortcuts, a custom alias, explicit add ---
Type "mdd Ship the Q3 newsletter"
Sleep 800ms
Enter
Sleep 500ms
Type `mdd "Fix broken image links @bb"`
Sleep 800ms
Enter
Sleep 500ms
Type `mdd "Voice search @ff"`
Sleep 800ms
Enter
Sleep 500ms
Type `mdd "Community forum @ii"`
Sleep 800ms
Enter
Sleep 500ms
Type `mdd "Sunset legacy exports @ww"`
Sleep 800ms
Enter
Sleep 500ms
Type `mdd "Rotate API keys @admin"`
Sleep 800ms
Enter
Sleep 500ms
Type `mdd add "Write release notes"`
Sleep 800ms
Enter
Sleep 500ms
Type "mdd list"
Sleep 800ms
Enter
Sleep 2500ms

# --- Finding things ---
Type "mdd find login"
Sleep 800ms
Enter
Sleep 1500ms

# --- Toggle: pending -> in-progress -> completed ---
Type "mdd toggle 5"
Sleep 800ms
Enter
Sleep 1500ms

# --- Completing tasks ---
Type "mdd complete 3"
Sleep 800ms
Enter
Sleep 500ms
Type "mdd complete 3"
Sleep 800ms
Enter
Sleep 1500ms

# --- Editing and annotating ---
Type `mdd edit 9 "Login button unresponsive on Safari and Firefox"`
Sleep 800ms
Enter
Sleep 500ms
Type `mdd annotate 10 "Needs design review before shipping"`
Sleep 800ms
Enter
Sleep 500ms
Type `mdd notes "Standup moved to 10am starting Monday"`
Sleep 800ms
Enter
Sleep 1500ms

# --- Clearing completed tasks ---
Type "mdd complete 14"
Sleep 800ms
Enter
Sleep 500ms
Type "mdd clear -y"
Sleep 800ms
Enter
Sleep 1500ms
Type "mdd list"
Sleep 800ms
Enter
Sleep 2000ms

# --- Archiving ---
Type "mdd complete 5"
Sleep 800ms
Enter
Sleep 300ms
Type "mdd complete 10"
Sleep 800ms
Enter
Sleep 500ms
Type "mdd archive"
Sleep 800ms
Enter
Sleep 1500ms
Type "mdd list"
Sleep 800ms
Enter
Sleep 2500ms

# --- Stable IDs: tag, complete/reference by ID, untag, undo ---
Type "mdd tag MDD"
Sleep 800ms
Enter
Sleep 1500ms
Type "mdd complete MDD-006"
Sleep 800ms
Enter
Sleep 1000ms
Type "mdd untag"
Sleep 800ms
Enter
Sleep 1000ms
Type "mdd undo"
Sleep 800ms
Enter
Sleep 1500ms

# --- Destructive actions ask first when confirmDestructive is on ---
Type "mdd remove MDD-003"
Sleep 800ms
Enter
Sleep 800ms
Type "y"
Sleep 500ms
Enter
Sleep 1500ms

# --- Recursive across subdirectories ---
Type "mdd list -r"
Sleep 800ms
Enter
Sleep 2500ms
Type "mdd find offline -r"
Sleep 800ms
Enter
Sleep 1500ms

# --- Shell completion ---
Type "mdd completion bash | head -n 5"
Sleep 800ms
Enter
Sleep 1800ms

# --- Open in your editor (EDITOR=cat here, so it's visible instead of blocking) ---
Hide
Type "export EDITOR=cat"
Enter
Show
Type "mdd open"
Sleep 800ms
Enter
Sleep 2500ms

# --- Staying current ---
Type "# mdd update installs the newest release via go install (skipped here)"
Sleep 800ms
Enter
Sleep 800ms

Type "mdd version"
Sleep 800ms
Enter
Sleep 2s
