import { existsSync } from 'node:fs'
import { readFile, writeFile } from 'node:fs/promises'
import {
	type LintIssue,
	type LintResult,
	type Section,
	type Task,
	type TaskStatus,
	formatTask,
	parseHeaderLine,
	parseTaskInput,
	parseTaskLine,
} from './task.js'

export class TodoFile {
	private lines: string[] = []
	private tasks: Task[] = []
	private sections: Section[] = []

	constructor(
		public readonly filePath: string,
		content?: string,
	) {
		if (content !== undefined) {
			this.parse(content)
		}
	}

	static async load(filePath: string): Promise<TodoFile> {
		if (!existsSync(filePath)) {
			return new TodoFile(filePath, '')
		}
		const content = await readFile(filePath, 'utf-8')
		return new TodoFile(filePath, content)
	}

	static async create(filePath: string): Promise<TodoFile> {
		const todoFile = new TodoFile(filePath, '# TODO\n\n')
		await todoFile.save()
		return todoFile
	}

	private parse(content: string): void {
		this.lines = content.split('\n')
		this.tasks = []
		this.sections = []

		let taskId = 1
		let currentSection: string | null = null

		for (let i = 0; i < this.lines.length; i++) {
			const line = this.lines[i]

			// Check for section header (## Header)
			const sectionName = parseHeaderLine(line)
			if (sectionName) {
				currentSection = sectionName
				this.sections.push({ name: sectionName, lineNumber: i })
				continue
			}

			// Check for task
			const task = parseTaskLine(line, i, taskId, currentSection)
			if (task) {
				this.tasks.push(task)
				taskId++
			}
		}
	}

	getTasks(): Task[] {
		return [...this.tasks]
	}

	getTask(id: number): Task | undefined {
		return this.tasks.find((t) => t.id === id)
	}

	getPendingTasks(): Task[] {
		return this.tasks.filter((t) => t.status === 'pending')
	}

	getCompletedTasks(): Task[] {
		return this.tasks.filter((t) => t.status === 'completed')
	}

	getSections(): Section[] {
		return [...this.sections]
	}

	getTasksBySection(sectionName: string | null): Task[] {
		return this.tasks.filter((t) => t.section === sectionName)
	}

	getTasksGroupedBySection(): Map<string | null, Task[]> {
		const groups = new Map<string | null, Task[]>()

		// Ensure we have entries for all sections in order
		groups.set(null, []) // Tasks without a section
		for (const section of this.sections) {
			groups.set(section.name, [])
		}

		// Group tasks
		for (const task of this.tasks) {
			const sectionTasks = groups.get(task.section)
			if (sectionTasks) {
				sectionTasks.push(task)
			}
		}

		// Remove empty groups
		for (const [key, tasks] of groups) {
			if (tasks.length === 0) {
				groups.delete(key)
			}
		}

		return groups
	}

	addTask(text: string): Task {
		const parsed = parseTaskInput(text)
		const taskText = parsed.text
		const newTaskLine = `- [ ] ${taskText}`

		let lineNumber: number
		if (parsed.sectionTag) {
			lineNumber = this.findOrCreateSection(parsed.sectionTag)
		} else {
			lineNumber = this.findInsertPosition()
		}

		this.lines.splice(lineNumber, 0, newTaskLine)

		// Re-parse to update all line numbers
		this.parse(this.lines.join('\n'))

		const task = this.tasks.find((t) => t.text === taskText)
		if (!task) {
			throw new Error('Failed to add task')
		}
		return task
	}

