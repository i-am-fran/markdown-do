package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// idTagRegex matches a stable ID tag's prefix and number. The dash is
// optional so IDs tagged before MDD29 (e.g. "ABC01") still parse.
var idTagRegex = regexp.MustCompile(`^([A-Z]{3})-?(\d+)$`)

// TodoFile manages a TODO.md file
type TodoFile struct {
	FilePath string
	lines    []string
	tasks    []Task
	sections []Section
}

// NewTodoFile creates a new TodoFile instance
func NewTodoFile(filePath string, content string) *TodoFile {
	tf := &TodoFile{
		FilePath: filePath,
	}
	tf.parse(content)
	return tf
}

// Load loads a TodoFile from disk
func Load(filePath string) (*TodoFile, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return NewTodoFile(filePath, ""), nil
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return NewTodoFile(filePath, string(content)), nil
}

// Create creates a new TODO.md file with default content
func Create(filePath string) (*TodoFile, error) {
	tf := NewTodoFile(filePath, "# TODO\n\n")
	if err := tf.Save(); err != nil {
		return nil, err
	}
	return tf, nil
}

func (tf *TodoFile) parse(content string) {
	tf.lines = strings.Split(content, "\n")
	tf.tasks = nil
	tf.sections = nil

	taskID := 1
	var currentSection *string

	for i := 0; i < len(tf.lines); i++ {
		line := tf.lines[i]

		// Check for section header (## Header)
		if sectionName := ParseHeaderLine(line); sectionName != nil {
			currentSection = sectionName
			tf.sections = append(tf.sections, Section{
				Name:       *sectionName,
				LineNumber: i,
			})
			continue
		}

		// Check for task
		if task := ParseTaskLine(line, i, taskID, CopySectionPtr(currentSection)); task != nil {
			notes, consumed := collectTaskNotes(tf.lines, i+1)
			task.Notes = notes
			tf.tasks = append(tf.tasks, *task)
			taskID++
			i += consumed
		}
	}
}

// collectTaskNotes scans forward from start, returning the trimmed text of
// each contiguous indented plain-bullet line and how many lines were
// consumed. A note run ends at a blank line, a section header, a checkbox
// line (which always wins over note detection), a non-indented line, or EOF.
func collectTaskNotes(lines []string, start int) (notes []string, consumed int) {
	for i := start; i < len(lines); i++ {
		line := lines[i]
		if strings.TrimSpace(line) == "" {
			break
		}
		if ParseHeaderLine(line) != nil {
			break
		}
		if ParseTaskLine(line, 0, 0, nil) != nil {
			break
		}
		m := noteRegex.FindStringSubmatch(line)
		if m == nil {
			break
		}
		notes = append(notes, strings.TrimSpace(m[1]))
		consumed++
	}
	return notes, consumed
}

// GetTasks returns a copy of all tasks
func (tf *TodoFile) GetTasks() []Task {
	result := make([]Task, len(tf.tasks))
	copy(result, tf.tasks)
	return result
}

// GetTask returns a task by ID
func (tf *TodoFile) GetTask(id int) *Task {
	for i := range tf.tasks {
		if tf.tasks[i].ID == id {
			return &tf.tasks[i]
		}
	}
	return nil
}

// GetTaskByStableID returns a task by its persistent StableID (case-insensitive)
func (tf *TodoFile) GetTaskByStableID(id string) *Task {
	for i := range tf.tasks {
		if strings.EqualFold(tf.tasks[i].StableID, id) {
			return &tf.tasks[i]
		}
	}
	return nil
}

// SetTaskIDs assigns sequential IDs (e.g. ABC-001, ABC-002, ...) to every
// task in the file, overwriting any IDs already present. Numbering starts at
// the bottom of the file (the oldest task, since new tasks are inserted at
// the top) and increases moving upward.
func (tf *TodoFile) SetTaskIDs(prefix string) error {
	normalizedPrefix, err := ValidateIDPrefix(prefix)
	if err != nil {
		return err
	}

	total := len(tf.tasks)
	for i, task := range tf.tasks {
		task.StableID = normalizedPrefix + "-" + idSuffix(total-i)
		tf.lines[task.LineNumber] = FormatTask(&task)
	}

	tf.parse(strings.Join(tf.lines, "\n"))
	return nil
}

// RemoveTaskIDs strips the StableID tag from every task in the file.
// Returns true if any task had an ID removed.
func (tf *TodoFile) RemoveTaskIDs() bool {
	changed := false
	for _, task := range tf.tasks {
		if task.StableID == "" {
			continue
		}
		task.StableID = ""
		tf.lines[task.LineNumber] = FormatTask(&task)
		changed = true
	}

	if changed {
		tf.parse(strings.Join(tf.lines, "\n"))
	}
	return changed
}

