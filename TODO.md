# MarkdownDO

- [ ] [MDD26] Wire up PerformMoveTask as mdd move, or delete it if not worth it — currently dead code only reachable by its own tests since the TUI's move-to-section hotkey that used it is gone
- [ ] [MDD25] rewrite in python? it's now in go (I think) but does it still makes sense since we don't have a tui? will it cause issues if we want  to add a tui later on?
- [ ] [MDD23] Add a flag to add completed date (only when done via CLI) to completed tasks
- [ ] [MDD12] Add manual at `man mdd`
- [ ] [MDD07] Add some tests
- [x] [MDD22] Remove TUI completely

## Bugs

- [ ] [MDD06] Allow to escape special characters i.e. `{` by typing `\{`
- [ ] [MDD03] If a task ends with `?` I get an error, unless I use quote marks
- [ ] [MDD18] Complete/edit/toggle can print the wrong task in their confirmation message when Save() reorders tasks (PerformCompleteTask/PerformEditTask/PerformToggleTask in logic.go call GetTask(id) after Save(), which can now belong to a different task)

## In Review / Ideas

- [ ] [MDD16] Allow to add tasks to todo files in subfolders
- [ ] [MDD05] Is it possible to add autocomplete suggestions? I want users to type `mdd -c ta` and hit `Tab` to get `mdd -c task`

## Admin

- [ ] [MDD01] Record intro video/demo
## Ideas

- [ ] [MDD24] Allow to filter mdd list archive/all/todo or something like that


## Archive

- [x] [MDD19] Rebuild website and documentation website
- [x] [MDD21] Create website and documentation using web-docs-template
- [x] [MDD15] Deprecate TUI for now
- [x] [MDD14] Replace delete, use `-r` instead of `-d`
- [x] [MDD17] Simplify all the commands
- [x] [MDD11] Add instructions to update and versioning. Versioning should be global claude instructions
- [x] [MDD08] Decouple CLI and TUI in two separate packages
- [x] [MDD10] Allow users to add unique IDs to tasks, so it's easier to reference them when using LLMs
- [x] [MDD09] Improve `--help` menu. it should contain more examples with default stuff like bugs and features etc.
- [x] [MDD04] Add option to add notes to tasks by having a bullet list follow a checklist item. this needs to be readable from the CLI as well
- [x] [MDD20] Add an option to archive tasks (move them at the bottom of the file). Still use sections? Maybe as H3s? (from In Review / Ideas)
- [x] [MDD02] Use `- [/]` to mark tasks as in progress (from In Review / Ideas)
- [x] [MDD13] Add config options, for example to enable in-progress tasks, which is not default markdown. How? (from In Review / Ideas)

