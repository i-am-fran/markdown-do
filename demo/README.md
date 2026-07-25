# mdd VHS demos

Two [VHS](https://github.com/charmbracelet/vhs) recordings of `mdd` in action, both operating on the `TODO.md` fixture in this folder:

- **`full-demo`** — a tour of essentially every command: add (implicit, `@section`, custom aliases, explicit), list, find, toggle, complete, edit, annotate, notes, lint, clear, archive, tag/untag, undo, remove (with the confirm prompt), recursive list/find, shell completion, config, and open.
- **`essential-demo`** — just the everyday loop: add, list, toggle, complete, edit, find, remove.

`projects/mobile-app/TODO.md` is a second fixture used only by the recursive (`-r`) scene in `full-demo`.

## Re-recording

```bash
make install          # from the repo root — puts a current build at $(go env GOPATH)/bin/mdd
cd demo
vhs full-demo.tape
vhs essential-demo.tape
```

Both tapes mutate `TODO.md` as they run (they add/complete/remove/tag tasks live), so re-running a tape leaves the fixture in whatever state that recording ends on. Restore it from git afterward:

```bash
git checkout -- TODO.md
```

Each tape isolates `$HOME` to a scratch directory for the duration of the recording, so `mdd config` writes never touch your real `~/.config/markdowndo/config.json`, and prepends `$(go env GOPATH)/bin` onto `PATH` so the recording always uses the binary `make install` just built rather than whatever `mdd` happens to be on your normal `PATH`.

Requires `vhs`, `ttyd`, and `ffmpeg` (`brew install vhs ttyd ffmpeg`).