// activeIDSequence returns the prefix and next available number of the ID
// scheme currently in use in the file, if any task has a StableID.
func (tf *TodoFile) activeIDSequence() (prefix string, nextNum int, ok bool) {
	maxNum := 0
	for _, t := range tf.tasks {
		match := idTagRegex.FindStringSubmatch(t.StableID)
		if match == nil {
			continue
		}
		prefix = match[1]
		ok = true
		if n, err := strconv.Atoi(match[2]); err == nil && n > maxNum {
			maxNum = n
		}
	}
	return prefix, maxNum + 1, ok
}

func idSuffix(n int) string {
	return fmt.Sprintf("%03d", n)
}

// GetPendingTasks returns all pending tasks
func (tf *TodoFile) GetPendingTasks() []Task {
	var result []Task
	for _, t := range tf.tasks {
		if t.Status == TaskPending {
			result = append(result, t)
		}
	}
	return result
}

// GetCompletedTasks returns all completed tasks
func (tf *TodoFile) GetCompletedTasks() []Task {
	var result []Task
	for _, t := range tf.tasks {
		if t.Status == TaskCompleted {
			result = append(result, t)
		}
	}
	return result
}

// GetSections returns a copy of all sections
func (tf *TodoFile) GetSections() []Section {
	result := make([]Section, len(tf.sections))
	copy(result, tf.sections)
	return result
}

// GetTasksBySection returns tasks in a specific section
func (tf *TodoFile) GetTasksBySection(sectionName *string) []Task {
	var result []Task
	for _, t := range tf.tasks {
		if (sectionName == nil && t.Section == nil) ||
			(sectionName != nil && t.Section != nil && *sectionName == *t.Section) {
			result = append(result, t)
		}
	}
	return result
}

// GetTasksGroupedBySection returns tasks grouped by section
func (tf *TodoFile) GetTasksGroupedBySection() map[*string][]Task {
	groups := make(map[*string][]Task)

	// nil key for tasks without section
	var nilKey *string
	groups[nilKey] = nil

	// Create entries for all sections in order
	for i := range tf.sections {
		groups[&tf.sections[i].Name] = nil
	}

	// Group tasks
	for _, task := range tf.tasks {
		if task.Section == nil {
			groups[nilKey] = append(groups[nilKey], task)
		} else {
			// Find the matching section key
			for i := range tf.sections {
				if tf.sections[i].Name == *task.Section {
					groups[&tf.sections[i].Name] = append(groups[&tf.sections[i].Name], task)
					break
				}
			}
		}
	}

	// Remove empty groups
	for key, tasks := range groups {
		if len(tasks) == 0 {
			delete(groups, key)
		}
	}

	return groups
}

// GetTasksGroupedBySectionOrdered returns tasks grouped by section in order
func (tf *TodoFile) GetTasksGroupedBySectionOrdered() []struct {
	Section *string
	Tasks   []Task
} {
	var result []struct {
		Section *string
		Tasks   []Task
	}

	// First, tasks without section
	noSectionTasks := tf.GetTasksBySection(nil)
	if len(noSectionTasks) > 0 {
		result = append(result, struct {
			Section *string
			Tasks   []Task
		}{nil, noSectionTasks})
	}

	// Then, sections in order
	for _, section := range tf.sections {
		sectionName := section.Name
		tasks := tf.GetTasksBySection(&sectionName)
		if len(tasks) > 0 {
			result = append(result, struct {
				Section *string
				Tasks   []Task
			}{&sectionName, tasks})
		}
	}

	return result
}

// AddTask adds a new task
func (tf *TodoFile) AddTask(text string) (*Task, error) {
	parsed := ParseTaskInput(text)
	taskText := parsed.Text

	if len(strings.TrimSpace(taskText)) < 3 {
		return nil, errors.New("task must be at least 3 characters")
	}

	newTaskLine := "- [ ] " + taskText
	if prefix, nextNum, ok := tf.activeIDSequence(); ok {
		newTaskLine = "- [ ] [" + prefix + "-" + idSuffix(nextNum) + "] " + taskText
	}

	var lineNumber int
	if parsed.SectionTag != nil {
		lineNumber = tf.findOrCreateSection(*parsed.SectionTag)
	} else {
		lineNumber = tf.findInsertPosition()
	}

	// Insert the new line
	tf.lines = append(tf.lines[:lineNumber], append([]string{newTaskLine}, tf.lines[lineNumber:]...)...)

	// Re-parse to update all line numbers
	tf.parse(strings.Join(tf.lines, "\n"))

	// Find the newly added task
	for i := range tf.tasks {
		if tf.tasks[i].Text == taskText {
			return &tf.tasks[i], nil
		}
	}

	return nil, errors.New("failed to add task")
}

