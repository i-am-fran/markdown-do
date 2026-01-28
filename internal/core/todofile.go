package core

import (
	"errors"
	"os"
	"regexp"
	"sort"
	"strings"
)

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

	for i, line := range tf.lines {
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
			tf.tasks = append(tf.tasks, *task)
			taskID++
		}
	}
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
	newTaskLine := "- [ ] " + taskText

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

	checkbox := " "
	if task.Status == TaskCompleted {
		checkbox = "x"
	}
	updatedLine := "- [" + checkbox + "] " + text
	tf.lines[task.LineNumber] = updatedLine
	tf.parse(strings.Join(tf.lines, "\n"))
	return true
}

// ToggleTask toggles the status of a task
func (tf *TodoFile) ToggleTask(id int) bool {
	task := tf.GetTask(id)
	if task == nil {
		return false
	}

	newStatus := TaskCompleted
	if task.Status == TaskCompleted {
		newStatus = TaskPending
	}

	updatedLine := FormatTask(&Task{
		ID:         task.ID,
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

// MoveTask moves a task to a different section
func (tf *TodoFile) MoveTask(taskID int, targetSection *string) bool {
	task := tf.GetTask(taskID)
	if task == nil {
		return false
	}

	// Remove from current location
	tf.lines = append(tf.lines[:task.LineNumber], tf.lines[task.LineNumber+1:]...)
	tf.parse(strings.Join(tf.lines, "\n"))

	// Find new position
	var insertLine int
	if targetSection == nil {
		insertLine = tf.findInsertPosition()
	} else {
		insertLine = tf.findOrCreateSection(*targetSection)
	}

	// Insert at new location
	var formattedTask string
	// If moving to Notes section, convert to plain list item (remove checkbox)
	if targetSection != nil && strings.EqualFold(*targetSection, "Notes") {
		formattedTask = "- " + task.Text
	} else {
		formattedTask = FormatTask(task)
	}
	tf.lines = append(tf.lines[:insertLine], append([]string{formattedTask}, tf.lines[insertLine:]...)...)
	tf.parse(strings.Join(tf.lines, "\n"))
	return true
}

// DeleteTask deletes a task by ID
func (tf *TodoFile) DeleteTask(id int) bool {
	task := tf.GetTask(id)
	if task == nil {
		return false
	}

	tf.lines = append(tf.lines[:task.LineNumber], tf.lines[task.LineNumber+1:]...)
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
		tf.lines = append(tf.lines[:task.LineNumber], tf.lines[task.LineNumber+1:]...)
	}

	tf.parse(strings.Join(tf.lines, "\n"))
	return len(completed)
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

	// Fix heading spacing - ensure one blank line above and below ## headers
	mainHeaderRegex := regexp.MustCompile(`^#\s+`)
	sectionHeaderRegex := regexp.MustCompile(`^##\s+`)

	for i := 0; i < len(tf.lines); i++ {
		line := tf.lines[i]

		// Check if this is a section header (## Header)
		if sectionHeaderRegex.MatchString(line) {
			// Check line above (should be blank, unless it's the first line or after main header)
			if i > 0 {
				lineAbove := tf.lines[i-1]
				// Allow no blank line if previous line is main header (# TODO)
				isAfterMainHeader := mainHeaderRegex.MatchString(lineAbove) && !strings.HasPrefix(lineAbove, "##")

				if !isAfterMainHeader && strings.TrimSpace(lineAbove) != "" {
					// Need blank line above
					tf.lines = append(tf.lines[:i], append([]string{""}, tf.lines[i:]...)...)
					issues = append(issues, LintIssue{
						Line:  i + 1,
						Issue: "Missing blank line before heading",
						Fixed: true,
					})
					fixedCount++
					i++ // Adjust index since we inserted a line
				}
			}

			// Check line below (should be blank, unless it's the last line)
			if i+1 < len(tf.lines) {
				lineBelow := tf.lines[i+1]
				if strings.TrimSpace(lineBelow) != "" {
					// Need blank line below
					tf.lines = append(tf.lines[:i+1], append([]string{""}, tf.lines[i+1:]...)...)
					issues = append(issues, LintIssue{
						Line:  i + 2,
						Issue: "Missing blank line after heading",
						Fixed: true,
					})
					fixedCount++
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
		shouldRemove := false

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

func (tf *TodoFile) reorderTasks() {
	if len(tf.tasks) <= 1 {
		return
	}

	// Group tasks by section
	groups := tf.GetTasksGroupedBySectionOrdered()

	// Helper to sort tasks: pending -> completed
	sortByStatus := func(tasks []Task) []Task {
		var pending, completed []Task
		for _, t := range tasks {
			if t.Status == TaskPending {
				pending = append(pending, t)
			} else {
				completed = append(completed, t)
			}
		}
		return append(pending, completed...)
	}

	// Collect sorted tasks
	var sortedTasks []Task
	for _, group := range groups {
		sortedTasks = append(sortedTasks, sortByStatus(group.Tasks)...)
	}

	// Get the line numbers where tasks currently are
	taskLineNumbers := make([]int, len(tf.tasks))
	for i, t := range tf.tasks {
		taskLineNumbers[i] = t.LineNumber
	}
	sort.Ints(taskLineNumbers)

	// Replace task lines with sorted tasks
	for i, task := range sortedTasks {
		lineNum := taskLineNumbers[i]
		tf.lines[lineNum] = FormatTask(&task)
	}

	// Re-parse to update task references
	tf.parse(strings.Join(tf.lines, "\n"))
}

// Save saves the file to disk
func (tf *TodoFile) Save() error {
	tf.reorderTasks()
	content := tf.Serialize()
	// Ensure file ends with newline
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return os.WriteFile(tf.FilePath, []byte(content), 0644)
}
