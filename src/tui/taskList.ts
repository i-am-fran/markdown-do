import * as p from '@clack/prompts'
import pc from 'picocolors'
import { getSettings } from '../config/settings.js'
import type { Task } from '../core/task.js'
import type { TodoFile } from '../core/todo.js'
import { LIST_FOOTER, getTaskFooter } from './hints.js'
import {
	type HotkeyResult,
	isCancel,
	isHotkeyResult,
	selectWithFooter,
} from './select.js'

const TASK_HOTKEYS = {
	c: 'complete',
	d: 'delete',
	e: 'edit',
	m: 'move',
}

export async function showTaskList(
	todoFile: TodoFile,
): Promise<number | 'add' | null | symbol> {
	const settings = getSettings()

	while (true) {
		const tasksBySection = todoFile.getTasksGroupedBySection()

		if (tasksBySection.size === 0) {
			p.log.warn('No tasks found')
			return null
		}

		const options: {
			value: number | string
			label: string
			hint?: string
			disabled?: boolean
		}[] = []
		let hasVisibleTasks = false

		for (const [section, tasks] of tasksBySection) {
			// Filter tasks based on showCompleted setting
			const visibleTasks = settings.showCompleted
				? tasks
				: tasks.filter((t) => t.status !== 'completed')

			if (visibleTasks.length === 0) continue

			hasVisibleTasks = true

			if (section !== null) {
				// Add section header as a disabled visual separator
				options.push({
					value: `section:${section}`,
					label: pc.bold(pc.magenta(`── ${section} ──`)),
					disabled: true,
				})
			}

			for (const task of visibleTasks) {
				options.push({
					value: task.id,
					label: formatTaskOption(task),
				})
			}
		}

		if (!hasVisibleTasks) {
			p.log.warn(
				settings.showCompleted
					? 'No tasks found'
					: 'No pending tasks (completed tasks hidden)',
			)
			return null
		}

		console.log()

		const result = await selectWithFooter({
			message: 'Select a task:',
			options: [
				...options,
				{ value: 'add', label: pc.green('+ Add new task') },
				{ value: -1, label: pc.dim('Back') },
			],
			footer: LIST_FOOTER,
			hotkeys: TASK_HOTKEYS,
		})

		// Handle hotkey actions
		if (isHotkeyResult(result)) {
			const taskId = result.value as number
			const handled = await handleHotkeyAction(todoFile, taskId, result.action)
			if (handled) continue // Stay in the list
			return taskId // Fallback to menu
		}

		// If a section header was selected, treat it as no selection
		if (typeof result === 'string' && result.startsWith('section:')) {
			continue
		}

		// Quick add action
		if (result === 'add') {
			return 'add'
		}

		return result as number | symbol
	}
}

async function handleHotkeyAction(
	todoFile: TodoFile,
	taskId: number,
	action: string,
): Promise<boolean> {
	const task = todoFile.getTask(taskId)
	if (!task) return false

	switch (action) {
		case 'complete': {
			const prevStatus = task.status
			todoFile.toggleTask(taskId)
			await todoFile.save()
			const messageMap = {
				pending: 'Task started',
				'in-progress': 'Task completed',
				completed: 'Task reopened',
			}
			p.log.success(messageMap[prevStatus])
			return true
		}

		case 'delete': {
			const shouldDelete = await p.confirm({
				message: `Delete "${task.text}"?`,
			})
			if (shouldDelete === true) {
				todoFile.deleteTask(taskId)
				await todoFile.save()
				p.log.success('Task deleted')
			}
			return true
		}

		case 'edit': {
			const newText = await p.text({
				message: 'Enter new text:',
				initialValue: task.text,
				validate: (value) => {
					if (!value.trim()) return 'Task cannot be empty'
				},
			})
			if (!p.isCancel(newText)) {
				todoFile.updateTask(taskId, newText as string)
				await todoFile.save()
				p.log.success('Task updated')
			}
			return true
		}

		case 'move': {
			await showMoveToSectionMenu(todoFile, taskId, task)
			return true
		}
	}

	return false
}

