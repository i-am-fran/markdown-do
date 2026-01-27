import { spawn } from 'node:child_process'
import pc from 'picocolors'
import { getSettings } from '../config/settings.js'
import {
	findDefaultTodoFile,
	findTodoFiles,
	findTodoFilesInSubdirs,
} from '../core/finder.js'
import type { Task } from '../core/task.js'
import { TodoFile } from '../core/todo.js'

// Version is set at build time
const VERSION = '1.0.0'

export interface ParsedArgs {
	text: string | null
	value: string | null
	flags: Set<string>
}

export function parseArgs(args: string[]): ParsedArgs {
	const flags = new Set<string>()
	const positional: string[] = []
	let value: string | null = null

	for (let i = 0; i < args.length; i++) {
		const arg = args[i]

		if (arg.startsWith('-')) {
			const flag = arg.replace(/^-+/, '')
			flags.add(flag)

			// Check if next arg is a value for this flag
			if (i + 1 < args.length && !args[i + 1].startsWith('-')) {
				if (
					flag === 'c' ||
					flag === 'd' ||
					flag === 'e' ||
					flag === 'f' ||
					flag === 'fs' ||
					flag === 'n' ||
					flag === 't'
				) {
					value = args[i + 1]
					i++
				}
			}
		} else {
			positional.push(arg)
		}
	}

	// Join all positional args as the task text
	const text = positional.length > 0 ? positional.join(' ') : null

	return { text, value, flags }
}

function formatTaskLine(
	task: Task,
	showFile?: string,
	displayId?: number,
): string {
	let checkbox: string
	let text: string
	switch (task.status) {
		case 'completed':
			checkbox = pc.green('[x]')
			text = pc.strikethrough(pc.dim(task.text))
			break
		default:
			checkbox = pc.dim('[ ]')
			text = task.text
	}
	const id = pc.dim(`${displayId ?? task.id}.`)
	const file = showFile ? pc.dim(` (${showFile})`) : ''

	return `  ${id} ${checkbox} ${text}${file}`
}

export async function listTasks(recursive = false): Promise<void> {
	const cwd = process.cwd()

	if (recursive) {
		const filesByDir = await findTodoFilesInSubdirs(cwd)

		if (filesByDir.size === 0) {
			console.log(pc.dim('No TODO files found'))
			return
		}

		let globalTaskId = 1

		for (const [dir, files] of filesByDir) {
			console.log(pc.bold(pc.cyan(dir)))

			for (const file of files) {
				const todoFile = await TodoFile.load(file.path)
				const tasks = todoFile.getTasks()

				if (tasks.length === 0) {
					console.log(pc.dim('  (no tasks)'))
				} else {
					for (const task of tasks) {
						console.log(formatTaskLine(task, undefined, globalTaskId))
						globalTaskId++
					}
				}
			}
			console.log()
		}
	} else {
		const filePath = await findDefaultTodoFile(cwd)
		const todoFile = await TodoFile.load(filePath)
		const tasksBySection = todoFile.getTasksGroupedBySection()

		if (tasksBySection.size === 0) {
			console.log(pc.dim('No tasks found'))
			return
		}

		for (const [section, tasks] of tasksBySection) {
			if (section === null) {
				console.log(pc.bold(pc.cyan('Tasks:')))
			} else {
				console.log()
				console.log(pc.bold(pc.magenta(`## ${section}`)))
			}

			for (const task of tasks) {
				console.log(formatTaskLine(task))
			}
		}
	}
}

export async function addTask(text: string): Promise<void> {
	const filePath = await findDefaultTodoFile()
	let todoFile = await TodoFile.load(filePath)

	// Create file if it doesn't exist
	if (todoFile.getTasks().length === 0 && todoFile.serialize().trim() === '') {
		todoFile = await TodoFile.create(filePath)
	}

	const task = todoFile.addTask(text)
	await todoFile.save()

	console.log(pc.green('Added:'), formatTaskLine(task))
}

export async function deleteTask(idStr: string): Promise<void> {
	const id = Number.parseInt(idStr, 10)
	if (Number.isNaN(id)) {
		console.error(pc.red('Invalid task ID'))
		process.exit(1)
	}

	const filePath = await findDefaultTodoFile()
	const todoFile = await TodoFile.load(filePath)
	const task = todoFile.getTask(id)

	if (!task) {
		console.error(pc.red(`Task ${id} not found`))
		process.exit(1)
	}

	todoFile.deleteTask(id)
	await todoFile.save()

	console.log(pc.yellow('Deleted:'), task.text)
}

