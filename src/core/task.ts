export type TaskStatus = 'pending' | 'completed'

export interface Task {
	id: number
	text: string
	status: TaskStatus
	lineNumber: number
	section: string | null
}

export interface Section {
	name: string
	lineNumber: number
}

export interface LintIssue {
	line: number
	issue: string
	fixed: boolean
}

export interface LintResult {
	issues: LintIssue[]
	fixedCount: number
}

function createTask(
	id: number,
	text: string,
	status: TaskStatus,
	lineNumber: number,
	section: string | null = null,
): Task {
	return { id, text, status, lineNumber, section }
}

export function parseHeaderLine(line: string): string | null {
	const match = line.match(/^##\s+(.+)$/)
	return match ? match[1].trim() : null
}

export function formatTask(task: Task): string {
	const checkboxMap: Record<TaskStatus, string> = {
		pending: '[ ]',
		completed: '[x]',
	}
	return `- ${checkboxMap[task.status]} ${task.text}`
}

export function parseTaskLine(
	line: string,
	lineNumber: number,
	id: number,
	section: string | null = null,
): Task | null {
	const match = line.match(/^(\s*)-\s*\[([ xX/])\]\s*(.*)$/)
	if (!match) return null

	const marker = match[2].toLowerCase()
	let status: TaskStatus = 'pending'
	if (marker === 'x') {
		status = 'completed'
	}
	// Note: '/' is now treated as 'pending' since in-progress status is removed
	const text = match[3].trim()

	return createTask(id, text, status, lineNumber, section)
}

export interface ParsedTaskInput {
	text: string
	sectionTag: string | null
}

const SECTION_ALIASES: Record<string, string> = {
	ff: 'Features',
	bb: 'Bugs',
	ii: 'Ideas',
	ww: 'Warnings',
}

export function parseTaskInput(input: string): ParsedTaskInput {
	// Match @Word at the end of the input (single word only)
	// Requires @ to be preceded by whitespace or be at start of string
	const match = input.match(/(?:^|\s)@(\w+)\s*$/)
	if (!match) {
		return { text: input.trim(), sectionTag: null }
	}

	const rawTag = match[1]
	const sectionTag = SECTION_ALIASES[rawTag.toLowerCase()] || rawTag
	const text = input.slice(0, match.index).trim()

	return { text, sectionTag }
}
