# Markdown-do

A fast, minimal, opinionated CLI tool for managing TODO.md files using standard markdown syntax.

![mdd demo](demo/essential-demo.gif)

<script type="text/javascript" src="https://cdnjs.buymeacoffee.com/1.0.0/button.prod.min.js" data-name="bmc-button" data-slug="iamfran" data-color="#FFDD00" data-emoji=""  data-font="Cookie" data-text="Buy me a coffee" data-outline-color="#000000" data-font-color="#000000" data-coffee-color="#ffffff" ></script>

## Features

- **Plain markdown** - Tasks are standard `- [ ]` checkboxes, readable anywhere
- **Sections** - Organize tasks with `## Headers`, quick-add with `@aliases`
- **Fast CLI** - Add, toggle, complete, find, and manage tasks without leaving your terminal, with short-hand aliases (`-c`, `-l`, `-e`, ...) for everyday commands
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
go install github.com/i-am-fran/markdown-do/v3/cmd/mdd@latest
```

Once you have `mdd` installed, `mdd update` downloads and swaps in the latest release binary for you (it needs write access to wherever `mdd` currently lives, not Go).

## Docs & full command reference

**[i-am-fran.github.io/markdown-do](https://i-am-fran.github.io/markdown-do/)** — every command, flag, config setting, and task format detail.

## License

MIT, see [LICENSE](LICENSE).