export async function editTask(idStr: string, newText: string): Promise<void> {
	const id = Number.parseInt(idStr, 10)
	if (Number.isNaN(id)) {
		console.error(pc.red('Invalid task ID'))
		process.exit(1)
	}

	const filePath = await findDefaultTodoFile()
	const todoFile = await TodoFile.load(filePath)

	const task = todoFile.getTask(id)
	if (!task) {
		console.error(pc.red(`Task ${id} not found`))
		process.exit(1)
	}

	const taskStatus = task.status

	if (!todoFile.updateTask(id, newText)) {
		console.error(pc.red(`Task ${id} not found`))
		process.exit(1)
	}

	await todoFile.save()
	console.log(
		pc.green('Updated:'),
		formatTaskLine({
			id,
			text: newText,
			status: taskStatus,
			lineNumber: 0,
			section: null,
		}),
	)
}

export async function toggleTask(idStr: string): Promise<void> {
	const id = Number.parseInt(idStr, 10)
	if (Number.isNaN(id)) {
		console.error(pc.red('Invalid task ID'))
		process.exit(1)
	}

	const filePath = await findDefaultTodoFile()
	const todoFile = await TodoFile.load(filePath)

	const task = todoFile.getTask(id)
	if (!task) {
		console.error(pc.red(`Task ${id} not found`))
		process.exit(1)
	}

	const taskText = task.text
	const prevStatus = task.status

	if (!todoFile.toggleTask(id)) {
		console.error(pc.red(`Task ${id} not found`))
		process.exit(1)
	}

	await todoFile.save()

	const newStatus = prevStatus === 'pending' ? 'completed' : 'pending'
	const actionMap = {
		pending: 'Reopened',
		completed: 'Completed',
	}
	console.log(
		pc.green(`${actionMap[newStatus]}:`),
		formatTaskLine({
			id,
			text: taskText,
			status: newStatus,
			lineNumber: 0,
			section: null,
		}),
	)
}

export async function completeTask(idStr: string): Promise<void> {
	const id = Number.parseInt(idStr, 10)
	if (Number.isNaN(id)) {
		console.error(pc.red('Invalid task ID'))
		process.exit(1)
	}

	const filePath = await findDefaultTodoFile()
	const todoFile = await TodoFile.load(filePath)

	const task = todoFile.getTask(id)
	if (!task) {
		console.error(pc.red(`Task ${id} not found`))
		process.exit(1)
	}

	const taskText = task.text

	if (!todoFile.setTaskStatus(id, 'completed')) {
		console.error(pc.red(`Task ${id} not found`))
		process.exit(1)
	}

	await todoFile.save()
	console.log(
		pc.green('Completed:'),
		formatTaskLine({
			id,
			text: taskText,
			status: 'completed',
			lineNumber: 0,
			section: null,
		}),
	)
}

export async function addNote(text: string): Promise<void> {
	const filePath = await findDefaultTodoFile()
	let todoFile = await TodoFile.load(filePath)

	// Create file if it doesn't exist
	if (todoFile.getTasks().length === 0 && todoFile.serialize().trim() === '') {
		todoFile = await TodoFile.create(filePath)
	}

	todoFile.addNote(text)
	await todoFile.save()

	console.log(pc.green('Added note:'), text)
}

export async function findTasksByKeyword(
	keyword: string,
	recursive = false,
): Promise<void> {
	const cwd = process.cwd()
	const files = await findTodoFiles(cwd, recursive)

	if (files.length === 0) {
		console.log(pc.dim('No TODO files found'))
		return
	}

	let found = false
	console.log(pc.bold(pc.cyan(`Searching for: "${keyword}"`)))
	console.log()

	for (const file of files) {
		const todoFile = await TodoFile.load(file.path)
		const matches = todoFile.findTasks(keyword)

		if (matches.length > 0) {
			found = true
			if (recursive) {
				console.log(pc.dim(file.relativePath))
			}
			for (const task of matches) {
				console.log(formatTaskLine(task))
			}
		}
	}

	if (!found) {
		console.log(pc.dim('No matching tasks found'))
	}
}

