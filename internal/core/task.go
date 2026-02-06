package core

import (
	"regexp"
	"strings"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskPending   TaskStatus = "pending"
	TaskCompleted TaskStatus = "completed"
)

// Task represents a single task item
type Task struct {
	ID         int
	Text       string
	Status     TaskStatus
	LineNumber int
	Section    *string // nil means no section
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

var (
	headerRegex  = regexp.MustCompile(`^##\s+(.+)$`)
	taskRegex    = regexp.MustCompile(`^(\s*)-\s*\[([ xX/])\]\s*(.*)$`)
	sectionRegex = regexp.MustCompile(`(?:^|\s)@(\w+)\s*$`)
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
	if task.Status == TaskCompleted {
		checkbox = "[x]"
	}
	return "- " + checkbox + " " + task.Text
}

// ParseTaskLine parses a markdown task line
func ParseTaskLine(line string, lineNumber int, id int, section *string) *Task {
	match := taskRegex.FindStringSubmatch(line)
	if match == nil {
		return nil
	}

	marker := strings.ToLower(match[2])
	status := TaskPending
	if marker == "x" {
		status = TaskCompleted
	}
	// Note: '/' is treated as 'pending' since in-progress status is removed

	text := strings.TrimSpace(match[3])

	return &Task{
		ID:         id,
		Text:       text,
		Status:     status,
		LineNumber: lineNumber,
		Section:    section,
	}
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