func (tf *TodoFile) findOrCreateSection(sectionTag string) int {
	// Case-insensitive search for existing section
	var existingSection *Section
	for i := range tf.sections {
		if strings.EqualFold(tf.sections[i].Name, sectionTag) {
			existingSection = &tf.sections[i]
			break
		}
	}

	if existingSection != nil {
		// Find the last task in this section, or insert right after the header
		var lastTask *Task
		for i := range tf.tasks {
			if tf.tasks[i].Section != nil && strings.EqualFold(*tf.tasks[i].Section, sectionTag) {
				lastTask = &tf.tasks[i]
			}
		}

		if lastTask != nil {
			return lastTask.LineNumber + 1
		}

		// No tasks in section yet, insert after the header with blank line if needed
		insertLine := existingSection.LineNumber + 1
		if insertLine < len(tf.lines) && strings.TrimSpace(tf.lines[insertLine]) != "" {
			// No blank line, add one
			tf.lines = append(tf.lines[:insertLine], append([]string{""}, tf.lines[insertLine:]...)...)
			tf.parse(strings.Join(tf.lines, "\n"))
			return insertLine + 1
		}
		return insertLine + 1 // After the blank line
	}

	// Section doesn't exist, create it
	sectionHeader := "## " + sectionTag

	// Special handling for Notes section - it should always be last
	isNotesSection := strings.EqualFold(sectionTag, "Notes")

	var insertPos int
	if isNotesSection {
		// Notes goes at the very end
		insertPos = len(tf.lines)
		// Skip trailing empty lines
		for insertPos > 0 && strings.TrimSpace(tf.lines[insertPos-1]) == "" {
			insertPos--
		}
	} else {
		// For non-Notes sections, insert before Notes if it exists, otherwise at end
		notesLineNum := -1
		for i := range tf.sections {
			if strings.EqualFold(tf.sections[i].Name, "Notes") {
				notesLineNum = tf.sections[i].LineNumber
				break
			}
		}

		if notesLineNum >= 0 {
			// Insert before Notes section (before the blank line preceding it)
			insertPos = notesLineNum
			// Look for blank line before Notes
			if insertPos > 0 && strings.TrimSpace(tf.lines[insertPos-1]) == "" {
				insertPos--
			}
		} else {
			// No Notes section, insert at end
			insertPos = len(tf.lines)
			// Skip trailing empty lines
			for insertPos > 0 && strings.TrimSpace(tf.lines[insertPos-1]) == "" {
				insertPos--
			}
		}
	}

	// Add blank line before section if there's content
	if insertPos > 0 {
		newLines := []string{"", sectionHeader, ""}
		tf.lines = append(tf.lines[:insertPos], append(newLines, tf.lines[insertPos:]...)...)
		return insertPos + 3 // After blank line, header, and blank line
	}

	newLines := []string{sectionHeader, ""}
	tf.lines = append(tf.lines[:insertPos], append(newLines, tf.lines[insertPos:]...)...)
	return insertPos + 2 // After header and blank line
}

func (tf *TodoFile) findInsertPosition() int {
	// Insert at the top (inbox behavior) - after the main header but before existing tasks
	// Find the first # header (main title like "# TODO")
	mainHeaderRegex := regexp.MustCompile(`^#\s+`)

	for i, line := range tf.lines {
		// Found main header (single #, not ## section)
		if mainHeaderRegex.MatchString(line) && !strings.HasPrefix(line, "##") {
			// Skip any blank lines immediately after the header
			insertPos := i + 1
			for insertPos < len(tf.lines) && strings.TrimSpace(tf.lines[insertPos]) == "" {
				insertPos++
			}
			return insertPos
		}
	}

	// No header found, insert at start
	return 0
}

// UpdateTask updates the text of a task
func (tf *TodoFile) UpdateTask(id int, text string) bool {
	task := tf.GetTask(id)
	if task == nil {
		return false
	}

	updatedLine := FormatTask(&Task{
		ID:         task.ID,
		StableID:   task.StableID,
		Text:       text,
		Status:     task.Status,
		LineNumber: task.LineNumber,
		Section:    task.Section,
	})
	tf.lines[task.LineNumber] = updatedLine
	tf.parse(strings.Join(tf.lines, "\n"))
	return true
}

