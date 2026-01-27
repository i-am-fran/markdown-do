import * as p from '@clack/prompts'
import pc from 'picocolors'
import { openInEditor } from '../cli/commands.js'
import { getSettings } from '../config/settings.js'
import {
	findDefaultTodoFile,
	findTodoFiles,
	findTodoFilesInSubdirs,
} from '../core/finder.js'
import { TodoFile } from '../core/todo.js'
import { LIST_FOOTER, MENU_FOOTER } from './hints.js'
import { isCancel, selectWithFooter } from './select.js'
import { showSettingsMenu } from './settings.js'
import { showTaskList, showTaskMenu } from './taskList.js'

function clearScreen(): void {
	process.stdout.write('\x1B[2J\x1B[0f')
}

function enterAlternateScreen(): void {
	process.stdout.write('\x1B[?1049h')
	process.stdout.write('\x1B[2J\x1B[0f')
}

function exitAlternateScreen(): void {
	process.stdout.write('\x1B[?1049l')
}

function showTitle(): void {
	console.log(`${pc.cyan('┌')}  ${pc.bgCyan(pc.black(' MarkdownDO '))}`)
}

const DOUBLE_ESC_WINDOW_MS = 500
let lastEscTime = 0

type MainMenuAction =
	| 'list'
	| 'add'
	| 'addNote'
	| 'find'
	| 'deleteCompleted'
	| 'lint'
	| 'open'
	| 'subfolders'
	| 'settings'
	| 'quit'

export async function runTui(): Promise<void> {
	const settings = getSettings()
	let useAlternateScreen = settings.fullscreen

	if (useAlternateScreen) {
		enterAlternateScreen()
		// Handle cleanup on exit signals
		const cleanup = () => {
			exitAlternateScreen()
			process.exit(0)
		}
		process.on('SIGINT', cleanup)
		process.on('SIGTERM', cleanup)
	}

	console.log()
	showTitle()

	while (true) {
		const action = await showMainMenu()

		if (action === 'quit') {
			if (useAlternateScreen) {
				exitAlternateScreen()
			}
			p.outro(pc.dim('Goodbye!'))
			break
		}

		if (isCancel(action)) {
			const now = Date.now()
			if (now - lastEscTime < DOUBLE_ESC_WINDOW_MS) {
				if (useAlternateScreen) {
					exitAlternateScreen()
				}
				p.outro(pc.dim('Goodbye!'))
				break
			}
			lastEscTime = now
			p.log.info('Press ESC again to quit')
			continue
		}

		// Reset ESC timer on any other action
		lastEscTime = 0

		console.log()
		await handleAction(action)

		// Re-check fullscreen setting (might have changed in settings menu)
		const currentSettings = getSettings()
		if (currentSettings.fullscreen && !useAlternateScreen) {
			enterAlternateScreen()
			useAlternateScreen = true
		} else if (!currentSettings.fullscreen && useAlternateScreen) {
			exitAlternateScreen()
			useAlternateScreen = false
		}

		if (useAlternateScreen) {
			clearScreen()
			showTitle()
		}
	}
}

async function showMainMenu(): Promise<MainMenuAction | symbol> {
	const filePath = await findDefaultTodoFile()
	const todoFile = await TodoFile.load(filePath)
	const tasks = todoFile.getTasks()
	const pending = tasks.filter((t) => !t.completed).length

	const taskSummary =
		tasks.length > 0
			? pc.dim(` (${pending} pending, ${tasks.length} total)`)
			: pc.dim(' (no tasks)')

	return selectWithFooter({
		message: 'What would you like to do?',
		options: [
			{ value: 'list' as const, label: `List tasks${taskSummary}` },
			{ value: 'add' as const, label: 'Add new task' },
			{ value: 'addNote' as const, label: 'Add note' },
			{ value: 'find' as const, label: 'Find tasks' },
			{ value: 'deleteCompleted' as const, label: 'Delete completed tasks' },
			{ value: 'lint' as const, label: 'Lint and fix formatting' },
			{ value: 'open' as const, label: 'Open in editor' },
			{ value: 'subfolders' as const, label: 'List subfolders [WIP]' },
			{ value: 'settings' as const, label: 'Settings' },
			{ value: 'quit' as const, label: pc.dim('Quit') },
		],
		footer: MENU_FOOTER,
	}) as Promise<MainMenuAction | symbol>
}

