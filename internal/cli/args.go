package cli

import (
	"regexp"
	"strings"
)

// ParsedArgs is the result of parsing CLI arguments into a command and its
// arguments. Command is "" only when the input was empty or contained
// nothing but flags.
type ParsedArgs struct {
	Command      string   // "add", a known verb, or "unknown" for an unrecognized dash flag
	Args         []string // remaining positional tokens (the command word removed)
	Text         string   // Args joined with " "
	Recursive    bool     // -r / --recursive present anywhere in the input
	Yes          bool     // -y / --yes present anywhere in the input
	StatusFilter string   // "active" (-a/--active), "done" (--done), or "" for no filter; used by list
	Path         string   // -p / --path value; directory to target instead of cwd, used by add
}

// idShapeRegex matches tokens that look like a task reference: a position
// number, or a stable ID tag (e.g. "ABC-001", matching the pattern used by
// core.idTagRegex/stableIDRegex). The dash is optional to also match IDs
// tagged before MDD29 (e.g. "ABC01").
var idShapeRegex = regexp.MustCompile(`^(?:\d+|[A-Za-z]{3}-?\d+)$`)

// idArgCommands are only recognized as that command when their first
// argument is ID-shaped; otherwise the whole input falls through to an
// implicit "add" so that task text like "Complete the tax return" isn't
// misinterpreted as a command.
var idArgCommands = map[string]bool{
	"complete": true,
	"toggle":   true,
	"edit":     true,
	"remove":   true,
	"annotate": true,
}

// zeroArgCommands are only recognized as that command when nothing follows
// them (other than -r/--recursive/-a/--active/--done, already stripped by
// the time this check runs); otherwise the input falls through to an
// implicit "add" so that task text like "List of birthday gifts" isn't
// misinterpreted.
var zeroArgCommands = map[string]bool{
	"list":    true,
	"open":    true,
	"lint":    true,
	"clear":   true,
	"archive": true,
	"untag":   true,
	"help":    true,
	"version": true,
	"undo":    true,
	"update":  true,
}

// freeTextCommands take arbitrary trailing text, so there's no shape to
// disambiguate against — they're recognized purely by their first word.
var freeTextCommands = map[string]bool{
	"find":       true,
	"notes":      true,
	"tag":        true,
	"add":        true,
	"config":     true,
	"completion": true,
}

// ParseArgs parses command line arguments into a command and its arguments.
func ParseArgs(args []string) ParsedArgs {
	for _, a := range args {
		if a == "-h" || a == "--help" {
			return ParsedArgs{Command: "help"}
		}
		if a == "-v" || a == "--version" {
			return ParsedArgs{Command: "version"}
		}
	}

	var rest []string
	recursive := false
	yes := false
	statusFilter := ""
	path := ""
	malformedFlag := ""
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "-r" || a == "--recursive" {
			recursive = true
			continue
		}
		if a == "-y" || a == "--yes" {
			yes = true
			continue
		}
		if a == "-a" || a == "--active" {
			if statusFilter != "" {
				malformedFlag = a
			}
			statusFilter = "active"
			continue
		}
		if a == "--done" {
			if statusFilter != "" {
				malformedFlag = a
			}
			statusFilter = "done"
			continue
		}
		if a == "-p" || a == "--path" {
			if i+1 >= len(args) {
				malformedFlag = a
				continue
			}
			i++
			path = args[i]
			continue
		}
		rest = append(rest, a)
	}

	if malformedFlag != "" {
		return ParsedArgs{Command: "unknown", Args: []string{malformedFlag}, Text: malformedFlag}
	}

	if len(rest) == 0 {
		return ParsedArgs{Recursive: recursive, Yes: yes, StatusFilter: statusFilter, Path: path}
	}

	word := rest[0]
	tail := rest[1:]

	switch {
	case idArgCommands[word] && len(tail) > 0 && idShapeRegex.MatchString(tail[0]):
		return newParsedArgs(word, tail, recursive, yes, statusFilter, path)
	case zeroArgCommands[word] && len(tail) == 0:
		return newParsedArgs(word, tail, recursive, yes, statusFilter, path)
	case freeTextCommands[word]:
		return newParsedArgs(word, tail, recursive, yes, statusFilter, path)
	case strings.HasPrefix(word, "-"):
		// An unrecognized dash flag (likely a pre-2.0 flag like -c/-d/-lint)
		// is reported loudly instead of being silently added as task text.
		return newParsedArgs("unknown", rest, recursive, yes, statusFilter, path)
	default:
		return newParsedArgs("add", rest, recursive, yes, statusFilter, path)
	}
}

func newParsedArgs(command string, args []string, recursive, yes bool, statusFilter, path string) ParsedArgs {
	return ParsedArgs{
		Command:      command,
		Args:         args,
		Text:         strings.Join(args, " "),
		Recursive:    recursive,
		Yes:          yes,
		StatusFilter: statusFilter,
		Path:         path,
	}
}