// ToggleTask toggles the status of a task. When enableInProgress is true,
// status cycles pending -> in-progress -> completed -> pending; otherwise it's
// a plain pending<->completed flip (in-progress tasks toggle to completed).
func (tf *TodoFile) ToggleTask(id int, enableInProgress bool) bool {
	task := tf.GetTask(id)
	if task == nil {
		return false
	}

	var newStatus TaskStatus
	if enableInProgress {
		switch task.Status {
		case TaskPending:
			newStatus = TaskInProgress
		case TaskInProgress:
			newStatus = TaskCompleted
		default:
			newStatus = TaskPending
		}
	} else {
		newStatus = TaskCompleted
		if task.Status == TaskCompleted {
			newStatus = TaskPending
		}
	}

	updatedLine := FormatTask(&Task{
		ID:         task.ID,
		StableID:   task.StableID,
		Text:       task.Text,
		Status:     newStatus,
		LineNumber: task.LineNumber,
		Section:    task.Section,
	})
	tf.lines[task.LineNumber] = updatedLine
	tf.parse(strings.Join(tf.lines, "\n"))
	return true
}

// SetTaskStatus sets the status of a task
func (tf *TodoFile) SetTaskStatus(id int, status TaskStatus) bool {
	task := tf.GetTask(id)
	if task == nil {
		return false
	}

	updatedLine := FormatTask(&Task{
		ID:         task.ID,
		StableID:   task.StableID,
		Text:       task.Text,
		Status:     status,
		LineNumber: task.LineNumber,
		Section:    task.Section,
	})
	tf.lines[task.LineNumber] = updatedLine
	tf.parse(strings.Join(tf.lines, "\n"))
	return true
}

// AddNote adds a note to the Notes section
func (tf *TodoFile) AddNote(text string) {
	noteLine := "- " + text
	lineNumber := tf.findOrCreateSection("Notes")
	tf.lines = append(tf.lines[:lineNumber], append([]string{noteLine}, tf.lines[lineNumber:]...)...)
	tf.parse(strings.Join(tf.lines, "\n"))
}

// GetSectionNames returns the names of all sections
func (tf *TodoFile) GetSectionNames() []string {
	names := make([]string, len(tf.sections))
	for i, s := range tf.sections {
		names[i] = s.Name
	}
	return names
}

// formatNoteLines formats a task's notes as indented plain-bullet lines.
func formatNoteLines(notes []string) []string {
	lines := make([]string, len(notes))
	for i, note := range notes {
		lines[i] = "  - " + note
	}
	return lines
}

// MoveTask moves a task to a different section
func (tf *TodoFile) MoveTask(taskID int, targetSection *string) bool {
	task := tf.GetTask(taskID)
	if task == nil {
		return false
	}

	// Remove block (task line + notes) from current location
	end := task.LineNumber + task.BlockLineCount()
	tf.lines = append(tf.lines[:task.LineNumber], tf.lines[end:]...)
	tf.parse(strings.Join(tf.lines, "\n"))

	// Find new position
	var insertLine int
	if targetSection == nil {
		insertLine = tf.findInsertPosition()
	} else {
		insertLine = tf.findOrCreateSection(*targetSection)
	}

	// Build the block to insert
	// If moving to Notes section, convert the task line to a plain list item
	// (remove checkbox) but keep its note bullets underneath it.
	var block []string
	if targetSection != nil && strings.EqualFold(*targetSection, "Notes") {
		block = append([]string{"- " + task.Text}, formatNoteLines(task.Notes)...)
	} else {
		block = FormatTaskBlock(task)
	}
	tf.lines = append(tf.lines[:insertLine], append(block, tf.lines[insertLine:]...)...)
	tf.parse(strings.Join(tf.lines, "\n"))
	return true
}

// DeleteTask deletes a task (and its notes) by ID
func (tf *TodoFile) DeleteTask(id int) bool {
	task := tf.GetTask(id)
	if task == nil {
		return false
	}

	end := task.LineNumber + task.BlockLineCount()
	tf.lines = append(tf.lines[:task.LineNumber], tf.lines[end:]...)
	tf.parse(strings.Join(tf.lines, "\n"))
	return true
}

// AddTaskNote appends a note to the given task, after any existing notes.
func (tf *TodoFile) AddTaskNote(id int, text string) bool {
	task := tf.GetTask(id)
	if task == nil {
		return false
	}

	insertAt := task.LineNumber + task.BlockLineCount()
	noteLine := "  - " + strings.TrimSpace(text)
	tf.lines = append(tf.lines[:insertAt], append([]string{noteLine}, tf.lines[insertAt:]...)...)
	tf.parse(strings.Join(tf.lines, "\n"))
	return true
}