export async function openInEditor(): Promise<void> {
	const filePath = await findDefaultTodoFile()
	const settings = getSettings()

	let command: string
	let args: string[]

	switch (settings.editor) {
		case 'vim':
			command = 'vim'
			args = [filePath]
			break
		case 'nano':
			command = 'nano'
			args = [filePath]
			break
		case 'default-app':
			command = 'open'
			args = [filePath]
			break
		default:
			command = process.env.EDITOR || process.env.VISUAL || 'vim'
			args = [filePath]
			break
	}

	console.log(pc.dim(`Opening ${filePath}...`))

	const child = spawn(command, args, {
		stdio: 'inherit',
	})

	child.on('error', (err) => {
		console.error(pc.red(`Failed to open editor: ${err.message}`))
		process.exit(1)
	})
}

export async function deleteCompletedTasks(): Promise<void> {
	const filePath = await findDefaultTodoFile()
	const todoFile = await TodoFile.load(filePath)

	const count = todoFile.deleteCompletedTasks()

	if (count === 0) {
		console.log(pc.dim('No completed tasks to delete'))
		return
	}

	await todoFile.save()
	console.log(
		pc.green(`Deleted ${count} completed task${count === 1 ? '' : 's'}`),
	)
}

export async function lintFile(): Promise<void> {
	const filePath = await findDefaultTodoFile()
	const todoFile = await TodoFile.load(filePath)

	console.log(pc.dim(`Linting ${filePath}...`))
	console.log()

	const tasks = todoFile.getTasks()
	const pendingCount = tasks.filter((t) => t.status === 'pending').length
	const completedCount = tasks.filter((t) => t.status === 'completed').length

	const result = todoFile.lint()

	if (result.issues.length === 0) {
		console.log(pc.green('✓ No issues found'))
		console.log()
		console.log(
			pc.dim(
				`Checked ${tasks.length} task${tasks.length === 1 ? '' : 's'} (${pendingCount} pending, ${completedCount} completed)`,
			),
		)
		return
	}

	console.log(pc.bold(pc.cyan('Lint results:')))
	console.log()

	for (const issue of result.issues) {
		const status = issue.fixed ? pc.green('fixed') : pc.yellow('found')
		console.log(`  Line ${issue.line}: ${issue.issue} [${status}]`)
	}

	console.log()

	if (result.fixedCount > 0) {
		await todoFile.save()
		console.log(
			pc.green(
				`✓ Fixed ${result.fixedCount} issue${result.fixedCount === 1 ? '' : 's'}`,
			),
		)
	}

	console.log(
		pc.dim(
			`Checked ${tasks.length} task${tasks.length === 1 ? '' : 's'} (${pendingCount} pending, ${completedCount} completed)`,
		),
	)
}

export function showVersion(): void {
	console.log(`markdown-do v${VERSION}`)
}

export function showHelp(): void {
	console.log(`
${pc.bold(pc.cyan('mdd'))} - MarkdownDO: Manage TODO.md files

${pc.bold('Usage:')}
  mdd                    Open interactive TUI
  mdd <task text>        Add a new task (quotes optional)
  mdd -l                 List tasks
  mdd -ls                List tasks recursively
  mdd -t <id>            Toggle task status
  mdd -c <id>            Complete task by ID
  mdd -e <id> <text>     Edit task text
  mdd -d <id>            Delete task by ID
  mdd -dc                Delete all completed tasks
  mdd -f <keyword>       Find tasks by keyword
  mdd -fs <keyword>      Find tasks recursively
  mdd -n <text>          Add a note to ## Notes section
  mdd -o                 Open TODO file in editor
  mdd -lint              Lint and fix TODO file formatting
  mdd -v, --version      Show version
  mdd -h, --help         Show this help

${pc.bold('Examples:')}
  mdd Buy groceries      Add new task (no quotes needed)
  mdd -l                 Show all tasks
  mdd -t 2               Toggle task #2
  mdd -c 1               Complete task #1
  mdd -e 1 Fix the bug   Edit task #1
  mdd -d 3               Delete task #3
  mdd -f bug             Find tasks containing "bug"
`)
}
