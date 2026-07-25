# MarkdownDO

- [ ] [MDD23] Add a flag to add completed date (only when done via CLI) to completed tasks
- [ ] [MDD22] Remove TUI completely
- [ ] [MDD12] Add manual at `man mdd`
- [ ] [MDD07] Add some tests
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

## Bugs

- [ ] [MDD06] Allow to escape special characters i.e. `{` by typing `\{`
- [ ] [MDD03] If a task ends with `?` I get an error, unless I use quote marks
- [ ] [MDD18] Complete/edit/toggle can print the wrong task in their confirmation message when Save() reorders tasks (PerformCompleteTask/PerformEditTask/PerformToggleTask in logic.go call GetTask(id) after Save(), which can now belong to a different task)

## In Review / Ideas

- [ ] [MDD16] Allow to add tasks to todo files in subfolders
- [ ] [MDD05] Is it possible to add autocomplete suggestions? I want users to type `mdd -c ta` and hit `Tab` to get `mdd -c task`
- [x] [MDD20] Add an option to archive tasks (move them at the bottom of the file). Still use sections? Maybe as H3s?
- [x] [MDD02] Use `- [/]` to mark tasks as in progress
- [x] [MDD13] Add config options, for example to enable in-progress tasks, which is not default markdown. How?

## Admin

- [ ] [MDD01] Record intro video/demo