// DeleteCompletedTasks deletes all completed tasks
func (tf *TodoFile) DeleteCompletedTasks() int {
	completed := tf.GetCompletedTasks()
	if len(completed) == 0 {
		return 0
	}

	// Delete from bottom to top to preserve line numbers
	sort.Slice(completed, func(i, j int) bool {
		return completed[i].LineNumber > completed[j].LineNumber
	})

	for _, task := range completed {
		end := task.LineNumber + task.BlockLineCount()
		tf.lines = append(tf.lines[:task.LineNumber], tf.lines[end:]...)
	}

	tf.parse(strings.Join(tf.lines, "\n"))
	return len(completed)
}

// ArchiveCompletedTasks moves all completed tasks (that aren't already in the
// Archive section) into the Archive section, appending "(from <Section>)" to
// each task's text to record where it came from. Tasks with no section keep
// their text unchanged. Returns the number of tasks archived.
func (tf *TodoFile) ArchiveCompletedTasks() int {
	var toArchive []Task
	for _, t := range tf.GetCompletedTasks() {
		if t.Section != nil && strings.EqualFold(*t.Section, "Archive") {
			continue
		}
		toArchive = append(toArchive, t)
	}
	if len(toArchive) == 0 {
		return 0
	}

	// Build the archived blocks (in original file order) before mutating lines.
	var block []string
	for _, task := range toArchive {
		archived := task
		if task.Section != nil {
			archived.Text = fmt.Sprintf("%s (from %s)", task.Text, *task.Section)
		}
		block = append(block, FormatTaskBlock(&archived)...)
	}

	// Remove the tasks from bottom to top so earlier LineNumbers stay valid.
	sort.Slice(toArchive, func(i, j int) bool {
		return toArchive[i].LineNumber > toArchive[j].LineNumber
	})
	for _, task := range toArchive {
		end := task.LineNumber + task.BlockLineCount()
		tf.lines = append(tf.lines[:task.LineNumber], tf.lines[end:]...)
	}
	tf.parse(strings.Join(tf.lines, "\n"))

	insertLine := tf.findOrCreateSection("Archive")
	tf.lines = append(tf.lines[:insertLine], append(block, tf.lines[insertLine:]...)...)
	tf.parse(strings.Join(tf.lines, "\n"))
	return len(toArchive)
}

// FindTasks finds tasks matching a keyword
func (tf *TodoFile) FindTasks(keyword string) []Task {
	lower := strings.ToLower(keyword)
	var result []Task
	for _, t := range tf.tasks {
		if strings.Contains(strings.ToLower(t.Text), lower) {
			result = append(result, t)
		}
	}
	return result
}