	private findOrCreateSection(sectionTag: string): number {
		// Case-insensitive search for existing section
		const existingSection = this.sections.find(
			(s) => s.name.toLowerCase() === sectionTag.toLowerCase(),
		)

		if (existingSection) {
			// Find the last task in this section, or insert right after the header
			const sectionTasks = this.tasks.filter(
				(t) => t.section?.toLowerCase() === sectionTag.toLowerCase(),
			)
			if (sectionTasks.length > 0) {
				const lastTask = sectionTasks[sectionTasks.length - 1]
				return lastTask.lineNumber + 1
			}
			// No tasks in section yet, insert after the header with blank line if needed
			const insertLine = existingSection.lineNumber + 1
			// Check if there's already a blank line after the header
			if (
				insertLine < this.lines.length &&
				this.lines[insertLine].trim() !== ''
			) {
				// No blank line, add one
				this.lines.splice(insertLine, 0, '')
				this.parse(this.lines.join('\n'))
				return insertLine + 1
			}
			return insertLine + 1 // After the blank line
		}

		// Section doesn't exist, create it at the end
		const sectionHeader = `## ${sectionTag}`

		// Find insert position for new section (after all existing content)
		let insertPos = this.lines.length
		// Skip trailing empty lines
		while (insertPos > 0 && this.lines[insertPos - 1].trim() === '') {
			insertPos--
		}

		// Add blank line before section if there's content
		if (insertPos > 0) {
			this.lines.splice(insertPos, 0, '', sectionHeader, '')
			return insertPos + 3 // After blank line, header, and blank line
		}

		this.lines.splice(insertPos, 0, sectionHeader, '')
		return insertPos + 2 // After header and blank line
	}

