package cli

import (
	"reflect"
	"testing"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want ParsedArgs
	}{
		// idArgCommands: recognized only when the next token is ID-shaped.
		{"complete with numeric id", []string{"complete", "2"}, ParsedArgs{Command: "complete", Args: []string{"2"}, Text: "2"}},
		{"complete with stable id", []string{"complete", "ABC-001"}, ParsedArgs{Command: "complete", Args: []string{"ABC-001"}, Text: "ABC-001"}},
		{"complete with legacy id", []string{"complete", "ABC01"}, ParsedArgs{Command: "complete", Args: []string{"ABC01"}, Text: "ABC01"}},
		{"complete with lowercase stable id", []string{"complete", "abc-001"}, ParsedArgs{Command: "complete", Args: []string{"abc-001"}, Text: "abc-001"}},
		{"complete falls through to add", []string{"complete", "the", "tax", "return"}, ParsedArgs{Command: "add", Args: []string{"complete", "the", "tax", "return"}, Text: "complete the tax return"}},
		{"toggle with numeric id", []string{"toggle", "3"}, ParsedArgs{Command: "toggle", Args: []string{"3"}, Text: "3"}},
		{"edit with numeric id and text", []string{"edit", "1", "New", "text"}, ParsedArgs{Command: "edit", Args: []string{"1", "New", "text"}, Text: "1 New text"}},
		{"remove with numeric id", []string{"remove", "5"}, ParsedArgs{Command: "remove", Args: []string{"5"}, Text: "5"}},
		{"annotate with numeric id", []string{"annotate", "1", "note", "text"}, ParsedArgs{Command: "annotate", Args: []string{"1", "note", "text"}, Text: "1 note text"}},
		{"bare complete (missing id) is a typo suggesting complete", []string{"complete"}, ParsedArgs{Command: "typo", Args: []string{"complete"}, Text: "complete", Suggestion: "complete"}},

		// zeroArgCommands: recognized only when the tail is empty.
		{"list alone", []string{"list"}, ParsedArgs{Command: "list", Args: []string{}, Text: ""}},
		{"open alone", []string{"open"}, ParsedArgs{Command: "open", Args: []string{}, Text: ""}},
		{"lint alone", []string{"lint"}, ParsedArgs{Command: "lint", Args: []string{}, Text: ""}},
		{"clear alone", []string{"clear"}, ParsedArgs{Command: "clear", Args: []string{}, Text: ""}},
		{"archive alone", []string{"archive"}, ParsedArgs{Command: "archive", Args: []string{}, Text: ""}},
		{"untag alone", []string{"untag"}, ParsedArgs{Command: "untag", Args: []string{}, Text: ""}},
		{"help alone", []string{"help"}, ParsedArgs{Command: "help", Args: []string{}, Text: ""}},
		{"version alone", []string{"version"}, ParsedArgs{Command: "version", Args: []string{}, Text: ""}},
		{"undo alone", []string{"undo"}, ParsedArgs{Command: "undo", Args: []string{}, Text: ""}},
		{"update alone", []string{"update"}, ParsedArgs{Command: "update", Args: []string{}, Text: ""}},
		{"list with trailing text falls through to add", []string{"list", "of", "birthday", "gifts"}, ParsedArgs{Command: "add", Args: []string{"list", "of", "birthday", "gifts"}, Text: "list of birthday gifts"}},
		{"list with -r stays list", []string{"list", "-r"}, ParsedArgs{Command: "list", Args: []string{}, Text: "", Recursive: true}},

		// freeTextCommands: always recognized by first word.
		{"find with keyword", []string{"find", "bug"}, ParsedArgs{Command: "find", Args: []string{"bug"}, Text: "bug"}},
		{"find with no keyword", []string{"find"}, ParsedArgs{Command: "find", Args: []string{}, Text: ""}},
		{"notes with text", []string{"notes", "remember", "this"}, ParsedArgs{Command: "notes", Args: []string{"remember", "this"}, Text: "remember this"}},
		{"tag with prefix", []string{"tag", "ABC"}, ParsedArgs{Command: "tag", Args: []string{"ABC"}, Text: "ABC"}},
		{"add with text", []string{"add", "Buy", "milk"}, ParsedArgs{Command: "add", Args: []string{"Buy", "milk"}, Text: "Buy milk"}},
		{"config with subcommand", []string{"config", "list"}, ParsedArgs{Command: "config", Args: []string{"list"}, Text: "list"}},
		{"completion with shell", []string{"completion", "bash"}, ParsedArgs{Command: "completion", Args: []string{"bash"}, Text: "bash"}},

		// implicit add
		{"plain text is an implicit add", []string{"Buy", "groceries"}, ParsedArgs{Command: "add", Args: []string{"Buy", "groceries"}, Text: "Buy groceries"}},
		{"empty args", []string{}, ParsedArgs{}},

		// typo/"did you mean" guard (MDD-037): only fires for a single bare
		// word, never for multi-word task text like "complete falls through
		// to add" above.
		{"typo: lis suggests list", []string{"lis"}, ParsedArgs{Command: "typo", Args: []string{"lis"}, Text: "lis", Suggestion: "list"}},
		{"typo: lsit (transposition) suggests list", []string{"lsit"}, ParsedArgs{Command: "typo", Args: []string{"lsit"}, Text: "lsit", Suggestion: "list"}},
		{"typo: toggel suggests toggle", []string{"toggel"}, ParsedArgs{Command: "typo", Args: []string{"toggel"}, Text: "toggel", Suggestion: "toggle"}},
		{"typo: achive suggests archive", []string{"achive"}, ParsedArgs{Command: "typo", Args: []string{"achive"}, Text: "achive", Suggestion: "archive"}},
		{"typo: compelte suggests complete", []string{"compelte"}, ParsedArgs{Command: "typo", Args: []string{"compelte"}, Text: "compelte", Suggestion: "complete"}},
		{"typo: wrong case exact match suggests complete", []string{"Complete"}, ParsedArgs{Command: "typo", Args: []string{"Complete"}, Text: "Complete", Suggestion: "complete"}},
		{"typo: bare edit (missing id) suggests edit", []string{"edit"}, ParsedArgs{Command: "typo", Args: []string{"edit"}, Text: "edit", Suggestion: "edit"}},
		{"typo guard ignores unrelated single word", []string{"Groceries"}, ParsedArgs{Command: "add", Args: []string{"Groceries"}, Text: "Groceries"}},
		{"typo guard ignores unrelated single word (2)", []string{"Standup"}, ParsedArgs{Command: "add", Args: []string{"Standup"}, Text: "Standup"}},
		{"typo guard excludes short commands from fuzzy match", []string{"Tap"}, ParsedArgs{Command: "add", Args: []string{"Tap"}, Text: "Tap"}},

		// flags
		{"-y sets Yes", []string{"clear", "-y"}, ParsedArgs{Command: "clear", Args: []string{}, Text: "", Yes: true}},
		{"--yes sets Yes", []string{"clear", "--yes"}, ParsedArgs{Command: "clear", Args: []string{}, Text: "", Yes: true}},
		{"--recursive sets Recursive", []string{"list", "--recursive"}, ParsedArgs{Command: "list", Args: []string{}, Text: "", Recursive: true}},
		{"-a sets StatusFilter active", []string{"list", "-a"}, ParsedArgs{Command: "list", Args: []string{}, Text: "", StatusFilter: "active"}},
		{"--active sets StatusFilter active", []string{"list", "--active"}, ParsedArgs{Command: "list", Args: []string{}, Text: "", StatusFilter: "active"}},
		{"--done sets StatusFilter done", []string{"list", "--done"}, ParsedArgs{Command: "list", Args: []string{}, Text: "", StatusFilter: "done"}},
		{"-a and --done together is unknown", []string{"list", "-a", "--done"}, ParsedArgs{Command: "unknown", Args: []string{"--done"}, Text: "--done"}},
		{"-p captures following value", []string{"add", "text", "-p", "/tmp/proj"}, ParsedArgs{Command: "add", Args: []string{"text"}, Text: "text", Path: "/tmp/proj"}},
		{"--path captures following value", []string{"add", "text", "--path", "/tmp/proj"}, ParsedArgs{Command: "add", Args: []string{"text"}, Text: "text", Path: "/tmp/proj"}},
		{"trailing -p with no value is unknown", []string{"add", "text", "-p"}, ParsedArgs{Command: "unknown", Args: []string{"-p"}, Text: "-p"}},

		// -h/-v short-circuit everything else
		{"-h short-circuits", []string{"list", "-h", "extra"}, ParsedArgs{Command: "help"}},
		{"--help short-circuits", []string{"--help"}, ParsedArgs{Command: "help"}},
		{"-v short-circuits", []string{"add", "text", "-v"}, ParsedArgs{Command: "version"}},
		{"--version short-circuits", []string{"--version"}, ParsedArgs{Command: "version"}},

		// unrecognized dash-prefixed tokens
		{"unrecognized dash flag is unknown", []string{"-c", "something"}, ParsedArgs{Command: "unknown", Args: []string{"-c", "something"}, Text: "-c something"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseArgs(tt.args)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseArgs(%v) = %+v, want %+v", tt.args, got, tt.want)
			}
		})
	}
}