// Lint checks and fixes formatting issues
func (tf *TodoFile) Lint() LintResult {
	var issues []LintIssue
	fixedCount := 0

	// Remove empty sections (sections with only blank lines)
	for i := 0; i < len(tf.lines); i++ {
		line := tf.lines[i]

		// Check if this is a section header (## Header)
		sectionRegex := regexp.MustCompile(`^##\s+`)
		if sectionRegex.MatchString(line) {
			// Look ahead to see if there is any content before the next section or end of file
			hasContent := false
			j := i + 1

			for j < len(tf.lines) {
				nextLine := tf.lines[j]

				// If we hit another section header, stop looking
				if sectionRegex.MatchString(nextLine) {
					break
				}

				// If we find any non-empty line, this section has content
				if strings.TrimSpace(nextLine) != "" {
					hasContent = true
					break
				}

				j++
			}

			// If no content found (only blank lines), remove this section header
			if !hasContent {
				sectionName := strings.TrimSpace(strings.TrimPrefix(line, "##"))
				tf.lines = append(tf.lines[:i], tf.lines[i+1:]...)
				issues = append(issues, LintIssue{
					Line:  i + 1,
					Issue: "Empty section \"" + sectionName + "\" removed",
					Fixed: true,
				})
				fixedCount++
				i-- // Adjust index since we removed a line
			}
		}
	}

	// Fix heading spacing - ensure exactly one blank line above and below ## headers
	mainHeaderRegex := regexp.MustCompile(`^#\s+`)
	sectionHeaderRegex := regexp.MustCompile(`^##\s+`)

	for i := 0; i < len(tf.lines); i++ {
		line := tf.lines[i]

		// Check if this is a section header (## Header)
		if sectionHeaderRegex.MatchString(line) {
			// Check line(s) above - ensure exactly one blank line
			if i > 0 {
				lineAbove := tf.lines[i-1]
				// Allow no blank line if previous line is main header (# TODO)
				isAfterMainHeader := mainHeaderRegex.MatchString(lineAbove) && !strings.HasPrefix(lineAbove, "##")

				if !isAfterMainHeader {
					// Count blank lines above
					blankLinesAbove := 0
					checkIdx := i - 1
					for checkIdx >= 0 && strings.TrimSpace(tf.lines[checkIdx]) == "" {
						blankLinesAbove++
						checkIdx--
					}

					if blankLinesAbove == 0 {
						// No blank line above - add one
						tf.lines = append(tf.lines[:i], append([]string{""}, tf.lines[i:]...)...)
						issues = append(issues, LintIssue{
							Line:  i + 1,
							Issue: "Missing blank line before heading",
							Fixed: true,
						})
						fixedCount++
						i++ // Adjust index since we inserted a line
					} else if blankLinesAbove > 1 {
						// Multiple blank lines above - remove extras, keep only one
						extraLines := blankLinesAbove - 1
						// Remove the extra blank lines
						tf.lines = append(tf.lines[:i-blankLinesAbove], tf.lines[i-1:]...)
						issues = append(issues, LintIssue{
							Line:  i - blankLinesAbove + 1,
							Issue: pluralize(extraLines, "extra blank line") + " before heading",
							Fixed: true,
						})
						fixedCount++
						i -= extraLines // Adjust index since we removed lines
					}
				}
			}

			// Check line(s) below - ensure exactly one blank line
			if i+1 < len(tf.lines) {
				// Count blank lines below
				blankLinesBelow := 0
				checkIdx := i + 1
				for checkIdx < len(tf.lines) && strings.TrimSpace(tf.lines[checkIdx]) == "" {
					blankLinesBelow++
					checkIdx++
				}

				if blankLinesBelow == 0 {
					// No blank line below - add one
					tf.lines = append(tf.lines[:i+1], append([]string{""}, tf.lines[i+1:]...)...)
					issues = append(issues, LintIssue{
						Line:  i + 2,
						Issue: "Missing blank line after heading",
						Fixed: true,
					})
					fixedCount++
					// No need to adjust i, as next iteration will skip the blank line
				} else if blankLinesBelow > 1 {
					// Multiple blank lines below - remove extras, keep only one
					extraLines := blankLinesBelow - 1
					// Remove the extra blank lines (keep the first one)
					tf.lines = append(tf.lines[:i+2], tf.lines[i+1+blankLinesBelow:]...)
					issues = append(issues, LintIssue{
						Line:  i + 3,
						Issue: pluralize(extraLines, "extra blank line") + " after heading",
						Fixed: true,
					})
					fixedCount++
					// No need to adjust i, we removed lines after current position
				}
			}
		}
	}

	// Fix consecutive blank lines (more than one in a row)
	consecutiveBlankStart := -1
	for i := 0; i < len(tf.lines); i++ {
		if strings.TrimSpace(tf.lines[i]) == "" {
			if consecutiveBlankStart == -1 {
				consecutiveBlankStart = i
			}
		} else {
			if consecutiveBlankStart != -1 {
				blankCount := i - consecutiveBlankStart
				if blankCount > 1 {
					// Remove extra blank lines, keep only one
					extraLines := blankCount - 1
					tf.lines = append(tf.lines[:consecutiveBlankStart+1], tf.lines[consecutiveBlankStart+1+extraLines:]...)
					issues = append(issues, LintIssue{
						Line:  consecutiveBlankStart + 2,
						Issue: pluralize(extraLines, "extra blank line"),
						Fixed: true,
					})
					fixedCount++
					i -= extraLines
				}
			}
			consecutiveBlankStart = -1
		}
	}

	// Handle trailing blank lines
	if consecutiveBlankStart != -1 {
		blankCount := len(tf.lines) - consecutiveBlankStart
		if blankCount > 1 {
			extraLines := blankCount - 1
			tf.lines = tf.lines[:consecutiveBlankStart+1]
			issues = append(issues, LintIssue{
				Line:  consecutiveBlankStart + 2,
				Issue: pluralize(extraLines, "extra trailing blank line"),
				Fixed: true,
			})
			fixedCount++
		}
	}

	// Fix malformed task lines and remove empty tasks
	taskPattern := regexp.MustCompile(`^(\s*)-\s*\[([^\]]*)\]\s*(.*)$`)

	// Track current section to detect Notes
	currentSection := ""
	mainHeaderRegex2 := regexp.MustCompile(`^#\s+`)
	sectionHeaderRegex2 := regexp.MustCompile(`^##\s+(.+)$`)

	for i := 0; i < len(tf.lines); i++ {
		line := tf.lines[i]

		// Track which section we're in
		if mainHeaderRegex2.MatchString(line) && !strings.HasPrefix(line, "##") {
			currentSection = ""
		} else if match := sectionHeaderRegex2.FindStringSubmatch(line); match != nil {
			currentSection = strings.TrimSpace(match[1])
		}

		// Skip empty lines and headers
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}

		match := taskPattern.FindStringSubmatch(line)
		if match == nil {
			continue
		}

		indent := match[1]
		checkbox := match[2]
		text := strings.TrimSpace(match[3])
		fixed := false

		// Fix: Remove tasks with empty text
		if text == "" {
			tf.lines = append(tf.lines[:i], tf.lines[i+1:]...)
			issues = append(issues, LintIssue{
				Line:  i + 1,
				Issue: "Empty task removed",
				Fixed: true,
			})
			fixedCount++
			i-- // Adjust index since we removed a line
			continue
		}

		// Fix: Convert tasks in Notes section to plain list items
		if strings.EqualFold(currentSection, "Notes") {
			tf.lines[i] = indent + "- " + text
			issues = append(issues, LintIssue{
				Line:  i + 1,
				Issue: "Task in Notes section converted to list item",
				Fixed: true,
			})
			fixedCount++
			continue
		}

		// Fix: empty checkbox [] -> [ ]
		if checkbox == "" {
			checkbox = " "
			fixed = true
			issues = append(issues, LintIssue{
				Line:  i + 1,
				Issue: "Empty checkbox",
				Fixed: true,
			})
		}

		// Fix: uppercase X -> lowercase x
		if checkbox == "X" {
			checkbox = "x"
			fixed = true
			issues = append(issues, LintIssue{
				Line:  i + 1,
				Issue: "Uppercase X in checkbox",
				Fixed: true,
			})
		}

		// Fix: multiple spaces or other characters in checkbox
		if checkbox != " " && checkbox != "x" {
			// Convert '/' to ' ' (in-progress status removed)
			if checkbox == "/" {
				checkbox = " "
				fixed = true
				issues = append(issues, LintIssue{
					Line:  i + 1,
					Issue: "In-progress status [/] converted to pending [ ]",
					Fixed: true,
				})
			} else if strings.TrimSpace(checkbox) == "" || strings.ToLower(strings.TrimSpace(checkbox)) == "x" {
				if strings.TrimSpace(checkbox) == "" {
					checkbox = " "
				} else {
					checkbox = "x"
				}
				fixed = true
				issues = append(issues, LintIssue{
					Line:  i + 1,
					Issue: "Malformed checkbox content",
					Fixed: true,
				})
			}
		}

		if fixed {
			tf.lines[i] = indent + "- [" + checkbox + "] " + text
			fixedCount++
		}
	}

	// Re-parse if changes were made
	if fixedCount > 0 {
		tf.parse(strings.Join(tf.lines, "\n"))
	}

	return LintResult{
		Issues:     issues,
		FixedCount: fixedCount,
	}
}

