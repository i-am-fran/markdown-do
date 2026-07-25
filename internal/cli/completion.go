package cli

import (
	"os"
	"sort"
	"strings"

	"github.com/i-am-fran/markdown-do/internal/core"
)

// allCommands is the full list of top-level command verbs, derived from the
// command-shape maps in args.go so it never drifts out of sync with them.
var allCommands = func() []string {
	seen := map[string]bool{}
	var out []string
	for _, m := range []map[string]bool{idArgCommands, zeroArgCommands, freeTextCommands} {
		for k := range m {
			if !seen[k] {
				seen[k] = true
				out = append(out, k)
			}
		}
	}
	sort.Strings(out)
	return out
}()

var configSubcommands = []string{"list", "get", "set", "edit"}

// Complete returns shell-completion candidates for the given words, where
// words is everything typed after "mdd" and the last element is the
// (possibly empty) word currently being completed. A nil/empty result means
// "no opinion" — the shell should fall back to its default completion.
func Complete(words []string) []string {
	if len(words) == 0 {
		return matchPrefix(allCommands, "")
	}

	cur := words[len(words)-1]
	prev := words[:len(words)-1]

	if strings.HasPrefix(cur, "@") {
		return sectionCandidates(cur[1:])
	}

	if len(prev) == 0 {
		return matchPrefix(allCommands, cur)
	}

	if prev[0] == "config" && len(prev) == 1 {
		return matchPrefix(configSubcommands, cur)
	}

	return nil
}

func matchPrefix(candidates []string, prefix string) []string {
	prefix = strings.ToLower(prefix)
	var out []string
	for _, c := range candidates {
		if strings.HasPrefix(strings.ToLower(c), prefix) {
			out = append(out, c)
		}
	}
	return out
}

// sectionCandidates returns "@Name" candidates matching prefix (the text
// typed after "@"): built-in/custom section aliases, plus section headers
// already present in the current directory's default TODO.md. Best-effort —
// any failure to locate/load a TODO.md just yields no file-based candidates.
func sectionCandidates(prefix string) []string {
	prefix = strings.ToLower(prefix)
	seen := map[string]bool{}
	var out []string

	add := func(name string) {
		key := strings.ToLower(name)
		if strings.HasPrefix(key, prefix) && !seen[key] {
			seen[key] = true
			out = append(out, "@"+name)
		}
	}

	for alias := range core.SectionAliases() {
		add(alias)
	}

	if cwd, err := os.Getwd(); err == nil {
		if path, err := core.FindDefaultTodoFile(cwd); err == nil {
			if tf, err := core.Load(path); err == nil {
				for _, name := range tf.GetSectionNames() {
					add(name)
				}
			}
		}
	}

	sort.Strings(out)
	return out
}
