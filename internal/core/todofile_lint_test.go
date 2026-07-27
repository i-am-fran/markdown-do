package core

import (
	"strings"
	"testing"
)

func issueTexts(result LintResult) []string {
	texts := make([]string, len(result.Issues))
	for i, issue := range result.Issues {
		texts[i] = issue.Issue
	}
	return texts
}

func containsIssue(result LintResult, substr string) bool {
	for _, issue := range result.Issues {
		if strings.Contains(issue.Issue, substr) {
			return true
		}
	}
	return false
}

func TestLintRemovesEmptySection(t *testing.T) {
	tf := NewTodoFile("TODO.md", "# TODO\n\n## Empty\n\n## Bugs\n\n- [ ] Fix bug\n")

	result := tf.Lint()

	if !containsIssue(result, `Empty section "Empty" removed`) {
		t.Errorf("expected an \"Empty section removed\" issue, got %v", issueTexts(result))
	}
	for _, name := range tf.GetSectionNames() {
		if name == "Empty" {
			t.Errorf("expected \"Empty\" section removed, sections are %v", tf.GetSectionNames())
		}
	}
}

func TestLintAddsMissingBlankLineBeforeHeading(t *testing.T) {
	tf := NewTodoFile("TODO.md", "# TODO\n- [ ] Task\n## Bugs\n\n- [ ] Bug1\n")

	result := tf.Lint()

	if !containsIssue(result, "Missing blank line before heading") {
		t.Errorf("expected a \"Missing blank line before heading\" issue, got %v", issueTexts(result))
	}
	if !strings.Contains(tf.Serialize(), "- [ ] Task\n\n## Bugs") {
		t.Errorf("expected a blank line inserted before \"## Bugs\", got:\n%s", tf.Serialize())
	}
}

func TestLintAddsMissingBlankLineAfterHeading(t *testing.T) {
	tf := NewTodoFile("TODO.md", "# TODO\n\n## Bugs\n- [ ] Bug1\n")

	result := tf.Lint()

	if !containsIssue(result, "Missing blank line after heading") {
		t.Errorf("expected a \"Missing blank line after heading\" issue, got %v", issueTexts(result))
	}
	if !strings.Contains(tf.Serialize(), "## Bugs\n\n- [ ] Bug1") {
		t.Errorf("expected a blank line inserted after \"## Bugs\", got:\n%s", tf.Serialize())
	}
}

func TestLintCollapsesExtraBlankLinesAroundHeading(t *testing.T) {
	tf := NewTodoFile("TODO.md", "# TODO\n\n\n\n## Bugs\n\n- [ ] Bug1\n")

	result := tf.Lint()

	if !containsIssue(result, "2 extra blank lines before heading") {
		t.Errorf("expected a \"2 extra blank lines before heading\" issue, got %v", issueTexts(result))
	}
	if strings.Contains(tf.Serialize(), "\n\n\n") {
		t.Errorf("expected extra blank lines collapsed, got:\n%s", tf.Serialize())
	}
}

func TestLintCollapsesConsecutiveBlankLines(t *testing.T) {
	tf := NewTodoFile("TODO.md", "# TODO\n\n- [ ] A\n\n\n\n- [ ] B\n")

	result := tf.Lint()

	if !containsIssue(result, "extra blank line") {
		t.Errorf("expected an \"extra blank line\" issue, got %v", issueTexts(result))
	}
	if strings.Contains(tf.Serialize(), "\n\n\n") {
		t.Errorf("expected consecutive blank lines collapsed, got:\n%s", tf.Serialize())
	}
}

func TestLintTrimsTrailingBlankLines(t *testing.T) {
	tf := NewTodoFile("TODO.md", "# TODO\n\n- [ ] A\n\n\n\n")

	result := tf.Lint()

	if !containsIssue(result, "extra trailing blank line") {
		t.Errorf("expected an \"extra trailing blank line\" issue, got %v", issueTexts(result))
	}
}