func pluralize(count int, singular string) string {
	if count == 1 {
		return "1 " + singular
	}
	return string(rune('0'+count)) + " " + singular + "s"
}

// Serialize returns the file content as a string
func (tf *TodoFile) Serialize() string {
	return strings.Join(tf.lines, "\n")
}

// reorderTasks sorts each section's tasks pending-before-completed, keeping
// every task's notes attached to it. It rewrites each section group's line
// span as a single block-formatted chunk (a permutation of the same lines,
// so the span's length - and every other group's line numbers - is
// unaffected). If a group's tasks don't exactly fill their line span (e.g. a
// stray blank line between two tasks in the same section), that group is
// left untouched rather than risk corrupting content.
func (tf *TodoFile) reorderTasks() {
	if len(tf.tasks) <= 1 {
		return
	}

	groups := tf.GetTasksGroupedBySectionOrdered()

	sortByStatus := func(tasks []Task) []Task {
		var active, completed []Task
		for _, t := range tasks {
			if t.Status == TaskCompleted {
				completed = append(completed, t)
			} else {
				active = append(active, t)
			}
		}
		return append(active, completed...)
	}

	changed := false
	for _, group := range groups {
		if len(group.Tasks) <= 1 {
			continue
		}

		rangeStart := group.Tasks[0].LineNumber
		last := group.Tasks[len(group.Tasks)-1]
		rangeEnd := last.LineNumber + last.BlockLineCount()

		totalBlockLines := 0
		for _, t := range group.Tasks {
			totalBlockLines += t.BlockLineCount()
		}
		if totalBlockLines != rangeEnd-rangeStart {
			continue
		}

		var newLines []string
		for _, t := range sortByStatus(group.Tasks) {
			newLines = append(newLines, FormatTaskBlock(&t)...)
		}

		tf.lines = append(tf.lines[:rangeStart], append(newLines, tf.lines[rangeEnd:]...)...)
		changed = true
	}

	if changed {
		tf.parse(strings.Join(tf.lines, "\n"))
	}
}

