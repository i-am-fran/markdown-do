---
title: Utilities
slug: utilities
order: 6
headings:
  - id: cmd-open
    label: open
  - id: cmd-lint
    label: lint
  - id: cmd-config
    label: config
  - id: cmd-completion
    label: completion
  - id: cmd-update
    label: update
  - id: cmd-version
    label: version
  - id: cmd-help
    label: help
---

### open {#cmd-open}

`mdd open` opens `TODO.md` in the editor set by the `editor` setting (see Configuration below).

`-o` is a short-hand alias for `open`.

```bash
mdd open
```

### lint {#cmd-lint}

`mdd lint` checks the file for formatting issues and fixes what it can safely fix automatically, printing a line number, description, and status (`found` or `fixed`) for everything it touched.

`-lint` is a short-hand alias for `lint`.

| Check | Behaviour |
| --- | --- |
| Empty sections | A `## Header` followed by nothing but blank lines is removed entirely |
| Heading spacing | Exactly one blank line enforced above and below every `## Header` |
| Blank lines | Runs of 2+ blank lines, including trailing ones at end of file, collapsed to one |
| Empty tasks | A `- [ ]` with no text after it is removed |
| Tasks under `## Notes` | Converted to plain `-` list items — checkboxes don't belong in a notes section |
| Empty checkbox | `- []` becomes `- [ ]` |
| Uppercase `X` | `- [X]` becomes `- [x]` |
| Malformed checkbox content | Anything else inside the brackets is normalized to whichever of `[ ]`/`[x]` it most resembles |

```bash
mdd lint
```

### config {#cmd-config}

`mdd config` reads and changes settings without hand-editing `config.json` — see Configuration below for the full settings table.

| Subcommand | Description |
| --- | --- |
| `list` | Print every setting and its current value |
| `get <key>` | Print one setting's value |
| `set <key> <value>` | Change a setting and persist it |
| `edit` | Open `config.json` directly in your configured editor, for changes the other three don't cover (e.g. hand-editing `sectionAliases` as JSON) |

```bash
mdd config list
mdd config get editor
mdd config set editor vim
mdd config set alias.wk Work
mdd config edit
```

<aside class="callout callout--tip not-prose" markdown="1">
**Tip:** `-h`/`--help` and `-v`/`--version` are recognized anywhere in the arguments and short-circuit immediately, before anything else is parsed — `mdd list -h` prints the general help, not a `list`-specific one, since there's no such thing.
</aside>

### completion {#cmd-completion}

`mdd completion bash` / `mdd completion zsh` print a shell completion script to stdout. It tab-completes command verbs (`mdd ar` → `mdd archive`) and section tags (`@Bu` → `@Bugs`, including your custom aliases and any section already in `TODO.md`).

```bash
source <(mdd completion bash)   # add to ~/.bashrc
source <(mdd completion zsh)    # add to ~/.zshrc
```

### update {#cmd-update}

`mdd update` downloads the latest release binary and swaps it in for the one currently running. The downloaded `checksums.txt` is verified against an ed25519 signature embedded in the binary before anything is replaced, so a compromised or tampered download is rejected rather than installed.

```bash
mdd update
```

### version {#cmd-version}

`mdd version` prints the current version. Same as passing `-v`/`--version` anywhere in the command.

```bash
mdd version
```

### help {#cmd-help}

`mdd help` prints the full command reference. Same as passing `-h`/`--help` anywhere in the command, or running `mdd` with no arguments at all.

```bash
mdd help
```
