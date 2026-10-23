# Markdown-do

[![A terminal recording of mdd adding, listing, toggling, completing, editing, finding and removing tasks in a TODO.md file. Click to watch it as a video](demo/essential-demo.gif)](demo/essential-demo.mp4)

Markdown-do (`mdd`) manages the `TODO.md` file in your project from the terminal. Tasks are plain markdown checkboxes, so the file stays readable in any editor and you commit it with the rest of your code.

- **[Install](#installation)**
- **[Documentation](https://i-am-fran.github.io/markdown-do/)**
- **[Report a bug](https://github.com/i-am-fran/markdown-do/issues)**

Markdown-do is minimal on purpose. There's no database, no app and no account: just one file per project, next to the code it's about. If you stop using `mdd` tomorrow, nothing breaks. You'll still have a `TODO.md` you can read and edit by hand.

It's also easy to hand to a coding agent. The commands are short and the output is plain text, so an agent can read the list, add what it finds and tick off what it finishes, and you can see every change in `git diff`.

## How it works

Run `mdd` followed by the task, in any folder. It finds the `TODO.md` there, or creates one:

```bash
mdd Reply to Sam about the invoice
mdd "Fix the date parser on leap years @bb"
mdd list
mdd complete 1
```

`@bb` files the task under Bugs. The file looks like this:

```markdown
# TODO

- [x] Reply to Sam about the invoice

## Bugs

- [ ] Fix the date parser on leap years
```

## Features

**Plain markdown.** Tasks are `- [ ]` and `- [x]` checkboxes under `## Sections`. New tasks land in an inbox at the top unless you end them with `@Section`, or a shortcut like `@bb` for Bugs and `@ff` for Features. On every save, open tasks move to the top of their section.

**Short commands.** Add, list, find, toggle, complete, edit and remove, each with a short alias (`-l`, `-c`, `-e`) for the ones you type every day. `mdd undo` reverts the last change, and `mdd annotate` adds a note under a task that stays with it.

**Every project at once.** `mdd list -r` and `mdd find -r` look through the `TODO.md` in every subfolder, so you can see what's open across all your projects and complete a task from anywhere.

**Task IDs that don't move.** `mdd tag ABC` gives every task an ID like `ABC-001`. Position numbers change as tasks are sorted and removed, but `mdd complete ABC-001` still works tomorrow, or in another agent session.

**Tidy files.** `mdd lint` fixes stray blank lines, `[X]` checkboxes and empty sections. `mdd archive` moves finished tasks to an `## Archive` section and notes which section each one came from.

The documentation has [every feature, including the little ones](https://i-am-fran.github.io/markdown-do/docs.html#all-features).

## Installation

With Go installed:

```bash
go install github.com/i-am-fran/markdown-do/v3/cmd/mdd@latest
```

Without Go:

1. Download the binary for your system from the [latest release](https://github.com/i-am-fran/markdown-do/releases/latest), for example `mdd-darwin-arm64` for an Apple Silicon Mac.
2. Rename it to `mdd`, make it executable (`chmod +x mdd`) and move it to a folder on your `PATH`.

After that, `mdd update` downloads the latest release and replaces the binary for you. It checks the release's signature before it replaces anything.

## Settings

`mdd config set` changes how lists look, which editor `mdd open` uses, whether toggling goes through an in-progress `[/]` state, whether deleting asks first, and your own `@section` shortcuts. The documentation lists [every setting and its default](https://i-am-fran.github.io/markdown-do/docs.html#settings).

## Support

Markdown-do is free. If it helps you, you can buy me a coffee:

<a href="https://www.buymeacoffee.com/iamfran" target="_blank"><img src="https://cdn.buymeacoffee.com/buttons/v2/default-yellow.png" alt="Buy Me a Coffee" height="60" width="217"></a>

## Development

```bash
make build   # build to ./build/mdd
make test    # run the tests
make lint    # golangci-lint
```

Changes are listed in the [changelog](CHANGELOG.md).

## License

MIT. See [LICENSE](LICENSE).
