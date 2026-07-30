package cli

import (
	"os"
	"sort"
	"strings"

	"github.com/i-am-fran/markdown-do/v3/internal/core"
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

// levenshtein returns the optimal string alignment distance between a and
// b: insertions, deletions, and substitutions, plus adjacent transpositions
// counted as a single edit (not two substitutions). Transpositions are the
// single most common real-world typing mistake (e.g. "toggel" for
// "toggle"), so counting them as one edit rather than two both matches
// typing intuition and — critically — breaks ties in suggestCommand's
// favor of the transposed word over an unrelated one that happens to sit
// at the same plain-Levenshtein distance (e.g. "lsit" is distance 2 from
// both "list" (a transposition) and "edit"/"lint" (unrelated) under plain
// Levenshtein, but distance 1 from "list" and still 2 from the others
// here).
func levenshtein(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	la, lb := len(ra), len(rb)

	d := make([][]int, la+1)
	for i := range d {
		d[i] = make([]int, lb+1)
		d[i][0] = i
	}
	for j := 0; j <= lb; j++ {
		d[0][j] = j
	}

	for i := 1; i <= la; i++ {
		for j := 1; j <= lb; j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			d[i][j] = min(d[i-1][j]+1, d[i][j-1]+1, d[i-1][j-1]+cost)
			if i > 1 && j > 1 && ra[i-1] == rb[j-2] && ra[i-2] == rb[j-1] {
				d[i][j] = min(d[i][j], d[i-2][j-2]+1)
			}
		}
	}
	return d[la][lb]
}

// fuzzyMatchMinLen is the shortest candidate command length considered for
// the edit-distance check in suggestCommand. Below this, almost any
// unrelated one-word task (e.g. "Tap"/"Tar" vs "tag") sits within the
// distance threshold, so short commands ("tag", "add") are only ever
// suggested on an exact (case-insensitive) match, never a fuzzy one.
const fuzzyMatchMinLen = 4

// fuzzyMatchMaxDist is the maximum edit distance treated as a likely typo.
// Every typo this guard needs to catch (a single wrong/missing/extra
// letter, or a single adjacent transposition like "toggel"/"compelte") is
// distance 1 under levenshtein's transposition-aware metric, so 1 is both
// sufficient and deliberately tight — it minimizes false positives against
// unrelated one-word tasks.
const fuzzyMatchMaxDist = 1

// suggestCommand looks for a known command close to word, for use when word
// was the CLI's entire input (a single bare word that didn't already
// resolve to a real command). An exact match (modulo case) always counts as
// a suggestion regardless of edit distance — this covers both a
// wrong-case command name and a correctly-spelled command missing its
// required argument (e.g. bare "complete", which never reaches the
// idArgCommands shape check since there's no tail to inspect). Beyond
// that, commands of fuzzyMatchMinLen letters or more get a small
// edit-distance check.
func suggestCommand(word string) (string, bool) {
	lower := strings.ToLower(word)
	for _, c := range allCommands {
		if c == lower {
			return c, true
		}
	}

	best, bestDist := "", -1
	for _, c := range allCommands {
		if len(c) < fuzzyMatchMinLen {
			continue
		}
		if d := levenshtein(lower, c); d <= fuzzyMatchMaxDist && (bestDist == -1 || d < bestDist) {
			best, bestDist = c, d
		}
	}
	return best, best != ""
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
