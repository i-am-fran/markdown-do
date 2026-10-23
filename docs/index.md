---
layout: home
sections:
  - id: how-it-works
    label: How it works
  - id: features
    label: Features
  - id: install
    label: Install
  - id: explore
    label: Explore
---
<p class="lede">{{ site.tagline }}</p>

<div class="actions">
  <a class="button button--primary" href="{{ '/docs.html#installation' | relative_url }}">Install{% include arrow-icon.html %}</a>
  <a class="button" href="{{ site.github_url }}">View on GitHub</a>
</div>

<img src="{{ '/assets/essential-demo.gif' | relative_url }}" alt="A terminal recording of mdd adding, listing, toggling, completing, editing, finding and removing tasks in a TODO.md file" class="showcase">

## How it works {#how-it-works}

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

Markdown-do is minimal on purpose. There's no database, no app and no account: just one file per project, next to the code it's about. If you stop using `mdd` tomorrow, nothing breaks. You'll still have a `TODO.md` you can read and edit by hand.

It's also easy to hand to a coding agent. The commands are short and the output is plain text, so an agent can read the list, add what it finds and tick off what it finishes, and you can see every change in `git diff`.

***

## Features {#features}

**Plain markdown.** Tasks are `- [ ]` and `- [x]` checkboxes under `## Sections`. New tasks land in an inbox at the top unless you end them with `@Section`, or a shortcut like `@bb` for Bugs. On every save, open tasks move to the top of their section.

**Short commands.** Add, list, find, toggle, complete, edit and remove, each with a short alias (`-l`, `-c`, `-e`) for the ones you type every day. `mdd undo` reverts the last change.

**Every project at once.** `mdd list -r` and `mdd find -r` look through the `TODO.md` in every subfolder, so you can see what's open across all your projects and complete a task from anywhere.

**Task IDs that don't move.** `mdd tag ABC` gives every task an ID like `ABC-001` that still works tomorrow, or in another agent session.

**Tidy files.** `mdd lint` fixes stray blank lines, `[X]` checkboxes and empty sections. `mdd archive` moves finished tasks out of the way without deleting them.

[Every feature, including the little ones]({{ '/docs.html#all-features' | relative_url }}).

***

## Install {#install}

With Go installed:

```bash
go install github.com/i-am-fran/markdown-do/v3/cmd/mdd@latest
```

Without Go, download the binary for your system from the [latest release](https://github.com/i-am-fran/markdown-do/releases/latest), rename it to `mdd` and put it on your `PATH`. After that, `mdd update` keeps it up to date.

***

<div class="section-head">
  <h2 id="explore">Explore</h2>
  <a class="pill" href="{{ '/docs.html' | relative_url }}">All docs</a>
</div>

<ul class="arrow-list">
  <li><a class="arrow-card" href="{{ '/docs.html#getting-started' | relative_url }}"><span class="arrow-card__text"><span class="arrow-card__title">Getting started</span><span class="arrow-card__desc">Install, first tasks and the file format.</span></span>{% include arrow-icon.html %}</a></li>
  <li><a class="arrow-card" href="{{ '/docs.html#adding' | relative_url }}"><span class="arrow-card__text"><span class="arrow-card__title">Adding tasks</span><span class="arrow-card__desc">The inbox, sections, shortcuts and notes.</span></span>{% include arrow-icon.html %}</a></li>
  <li><a class="arrow-card" href="{{ '/docs.html#viewing' | relative_url }}"><span class="arrow-card__text"><span class="arrow-card__title">Viewing and finding</span><span class="arrow-card__desc">List and search one file or every subfolder.</span></span>{% include arrow-icon.html %}</a></li>
  <li><a class="arrow-card" href="{{ '/docs.html#managing' | relative_url }}"><span class="arrow-card__text"><span class="arrow-card__title">Managing tasks</span><span class="arrow-card__desc">Toggle, complete, edit, annotate, remove, archive, undo.</span></span>{% include arrow-icon.html %}</a></li>
  <li><a class="arrow-card" href="{{ '/docs.html#task-ids' | relative_url }}"><span class="arrow-card__text"><span class="arrow-card__title">Task IDs</span><span class="arrow-card__desc">IDs that stay put when tasks move.</span></span>{% include arrow-icon.html %}</a></li>
  <li><a class="arrow-card" href="{{ '/docs.html#configuration' | relative_url }}"><span class="arrow-card__text"><span class="arrow-card__title">Settings</span><span class="arrow-card__desc">Every setting and its default.</span></span>{% include arrow-icon.html %}</a></li>
  <li><a class="arrow-card" href="{{ '/docs.html#all-features' | relative_url }}"><span class="arrow-card__text"><span class="arrow-card__title">All features</span><span class="arrow-card__desc">Every feature, including the little ones.</span></span>{% include arrow-icon.html %}</a></li>
  <li><a class="arrow-card" href="{{ site.github_url }}" target="_blank" rel="noopener noreferrer"><span class="arrow-card__text"><span class="arrow-card__title">GitHub</span><span class="arrow-card__desc">Source, releases and issues.</span></span>{% include arrow-icon.html %}</a></li>
</ul>
