package core

import (
	"errors"
	"regexp"
	"strings"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskPending    TaskStatus = "pending"
	TaskInProgress TaskStatus = "in-progress"
	TaskCompleted  TaskStatus = "completed"
)

// Task represents a single task item
type Task struct {
	ID         int
	StableID   string // persistent, user-visible ID tag, e.g. "ABC01"; empty unless assigned via -id
	Text       string
	Status     TaskStatus
	LineNumber int
	Section    *string  // nil means no section
	Notes      []string // trimmed text of trailing indented bullet lines
}

// BlockLineCount returns how many physical lines this task occupies,
// including its own checkbox line and all trailing note lines.
func (t *Task) BlockLineCount() int {
	return 1 + len(t.Notes)
}

// Section represents a section header in the TODO file
type Section struct {
	Name       string
	LineNumber int
}

// LintIssue represents a single lint issue
type LintIssue struct {
	Line  int
	Issue string
	Fixed bool
}

// LintResult contains the results of linting a file
type LintResult struct {
	Issues     []LintIssue
	FixedCount int
}

// ParsedTaskInput contains parsed task text and optional section tag
type ParsedTaskInput struct {
	Text       string
	SectionTag *string
}

// Section aliases for quick input
var sectionAliases = map[string]string{
	"ff": "Features",
	"bb": "Bugs",
	"ii": "Ideas",
	"ww": "Warnings",
}

// SetSectionAliases adds or overrides `@alias` shortcuts on top of the
// built-in ones (ff/bb/ii/ww). Intended to be called once at startup with the
// user's configured aliases, before any task input is parsed.
func SetSectionAliases(overrides map[string]string) {
	for k, v := range overrides {
		sectionAliases[strings.ToLower(k)] = v
	}
}

var (
	headerRegex   = regexp.MustCompile(`^##\s+(.+)$`)
	taskRegex     = regexp.MustCompile(`^(\s*)-\s*\[([ xX/])\]\s*(.*)$`)
	sectionRegex  = regexp.MustCompile(`(?:^|\s)@(\w+)\s*$`)
	stableIDRegex = regexp.MustCompile(`^\[([A-Z]{3}\d+)\]\s*`)
	idPrefixRegex = regexp.MustCompile(`^[A-Za-z]{3}$`)
	noteRegex     = regexp.MustCompile(`^\s+-\s+(\S.*)$`)
)

// ParseHeaderLine extracts section name from a header line
func ParseHeaderLine(line string) *string {
	match := headerRegex.FindStringSubmatch(line)
	if match == nil {
		return nil
	}
	name := strings.TrimSpace(match[1])
	return &name
}

// FormatTask formats a task as a markdown checkbox line
func FormatTask(task *Task) string {
	checkbox := "[ ]"
	switch task.Status {
	case TaskCompleted:
		checkbox = "[x]"
	case TaskInProgress:
		checkbox = "[/]"
	}
	text := task.Text
	if task.StableID != "" {
		text = "[" + task.StableID + "] " + text
	}
	return "- " + checkbox + " " + text
}

// FormatTaskBlock formats a task's checkbox line followed by its note lines.
func FormatTaskBlock(task *Task) []string {
	block := []string{FormatTask(task)}
	for _, note := range task.Notes {
		block = append(block, "  - "+note)
	}
	return block
}

// ParseTaskLine parses a markdown task line
func ParseTaskLine(line string, lineNumber int, id int, section *string) *Task {
	match := taskRegex.FindStringSubmatch(line)
	if match == nil {
		return nil
	}

	marker := strings.ToLower(match[2])
	status := TaskPending
	switch marker {
	case "x":
		status = TaskCompleted
	case "/":
		status = TaskInProgress
	}

	rawText := match[3]
	stableID := ""
	if idMatch := stableIDRegex.FindStringSubmatch(rawText); idMatch != nil {
		stableID = idMatch[1]
		rawText = stableIDRegex.ReplaceAllString(rawText, "")
	}
	text := strings.TrimSpace(rawText)

	return &Task{
		ID:         id,
		StableID:   stableID,
		Text:       text,
		Status:     status,
		LineNumber: lineNumber,
		Section:    section,
	}
}

// ValidateIDPrefix normalizes and validates a 3-letter ID prefix (e.g. "abc" -> "ABC")
func ValidateIDPrefix(prefix string) (string, error) {
	if !idPrefixRegex.MatchString(prefix) {
		return "", errors.New("id prefix must be exactly 3 letters (e.g. \"ABC\")")
	}
	return strings.ToUpper(prefix), nil
}

// ParseTaskInput parses user input to extract task text and optional section tag
func ParseTaskInput(input string) ParsedTaskInput {
	match := sectionRegex.FindStringSubmatchIndex(input)
	if match == nil {
		return ParsedTaskInput{
			Text:       strings.TrimSpace(input),
			SectionTag: nil,
		}
	}

	// Extract the tag (group 1)
	rawTag := input[match[2]:match[3]]

	// Check for alias
	sectionTag := rawTag
	if alias, ok := sectionAliases[strings.ToLower(rawTag)]; ok {
		sectionTag = alias
	}

	text := strings.TrimSpace(input[:match[0]])

	return ParsedTaskInput{
		Text:       text,
		SectionTag: &sectionTag,
	}
}

// CopySectionPtr creates a copy of a section pointer
func CopySectionPtr(s *string) *string {
	if s == nil {
		return nil
	}
	copy := *s
	return &copy
}
