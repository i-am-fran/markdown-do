# Markdown-do

A fast, minimal, opinionated CLI tool for managing TODO.md files using standard markdown syntax.

![mdd demo](demo/essential-demo.gif)

## Features

- **Plain markdown** - Tasks are standard `- [ ]` checkboxes, readable anywhere
- **Sections** - Organize tasks with `## Headers`, quick-add with `@aliases`
- **Fast CLI** - Add, toggle, complete, find, and manage tasks without leaving your terminal
- **Stable task IDs** - Tag tasks for reliable reference across sessions
- **Recursive search** - Find tasks across all TODO files in subdirectories
- **Lint & fix** - Auto-fix common formatting issues
- **No lock-in** - Your tasks are plain markdown, version-controlled with your code

## Installation

### From releases

Download the latest binary for your platform from the [releases page](https://github.com/i-am-fran/markdown-do/releases).

### From source

```bash
git clone https://github.com/i-am-fran/markdown-do.git
cd markdown-do
make build
# Binary is at ./build/mdd

# Or install to GOPATH/bin:
make install
```

### Go install

```bash
go install github.com/i-am-fran/markdown-do/cmd/mdd@<version>
```

Replace `<version>` with the latest tag from the [releases page](https://github.com/i-am-fran/markdown-do/releases) (e.g. `@3.2.0`) — plain `@latest` doesn't resolve correctly here, since releases are tagged without a leading `v`. Once you have `mdd` installed, `mdd update` looks up and installs the latest release for you.

## Docs & full command reference

**[i-am-fran.github.io/markdown-do](https://i-am-fran.github.io/markdown-do/)** — every command, flag, config setting, and task format detail.

## License

MIT, see [LICENSE](LICENSE).
