---
title: More
slug: more
order: 9
headings:
  - id: limits
    label: Limits
  - id: support
    label: Support
---

### Limits {#limits}

- `mdd` reads and writes one kind of file: markdown checkboxes under `##` headings. Other markdown in the file is kept, but `mdd lint` may tidy blank lines around it.
- Saves are atomic, but two `mdd` commands started at the same moment on the same file can both read it before either writes. The second save wins.
- `undo` keeps one step. Running it twice puts the change back.

### Support {#support}

Markdown-do is free. If it helps you, you can [buy me a coffee](https://www.buymeacoffee.com/iamfran). Found a bug? Please [open an issue](https://github.com/i-am-fran/markdown-do/issues).

Changes are listed in the [changelog](https://github.com/i-am-fran/markdown-do/blob/main/CHANGELOG.md). Markdown-do is [MIT licensed](https://github.com/i-am-fran/markdown-do/blob/main/LICENSE).