function formatTaskOption(task: Task): string {
	const id = pc.dim(`${task.id}.`)
	let checkbox: string
	let text: string
	switch (task.status) {
		case 'completed':
			checkbox = pc.dim(pc.green('[x]'))
			text = pc.strikethrough(pc.dim(task.text))
			break
		case 'in-progress':
			checkbox = pc.dim(pc.yellow('[/]'))
			text = pc.yellow(task.text)
			break
		default:
			checkbox = pc.dim('[ ]')
			text = task.text
	}
	return `${id} ${checkbox} ${text}`
}

export async function showTaskMenu(
	todoFile: TodoFile,
	taskId: number,
): Promise<void> {
	if (taskId === -1) return

	const task = todoFile.getTask(taskId)
	if (!task) return

	console.log()

	const toggleLabelMap = {
		pending: 'Start progress',
		'in-progress': 'Mark as complete',
		completed: 'Reopen task',
	}

	const action = await selectWithFooter({
		message: `Task: ${task.text}`,
		options: [
			{
				value: 'toggle',
				label: toggleLabelMap[task.status],
			},
			{ value: 'edit', label: 'Edit text' },
			{ value: 'move', label: 'Move to section' },
			{ value: 'delete', label: pc.red('Delete') },
			{ value: 'back', label: pc.dim('Back') },
		],
		footer: getTaskFooter(taskId),
	})

	if (isCancel(action) || action === 'back') return

	switch (action) {
		case 'toggle': {
			const prevStatus = task.status
			todoFile.toggleTask(taskId)
			await todoFile.save()
			const messageMap = {
				pending: 'Task started',
				'in-progress': 'Task completed',
				completed: 'Task reopened',
			}
			p.log.success(messageMap[prevStatus])
			break
		}

		case 'edit': {
			const newText = await p.text({
				message: 'Enter new text:',
				initialValue: task.text,
				validate: (value) => {
					if (!value.trim()) return 'Task cannot be empty'
				},
			})

			if (!p.isCancel(newText)) {
				todoFile.updateTask(taskId, newText as string)
				await todoFile.save()
				p.log.success('Task updated')
			}
			break
		}

		case 'move': {
			await showMoveToSectionMenu(todoFile, taskId, task)
			break
		}

		case 'delete': {
			const shouldDelete = await p.confirm({
				message: 'Are you sure you want to delete this task?',
			})

			if (shouldDelete === true) {
				todoFile.deleteTask(taskId)
				await todoFile.save()
				p.log.success('Task deleted')
			}
			break
		}
	}
}

async function showMoveToSectionMenu(
	todoFile: TodoFile,
	taskId: number,
	task: { section: string | null },
): Promise<void> {
	const sections = todoFile.getSectionNames()

	const options: { value: string | null; label: string }[] = []

	// Add Inbox option (null section)
	if (task.section !== null) {
		options.push({ value: null, label: 'Inbox (no section)' })
	}

	// Add existing sections (except current)
	for (const section of sections) {
		if (section !== task.section) {
			options.push({ value: section, label: section })
		}
	}

	// Add option to create new section
	options.push({ value: '__new__', label: pc.green('+ Create new section') })
	options.push({ value: '__back__', label: pc.dim('Back') })

	console.log()

	const selected = await selectWithFooter({
		message: 'Move to section:',
		options,
		footer: LIST_FOOTER,
	})

	if (isCancel(selected) || selected === '__back__') return

	let targetSection: string | null

	if (selected === '__new__') {
		const newSection = await p.text({
			message: 'Enter section name:',
			validate: (value) => {
				if (!value.trim()) return 'Section name cannot be empty'
			},
		})

		if (p.isCancel(newSection)) return

		targetSection = (newSection as string).trim()
	} else {
		targetSection = selected as string | null
	}

	todoFile.moveTask(taskId, targetSection)
	await todoFile.save()

	const targetName = targetSection ?? 'Inbox'
	p.log.success(`Moved to ${targetName}`)
}