// TestLintPluralizesTenOrMoreExtraBlankLines pins the fix for pluralize()'s
// old `string(rune('0'+count))` implementation, which produced a garbage
// character (e.g. ':' for 10) instead of the digits "10"/"11"/etc. for any
// count >= 10.
func TestLintPluralizesTenOrMoreExtraBlankLines(t *testing.T) {
	content := "# TODO\n\n- [ ] A\n" + strings.Repeat("\n", 12) + "- [ ] B\n"
	tf := NewTodoFile("TODO.md", content)

	result := tf.Lint()

	if !containsIssue(result, "11 extra blank lines") {
		t.Errorf("expected exactly \"11 extra blank lines\", got %v", issueTexts(result))
	}
}

func TestLintFixesEmptyCheckbox(t *testing.T) {
	tf := NewTodoFile("TODO.md", "# TODO\n\n- [] Do it\n")

	result := tf.Lint()

	if !containsIssue(result, "Empty checkbox") {
		t.Errorf("expected an \"Empty checkbox\" issue, got %v", issueTexts(result))
	}
	tasks := tf.GetTasks()
	if len(tasks) != 1 || tasks[0].Status != TaskPending || tasks[0].Text != "Do it" {
		t.Errorf("expected 1 pending task %q, got %+v", "Do it", tasks)
	}
}

func TestLintFixesUppercaseXCheckbox(t *testing.T) {
	tf := NewTodoFile("TODO.md", "# TODO\n\n- [X] Done thing\n")

	result := tf.Lint()

	if !containsIssue(result, "Uppercase X in checkbox") {
		t.Errorf("expected an \"Uppercase X in checkbox\" issue, got %v", issueTexts(result))
	}
	tasks := tf.GetTasks()
	if len(tasks) != 1 || tasks[0].Status != TaskCompleted {
		t.Errorf("expected 1 completed task, got %+v", tasks)
	}
}

func TestLintConvertsInProgressCheckboxToPending(t *testing.T) {
	tf := NewTodoFile("TODO.md", "# TODO\n\n- [/] Doing thing\n")

	result := tf.Lint()

	if !containsIssue(result, "In-progress status [/] converted to pending [ ]") {
		t.Errorf("expected an in-progress conversion issue, got %v", issueTexts(result))
	}
	tasks := tf.GetTasks()
	if len(tasks) != 1 || tasks[0].Status != TaskPending {
		t.Errorf("expected 1 pending task, got %+v", tasks)
	}
}

func TestLintRemovesEmptyTask(t *testing.T) {
	tf := NewTodoFile("TODO.md", "# TODO\n\n- [ ] \n- [ ] Real task\n")

	result := tf.Lint()

	if !containsIssue(result, "Empty task removed") {
		t.Errorf("expected an \"Empty task removed\" issue, got %v", issueTexts(result))
	}
	tasks := tf.GetTasks()
	if len(tasks) != 1 || tasks[0].Text != "Real task" {
		t.Errorf("expected only \"Real task\" to remain, got %+v", tasks)
	}
}

func TestLintConvertsTaskInNotesSectionToBullet(t *testing.T) {
	tf := NewTodoFile("TODO.md", "# TODO\n\n## Notes\n\n- [ ] Should be a note\n")

	result := tf.Lint()

	if !containsIssue(result, "Task in Notes section converted to list item") {
		t.Errorf("expected a Notes-conversion issue, got %v", issueTexts(result))
	}
	if len(tf.GetTasks()) != 0 {
		t.Errorf("expected no tasks left (converted to a plain bullet), got %+v", tf.GetTasks())
	}
	if !strings.Contains(tf.Serialize(), "- Should be a note") {
		t.Errorf("expected a plain bullet line, got:\n%s", tf.Serialize())
	}
}

func TestLintNoIssuesReturnsZeroFixedCount(t *testing.T) {
	tf := NewTodoFile("TODO.md", "# TODO\n\n- [ ] Task A\n\n## Bugs\n\n- [x] Task B\n")

	result := tf.Lint()

	if result.FixedCount != 0 {
		t.Errorf("expected FixedCount 0 for an already-clean file, got %d (issues: %v)", result.FixedCount, issueTexts(result))
	}
	if len(result.Issues) != 0 {
		t.Errorf("expected no issues for an already-clean file, got %v", issueTexts(result))
	}
}