async function handleAction(action: MainMenuAction): Promise<void> {
	switch (action) {
		case 'list':
			await handleListTasks()
			break
		case 'add':
			await handleAddTask()
			break
		case 'addNote':
			await handleAddNote()
			break
		case 'find':
			await handleFindTasks()
			break
		case 'deleteCompleted':
			await handleDeleteCompleted()
			break
		case 'lint':
			await handleLint()
			break
		case 'open':
			await openInEditor()
			break
		case 'subfolders':
			await handleSubfolders()
			break
		case 'settings':
			await showSettingsMenu()
			break
	}
}

async function handleListTasks(): Promise<void> {
	const filePath = await findDefaultTodoFile()
	let todoFile = await TodoFile.load(filePath)

	while (true) {
		const selectedTask = await showTaskList(todoFile)

		if (
			selectedTask === null ||
			isCancel(selectedTask) ||
			selectedTask === -1
		) {
			break
		}

		if (selectedTask === 'add') {
			await handleQuickAddTask(todoFile)
			todoFile = await TodoFile.load(filePath) // Reload after adding
			continue
		}

		await showTaskMenu(todoFile, selectedTask)
		todoFile = await TodoFile.load(filePath) // Reload after potential changes
	}
}

async function handleQuickAddTask(todoFile: TodoFile): Promise<void> {
	const text = await p.text({
		message: 'Enter task description:',
		placeholder: 'Buy groceries or "Fix bug @Features" to add to section',
	})

	if (p.isCancel(text)) return

	const trimmed = typeof text === 'string' ? text.trim() : ''
	if (!trimmed) {
		p.log.info('Empty input - cancelled')
		return
	}

	const task = todoFile.addTask(trimmed)
	await todoFile.save()

	p.log.success(`Added: ${task.text}`)
}

async function handleAddTask(): Promise<void> {
	const filePath = await findDefaultTodoFile()
	let todoFile = await TodoFile.load(filePath)

	if (todoFile.getTasks().length === 0 && todoFile.serialize().trim() === '') {
		todoFile = await TodoFile.create(filePath)
	}

	while (true) {
		const text = await p.text({
			message: 'Enter task description (empty to finish):',
			placeholder: 'Buy groceries or "Fix bug @Features" to add to section',
		})

		if (p.isCancel(text)) break

		const trimmed = typeof text === 'string' ? text.trim() : ''
		if (!trimmed) {
			p.log.info('Empty input - returning to menu')
			break
		}

		const task = todoFile.addTask(trimmed)
		await todoFile.save()

		p.log.success(`Added: ${task.text}`)
		console.log()
	}
}

async function handleSubfolders(): Promise<void> {
	const filesByDir = await findTodoFilesInSubdirs()

	if (filesByDir.size === 0) {
		p.log.warn('No TODO files found in subfolders')
		return
	}

	const options = Array.from(filesByDir.entries()).map(([dir, files]) => ({
		value: files[0].path,
		label: `${dir} (${files.length} file${files.length > 1 ? 's' : ''})`,
	}))

	const selected = await selectWithFooter({
		message: 'Select a folder:',
		options: [...options, { value: 'back', label: pc.dim('Back') }],
		footer: LIST_FOOTER,
	})

	if (isCancel(selected) || selected === 'back') return

	const filePath = selected as string
	let todoFile = await TodoFile.load(filePath)

	while (true) {
		const selectedTask = await showTaskList(todoFile)

		if (
			selectedTask === null ||
			isCancel(selectedTask) ||
			selectedTask === -1
		) {
			break
		}

		if (selectedTask === 'add') {
			await handleQuickAddTask(todoFile)
			todoFile = await TodoFile.load(filePath)
			continue
		}

		await showTaskMenu(todoFile, selectedTask)
		todoFile = await TodoFile.load(filePath)
	}
}

