package core

import "testing"

func TestParseTaskLineExtractsStableID(t *testing.T) {
	task := ParseTaskLine("- [ ] [ABC-001] Buy milk", 0, 1, nil)
	if task == nil {
		t.Fatal("expected task to parse, got nil")
	}
	if task.StableID != "ABC-001" {
		t.Errorf("expected StableID %q, got %q", "ABC-001", task.StableID)
	}
	if task.Text != "Buy milk" {
		t.Errorf("expected Text %q (tag stripped), got %q", "Buy milk", task.Text)
	}
}

func TestParseTaskLineExtractsOldFormatStableID(t *testing.T) {
	task := ParseTaskLine("- [ ] [ABC01] Buy milk", 0, 1, nil)
	if task == nil {
		t.Fatal("expected task to parse, got nil")
	}
	if task.StableID != "ABC01" {
		t.Errorf("expected old-format StableID %q to still parse, got %q", "ABC01", task.StableID)
	}
	if task.Text != "Buy milk" {
		t.Errorf("expected Text %q (tag stripped), got %q", "Buy milk", task.Text)
	}
}

func TestParseTaskLineWithoutStableID(t *testing.T) {
	task := ParseTaskLine("- [ ] Buy milk", 0, 1, nil)
	if task == nil {
		t.Fatal("expected task to parse, got nil")
	}
	if task.StableID != "" {
		t.Errorf("expected empty StableID, got %q", task.StableID)
	}
	if task.Text != "Buy milk" {
		t.Errorf("expected Text %q, got %q", "Buy milk", task.Text)
	}
}

func TestFormatTaskRoundTrip(t *testing.T) {
	original := "- [ ] [ABC-001] Buy milk"
	task := ParseTaskLine(original, 0, 1, nil)
	if task == nil {
		t.Fatal("expected task to parse, got nil")
	}
	if got := FormatTask(task); got != original {
		t.Errorf("round trip mismatch: got %q, want %q", got, original)
	}
}

func TestFormatTaskCompletedRoundTrip(t *testing.T) {
	original := "- [x] [ABC-001] Buy milk"
	task := ParseTaskLine(original, 0, 1, nil)
	if task == nil {
		t.Fatal("expected task to parse, got nil")
	}
	if got := FormatTask(task); got != original {
		t.Errorf("round trip mismatch: got %q, want %q", got, original)
	}
}

func TestFormatTaskInProgressRoundTrip(t *testing.T) {
	original := "- [/] [ABC-001] Buy milk"
	task := ParseTaskLine(original, 0, 1, nil)
	if task == nil {
		t.Fatal("expected task to parse, got nil")
	}
	if task.Status != TaskInProgress {
		t.Errorf("expected TaskInProgress, got %v", task.Status)
	}
	if got := FormatTask(task); got != original {
		t.Errorf("round trip mismatch: got %q, want %q", got, original)
	}
}

func TestFormatTaskBlockRoundTrip(t *testing.T) {
	task := &Task{Text: "Buy milk", Status: TaskPending, Notes: []string{"note one", "note two"}}
	block := FormatTaskBlock(task)
	want := []string{
		"- [ ] Buy milk",
		"  - note one",
		"  - note two",
	}
	if len(block) != len(want) {
		t.Fatalf("expected %d lines, got %d: %v", len(want), len(block), block)
	}
	for i := range want {
		if block[i] != want[i] {
			t.Errorf("line %d: expected %q, got %q", i, want[i], block[i])
		}
	}
}

func TestFormatTaskBlockNoNotes(t *testing.T) {
	task := &Task{Text: "Buy milk", Status: TaskPending}
	block := FormatTaskBlock(task)
	if len(block) != 1 || block[0] != "- [ ] Buy milk" {
		t.Errorf("expected single-line block, got %v", block)
	}
}

func TestBlockLineCount(t *testing.T) {
	task := &Task{Notes: []string{"a", "b"}}
	if got := task.BlockLineCount(); got != 3 {
		t.Errorf("expected BlockLineCount 3, got %d", got)
	}
	empty := &Task{}
	if got := empty.BlockLineCount(); got != 1 {
		t.Errorf("expected BlockLineCount 1 for no notes, got %d", got)
	}
}

func TestSetSectionAliasesAddsCustomAlias(t *testing.T) {
	SetSectionAliases(map[string]string{"pp": "Projects"})

	parsed := ParseTaskInput("Plan launch @pp")
	if parsed.SectionTag == nil || *parsed.SectionTag != "Projects" {
		t.Fatalf("expected custom alias to resolve to Projects, got %v", parsed.SectionTag)
	}

	// Built-in aliases still resolve alongside the custom one.
	parsed = ParseTaskInput("Fix bug @bb")
	if parsed.SectionTag == nil || *parsed.SectionTag != "Bugs" {
		t.Fatalf("expected built-in alias to still resolve to Bugs, got %v", parsed.SectionTag)
	}
}

func TestParseTaskInputUnescapedAtSignStillTagsSection(t *testing.T) {
	parsed := ParseTaskInput("Fix bug @bb")
	if parsed.SectionTag == nil || *parsed.SectionTag != "Bugs" {
		t.Fatalf("expected unescaped @bb to still resolve to Bugs, got %v", parsed.SectionTag)
	}
	if parsed.Text != "Fix bug" {
		t.Errorf("expected text %q, got %q", "Fix bug", parsed.Text)
	}
}

func TestParseTaskInputEscapedAtSignIsLiteral(t *testing.T) {
	parsed := ParseTaskInput(`Meet Bob \@bb`)
	if parsed.SectionTag != nil {
		t.Fatalf("expected no section tag for escaped @, got %v", *parsed.SectionTag)
	}
	if parsed.Text != "Meet Bob @bb" {
		t.Errorf("expected literal text %q, got %q", "Meet Bob @bb", parsed.Text)
	}
}

func TestParseTaskInputEscapedBraces(t *testing.T) {
	parsed := ParseTaskInput(`Do \{a thing\}`)
	if parsed.SectionTag != nil {
		t.Fatalf("expected no section tag, got %v", *parsed.SectionTag)
	}
	if parsed.Text != "Do {a thing}" {
		t.Errorf("expected literal text %q, got %q", "Do {a thing}", parsed.Text)
	}
}

func TestParseTaskInputDoubleBackslashIsLiteralBackslash(t *testing.T) {
	parsed := ParseTaskInput(`Path C:\\Users`)
	if parsed.Text != `Path C:\Users` {
		t.Errorf("expected literal text %q, got %q", `Path C:\Users`, parsed.Text)
	}
}

func TestValidateIDPrefix(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"abc", "ABC", false},
		{"ABC", "ABC", false},
		{"AbC", "ABC", false},
		{"ab", "", true},
		{"abcd", "", true},
		{"a1c", "", true},
		{"", "", true},
	}

	for _, c := range cases {
		got, err := ValidateIDPrefix(c.in)
		if c.wantErr {
			if err == nil {
				t.Errorf("ValidateIDPrefix(%q): expected error, got nil", c.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("ValidateIDPrefix(%q): unexpected error: %v", c.in, err)
		}
		if got != c.want {
			t.Errorf("ValidateIDPrefix(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