// ensureSectionLast ensures the named section (case-insensitive) is always at
// the end of the file, moving it there if anything follows it. No-op if the
// section doesn't exist or is already last.
func (tf *TodoFile) ensureSectionLast(name string) {
	// Find the section
	lineStart := -1
	lineEnd := -1
	sectionHeaderRegex := regexp.MustCompile(`^##\s+`)

	for i, line := range tf.lines {
		if match := ParseHeaderLine(line); match != nil && strings.EqualFold(*match, name) {
			lineStart = i
			// Find where this section ends (next section or end of file)
			for j := i + 1; j < len(tf.lines); j++ {
				if sectionHeaderRegex.MatchString(tf.lines[j]) {
					lineEnd = j
					break
				}
			}
			if lineEnd == -1 {
				lineEnd = len(tf.lines)
			}
			break
		}
	}

	// If the section doesn't exist or is already at the end, nothing to do
	if lineStart == -1 {
		return
	}

	// Check if there's content after it (excluding trailing blank lines)
	hasContentAfter := false
	for i := lineEnd; i < len(tf.lines); i++ {
		if strings.TrimSpace(tf.lines[i]) != "" {
			hasContentAfter = true
			break
		}
	}

	if !hasContentAfter {
		return // Already at the end
	}

	// Extract the section (including preceding blank line if any)
	startExtract := lineStart
	if startExtract > 0 && strings.TrimSpace(tf.lines[startExtract-1]) == "" {
		startExtract--
	}

	section := make([]string, lineEnd-startExtract)
	copy(section, tf.lines[startExtract:lineEnd])

	// Remove the section from its current position
	tf.lines = append(tf.lines[:startExtract], tf.lines[lineEnd:]...)

	// Find insertion point at the end (skip trailing blank lines)
	insertPos := len(tf.lines)
	for insertPos > 0 && strings.TrimSpace(tf.lines[insertPos-1]) == "" {
		insertPos--
	}

	// Add blank line before the section if there's content
	if insertPos > 0 && strings.TrimSpace(tf.lines[insertPos-1]) != "" {
		section = append([]string{""}, section...)
	}

	// Insert the section at the end
	tf.lines = append(tf.lines[:insertPos], append(section, tf.lines[insertPos:]...)...)

	// Re-parse to update line numbers
	tf.parse(strings.Join(tf.lines, "\n"))
}

// Save saves the file to disk. Before overwriting, it backs up the file's
// current on-disk content to an undo backup file (see undoFilePath) so
// "mdd undo" can restore it. If the file doesn't exist yet on disk (the
// very first save), there's nothing to back up.
func (tf *TodoFile) Save() error {
	tf.ensureSectionLast("Archive")
	tf.ensureSectionLast("Notes")
	tf.reorderTasks()
	content := tf.Serialize()
	// Ensure file ends with newline
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	if prev, err := os.ReadFile(tf.FilePath); err == nil {
		_ = os.WriteFile(undoFilePath(tf.FilePath), prev, 0644)
	}

	return os.WriteFile(tf.FilePath, []byte(content), 0644)
}

// undoFilePath returns the path to the undo backup file for a given TODO
// file, stored under stateDir (keyed by a hash of the absolute TODO file
// path) rather than as a sibling file in the project directory.
func undoFilePath(todoFilePath string) string {
	abs := todoFilePath
	if a, err := filepath.Abs(todoFilePath); err == nil {
		abs = a
	}

	dir := filepath.Join(stateDir, "undo")
	_ = os.MkdirAll(dir, 0755)

	return filepath.Join(dir, hashPath(abs))
}

// HasUndo reports whether an undo backup exists for filePath.
func HasUndo(filePath string) bool {
	_, err := os.Stat(undoFilePath(filePath))
	return err == nil
}

// LoadUndoBackup reads the undo backup's content for filePath. It returns
// an error if no backup exists.
func LoadUndoBackup(filePath string) (string, error) {
	data, err := os.ReadFile(undoFilePath(filePath))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Snapshot creates a snapshot of the current file state
func (tf *TodoFile) Snapshot() []string {
	// Return a deep copy of lines
	snapshot := make([]string, len(tf.lines))
	copy(snapshot, tf.lines)
	return snapshot
}

// Restore restores the file from a snapshot
func (tf *TodoFile) Restore(snapshot []string) {
	// Restore lines from snapshot
	tf.lines = make([]string, len(snapshot))
	copy(tf.lines, snapshot)

	// Re-parse the content
	tf.parse(strings.Join(tf.lines, "\n"))
}