async function handleAddNote(): Promise<void> {
	const filePath = await findDefaultTodoFile()
	let todoFile = await TodoFile.load(filePath)

	if (todoFile.getTasks().length === 0 && todoFile.serialize().trim() === '') {
		todoFile = await TodoFile.create(filePath)
	}

	const text = await p.text({
		message: 'Enter note text:',
		placeholder: 'Remember to review the PR',
	})

	if (p.isCancel(text)) return

	const trimmed = typeof text === 'string' ? text.trim() : ''
	if (!trimmed) {
		p.log.info('Empty input - cancelled')
		return
	}

	todoFile.addNote(trimmed)
	await todoFile.save()

	p.log.success(`Added note: ${trimmed}`)
}

async function handleFindTasks(): Promise<void> {
	const keyword = await p.text({
		message: 'Enter search keyword:',
		placeholder: 'bug, feature, etc.',
	})

	if (p.isCancel(keyword)) return

	const trimmed = typeof keyword === 'string' ? keyword.trim() : ''
	if (!trimmed) {
		p.log.info('Empty input - cancelled')
		return
	}

	const recursive = await p.confirm({
		message: 'Search in subfolders?',
		initialValue: false,
	})

	if (p.isCancel(recursive)) return

	const cwd = process.cwd()
	const files = await findTodoFiles(cwd, recursive === true)

	if (files.length === 0) {
		p.log.warn('No TODO files found')
		return
	}

	let found = false
	console.log()
	console.log(pc.bold(pc.cyan(`Search results for: "${trimmed}"`)))
	console.log()

	for (const file of files) {
		const todoFile = await TodoFile.load(file.path)
		const matches = todoFile.findTasks(trimmed)

		if (matches.length > 0) {
			found = true
			if (recursive === true) {
				console.log(pc.dim(file.relativePath))
			}
			for (const task of matches) {
				const checkbox =
					task.status === 'completed'
						? pc.dim(pc.green('[x]'))
						: task.status === 'in-progress'
							? pc.dim(pc.yellow('[/]'))
							: pc.dim('[ ]')
				const text =
					task.status === 'completed'
						? pc.strikethrough(pc.dim(task.text))
						: task.status === 'in-progress'
							? pc.yellow(task.text)
							: task.text
				console.log(`  ${pc.dim(`${task.id}.`)} ${checkbox} ${text}`)
			}
		}
	}

	if (!found) {
		p.log.warn('No matching tasks found')
	}
}

async function handleDeleteCompleted(): Promise<void> {
	const filePath = await findDefaultTodoFile()
	const todoFile = await TodoFile.load(filePath)

	const completedTasks = todoFile.getCompletedTasks()

	if (completedTasks.length === 0) {
		p.log.warn('No completed tasks to delete')
		return
	}

	const shouldDelete = await p.confirm({
		message: `Delete ${completedTasks.length} completed task${completedTasks.length === 1 ? '' : 's'}?`,
	})

	if (p.isCancel(shouldDelete) || shouldDelete !== true) {
		p.log.info('Cancelled')
		return
	}

	const count = todoFile.deleteCompletedTasks()
	await todoFile.save()

	p.log.success(`Deleted ${count} completed task${count === 1 ? '' : 's'}`)
}

async function handleLint(): Promise<void> {
	const filePath = await findDefaultTodoFile()
	const todoFile = await TodoFile.load(filePath)

	console.log()
	console.log(pc.dim(`Linting ${filePath}...`))
	console.log()

	const tasks = todoFile.getTasks()
	const pendingCount = tasks.filter((t) => t.status === 'pending').length
	const inProgressCount = tasks.filter((t) => t.status === 'in-progress').length
	const completedCount = tasks.filter((t) => t.status === 'completed').length

	const result = todoFile.lint()

	if (result.issues.length === 0) {
		p.log.success('✓ No issues found')
		console.log()
		console.log(
			pc.dim(
				`Checked ${tasks.length} task${tasks.length === 1 ? '' : 's'} (${pendingCount} pending, ${inProgressCount} in progress, ${completedCount} completed)`,
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
		p.log.success(
			`✓ Fixed ${result.fixedCount} issue${result.fixedCount === 1 ? '' : 's'}`,
		)
	}

	console.log(
		pc.dim(
			`Checked ${tasks.length} task${tasks.length === 1 ? '' : 's'} (${pendingCount} pending, ${inProgressCount} in progress, ${completedCount} completed)`,
		),
	)
}
