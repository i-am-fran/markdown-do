# MarkdownDO

- [ ] [MDD-021] Add UNDO option with `mdd undo`
- [ ] [MDD-030] Add `mdd update` to download and install the latest release from git repo
- [ ] [MDD-029] Wire up PerformMoveTask as mdd move, or delete it if not worth it — currently dead code only reachable by its own tests since the TUI's move-to-section hotkey that used it is gone (from AI)
- [ ] [MDD-028] Add manual at `man mdd`
- [ ] [MDD-027] Add some tests
- [x] [MDD-016] Is it possible to add autocomplete suggestions? I want users to type `mdd ar` and hit `Tab` to get `mdd archive`. This should also find and suggest sections (H2s) in the TODO.md file

## Bugs

- [ ] [MDD-024] Allow to escape special characters i.e. `{` by typing `\{`
- [ ] [MDD-023] If a task ends with `?` I get an error, unless I use quote marks
- [ ] [MDD-022] Complete/edit/toggle can print the wrong task in their confirmation message when Save() reorders tasks (PerformCompleteTask/PerformEditTask/PerformToggleTask in logic.go call GetTask(id) after Save(), which can now belong to a different task)

## Ideas

- [ ] [MDD-020] Add a flag to add completed date (only when done via CLI) to completed tasks
- [ ] [MDD-019] Add an option to sync with Github tasks??
- [ ] [MDD-018] Allow to filter mdd list archive/all/todo or something like that
- [ ] [MDD-017] Allow to add tasks to todo files in subfolders

## Admin

- [ ] [MDD-015] Record intro video/demo

## Archive

- [x] [MDD-014] Rebuild website and documentation website
- [x] [MDD-013] Create website and documentation using web-docs-template
- [x] [MDD-012] Deprecate TUI for now
- [x] [MDD-011] Replace delete, use `-r` instead of `-d`
- [x] [MDD-010] Simplify all the commands
- [x] [MDD-009] Add instructions to update and versioning. Versioning should be global claude instructions
- [x] [MDD-008] Decouple CLI and TUI in two separate packages
- [x] [MDD-007] Allow users to add unique IDs to tasks, so it's easier to reference them when using LLMs
- [x] [MDD-006] Improve `--help` menu. it should contain more examples with default stuff like bugs and features etc.
- [x] [MDD-005] Add option to add notes to tasks by having a bullet list follow a checklist item. this needs to be readable from the CLI as well
- [x] [MDD-004] Add an option to archive tasks (move them at the bottom of the file). Still use sections? Maybe as H3s? (from In Review / Ideas)
- [x] [MDD-003] Use `- [/]` to mark tasks as in progress (from In Review / Ideas)
- [x] [MDD-002] Add config options, for example to enable in-progress tasks, which is not default markdown. How? (from In Review / Ideas)
- [x] [MDD-001] Remove TUI completely
- [x] [MDD-026] What happens when issues go over 99? Should we have 3 digits from the beginning, starting with 001?
- [x] [MDD-025] IDs should include a separator, as in ABC-001