	private findInsertPosition(): number {
		// Insert at the top (inbox behavior) - after the main header but before existing tasks
		// Find the first # header (main title like "# TODO")
		for (let i = 0; i < this.lines.length; i++) {
			const line = this.lines[i]
			// Found main header (single #, not ## section)
			if (line.match(/^#\s+/) && !line.startsWith('##')) {
				// Skip any blank lines immediately after the header
				let insertPos = i + 1
				while (
					insertPos < this.lines.length &&
					this.lines[insertPos].trim() === ''
				) {
					insertPos++
				}
				return insertPos
			}
		}

		// No header found, insert at start
		return 0
	}

	updateTask(id: number, text: string): boolean {
		const task = this.getTask(id)
		if (!task) return false

		const checkboxMap: Record<TaskStatus, string> = {
			pending: ' ',
			completed: 'x',
		}
		const updatedLine = `- [${checkboxMap[task.status]}] ${text}`
		this.lines[task.lineNumber] = updatedLine
		this.parse(this.lines.join('\n'))
		return true
	}

	toggleTask(id: number): boolean {
		const task = this.getTask(id)
		if (!task) return false

		const nextStatus: Record<TaskStatus, TaskStatus> = {
			pending: 'completed',
			completed: 'pending',
		}
		const newStatus = nextStatus[task.status]
		const updatedLine = formatTask({ ...task, status: newStatus })
		this.lines[task.lineNumber] = updatedLine
		this.parse(this.lines.join('\n'))
		return true
	}

	setTaskStatus(id: number, status: TaskStatus): boolean {
		const task = this.getTask(id)
		if (!task) return false

		const updatedLine = formatTask({ ...task, status })
		this.lines[task.lineNumber] = updatedLine
		this.parse(this.lines.join('\n'))
		return true
	}

	addNote(text: string): void {
		const noteLine = `- ${text}`
		const lineNumber = this.findOrCreateSection('Notes')
		this.lines.splice(lineNumber, 0, noteLine)
		this.parse(this.lines.join('\n'))
	}

	getSectionNames(): string[] {
		return this.sections.map((s) => s.name)
	}

	moveTask(taskId: number, targetSection: string | null): boolean {
		const task = this.getTask(taskId)
		if (!task) return false

		// Remove from current location
		this.lines.splice(task.lineNumber, 1)
		this.parse(this.lines.join('\n'))

		// Find new position
		let insertLine: number
		if (targetSection === null) {
			insertLine = this.findInsertPosition()
		} else {
			insertLine = this.findOrCreateSection(targetSection)
		}

		// Insert at new location
		this.lines.splice(insertLine, 0, formatTask(task))
		this.parse(this.lines.join('\n'))
		return true
	}

	deleteTask(id: number): boolean {
		const task = this.getTask(id)
		if (!task) return false

		this.lines.splice(task.lineNumber, 1)
		this.parse(this.lines.join('\n'))
		return true
	}

	deleteCompletedTasks(): number {
		const completed = this.getCompletedTasks()
		if (completed.length === 0) return 0

		// Delete from bottom to top to preserve line numbers
		const sortedByLine = [...completed].sort(
			(a, b) => b.lineNumber - a.lineNumber,
		)
		for (const task of sortedByLine) {
			this.lines.splice(task.lineNumber, 1)
		}

		this.parse(this.lines.join('\n'))
		return completed.length
	}

	findTasks(keyword: string): Task[] {
		const lower = keyword.toLowerCase()
		return this.tasks.filter((t) => t.text.toLowerCase().includes(lower))
	}

	lint(): LintResult {
		const issues: LintIssue[] = []
		let fixedCount = 0

		// Remove empty sections (sections with only blank lines)
		// Need to check this before other fixes since we'll be removing lines
		for (let i = 0; i < this.lines.length; i++) {
			const line = this.lines[i]

			// Check if this is a section header (## Header)
			if (line.match(/^##\s+/)) {
				// Look ahead to see if there is any content before the next section or end of file
				let hasContent = false
				let j = i + 1

				while (j < this.lines.length) {
					const nextLine = this.lines[j]

					// If we hit another section header, stop looking
					if (nextLine.match(/^##\s+/)) {
						break
					}

					// If we find any non-empty line, this section has content
					if (nextLine.trim() !== '') {
						hasContent = true
						break
					}

					j++
				}

				// If no content found (only blank lines), remove this section header
				if (!hasContent) {
					const sectionName = line.replace(/^##\s+/, '').trim()
					this.lines.splice(i, 1)
					issues.push({
						line: i + 1,
						issue: `Empty section "${sectionName}" removed`,
						fixed: true,
					})
					fixedCount++
					// Adjust index since we removed a line
					i--
				}
			}
		}

		// Fix heading spacing - ensure one blank line above and below ## headers
		for (let i = 0; i < this.lines.length; i++) {
			const line = this.lines[i]

			// Check if this is a section header (## Header)
			if (line.match(/^##\s+/)) {
				let needsFix = false

				// Check line above (should be blank, unless it's the first line or after main header)
				if (i > 0) {
					const lineAbove = this.lines[i - 1]
					// Allow no blank line if previous line is main header (# TODO)
					const isAfterMainHeader =
						lineAbove.match(/^#\s+/) && !lineAbove.startsWith('##')

					if (!isAfterMainHeader && lineAbove.trim() !== '') {
						// Need blank line above
						this.lines.splice(i, 0, '')
						issues.push({
							line: i + 1,
							issue: 'Missing blank line before heading',
							fixed: true,
						})
						fixedCount++
						needsFix = true
						i++ // Adjust index since we inserted a line
					}
				}

				// Check line below (should be blank, unless it's the last line)
				if (i + 1 < this.lines.length) {
					const lineBelow = this.lines[i + 1]
					if (lineBelow.trim() !== '') {
						// Need blank line below
						this.lines.splice(i + 1, 0, '')
						issues.push({
							line: i + 2,
							issue: 'Missing blank line after heading',
							fixed: true,
						})
						fixedCount++
						needsFix = true
					}
				}
			}
		}

		// Fix consecutive blank lines (more than one in a row)
		let consecutiveBlankStart: number | null = null
		for (let i = 0; i < this.lines.length; i++) {
			if (this.lines[i].trim() === '') {
				if (consecutiveBlankStart === null) {
					consecutiveBlankStart = i
				}
			} else {
				if (consecutiveBlankStart !== null) {
					const blankCount = i - consecutiveBlankStart
					if (blankCount > 1) {
						// Remove extra blank lines, keep only one
						const extraLines = blankCount - 1
						this.lines.splice(consecutiveBlankStart + 1, extraLines)
						issues.push({
							line: consecutiveBlankStart + 2,
							issue: `${extraLines} extra blank line${extraLines > 1 ? 's' : ''}`,
							fixed: true,
						})
						fixedCount++
						// Adjust index since we removed lines
						i -= extraLines
					}
				}
				consecutiveBlankStart = null
			}
		}
		// Handle trailing blank lines
		if (consecutiveBlankStart !== null) {
			const blankCount = this.lines.length - consecutiveBlankStart
			if (blankCount > 1) {
				const extraLines = blankCount - 1
				this.lines.splice(consecutiveBlankStart + 1, extraLines)
				issues.push({
					line: consecutiveBlankStart + 2,
					issue: `${extraLines} extra trailing blank line${extraLines > 1 ? 's' : ''}`,
					fixed: true,
				})
				fixedCount++
			}
		}

		for (let i = 0; i < this.lines.length; i++) {
			const line = this.lines[i]
			const original = line

			// Skip empty lines and headers
			if (line.trim() === '' || line.startsWith('#')) continue

			// Check for malformed task lines that look like they should be tasks
			// Pattern: starts with - and has some form of checkbox
			const taskPattern = /^(\s*)-\s*\[([^\]]*)\]\s*(.*)$/
			const match = line.match(taskPattern)

			if (match) {
				const indent = match[1]
				let checkbox = match[2]
				const text = match[3]

				let fixed = false

				// Fix: empty checkbox [] → [ ]
				if (checkbox === '') {
					checkbox = ' '
					fixed = true
					issues.push({
						line: i + 1,
						issue: 'Empty checkbox',
						fixed: true,
					})
				}

				// Fix: uppercase X → lowercase x
				if (checkbox === 'X') {
					checkbox = 'x'
					fixed = true
					issues.push({
						line: i + 1,
						issue: 'Uppercase X in checkbox',
						fixed: true,
					})
				}

				// Fix: multiple spaces or other characters in checkbox
				if (checkbox !== ' ' && checkbox !== 'x') {
					// Convert '/' to ' ' (in-progress status removed)
					if (checkbox === '/') {
						checkbox = ' '
						fixed = true
						issues.push({
							line: i + 1,
							issue: 'In-progress status [/] converted to pending [ ]',
							fixed: true,
						})
					} else if (
						checkbox.trim() === '' ||
						checkbox.trim().toLowerCase() === 'x'
					) {
						checkbox = checkbox.trim() === '' ? ' ' : 'x'
						fixed = true
						issues.push({
							line: i + 1,
							issue: 'Malformed checkbox content',
							fixed: true,
						})
					}
				}

				if (fixed) {
					this.lines[i] = `${indent}- [${checkbox}] ${text.trim()}`
					fixedCount++
				}
			}
		}

		// Re-parse if changes were made
		if (fixedCount > 0) {
			this.parse(this.lines.join('\n'))
		}

		return { issues, fixedCount }
	}

	serialize(): string {
		return this.lines.join('\n')
	}

	private reorderTasks(): void {
		if (this.tasks.length <= 1) return

		// Group tasks by section
		const tasksBySection = this.getTasksGroupedBySection()

		// Helper to sort tasks: pending → completed
		const sortByStatus = (tasks: Task[]): Task[] => {
			const pending = tasks.filter((t) => t.status === 'pending')
			const completed = tasks.filter((t) => t.status === 'completed')
			return [...pending, ...completed]
		}

		// For each section, sort tasks: pending first, then completed
		const sortedTasks: Task[] = []

		// Process tasks without section first
		const noSection = tasksBySection.get(null) || []
		sortedTasks.push(...sortByStatus(noSection))

		// Process each section in order
		for (const section of this.sections) {
			const sectionTasks = tasksBySection.get(section.name) || []
			sortedTasks.push(...sortByStatus(sectionTasks))
		}

		// Get the line numbers where tasks currently are
		const taskLineNumbers = this.tasks
			.map((t) => t.lineNumber)
			.sort((a, b) => a - b)

		// Replace task lines with sorted tasks
		for (let i = 0; i < sortedTasks.length; i++) {
			const lineNum = taskLineNumbers[i]
			this.lines[lineNum] = formatTask(sortedTasks[i])
		}

		// Re-parse to update task references
		this.parse(this.lines.join('\n'))
	}

	async save(): Promise<void> {
		this.reorderTasks()
		let content = this.serialize()
		// Ensure file ends with newline
		if (!content.endsWith('\n')) {
			content += '\n'
		}
		await writeFile(this.filePath, content, 'utf-8')
	}
}
