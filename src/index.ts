import {
	addNote,
	addTask,
	completeTask,
	deleteCompletedTasks,
	deleteTask,
	editTask,
	findTasksByKeyword,
	lintFile,
	listTasks,
	openInEditor,
	parseArgs,
	showHelp,
	showVersion,
	toggleTask,
} from './cli/commands.js'
import { runTui } from './tui/app.js'

async function main() {
	const args = process.argv.slice(2)
	const { text, value, flags } = parseArgs(args)

	// Version flag
	if (flags.has('v') || flags.has('version')) {
		showVersion()
		return
	}

	// Help flag
	if (flags.has('h') || flags.has('help')) {
		showHelp()
		return
	}

	// List tasks
	if (flags.has('l')) {
		await listTasks(false)
		return
	}

	// List tasks recursively
	if (flags.has('ls')) {
		await listTasks(true)
		return
	}

	// Delete task
	if (flags.has('d') && value) {
		await deleteTask(value)
		return
	}

	// Delete completed tasks
	if (flags.has('dc')) {
		await deleteCompletedTasks()
		return
	}

	// Lint file
	if (flags.has('lint')) {
		await lintFile()
		return
	}

	// Edit task
	if (flags.has('e') && value && text) {
		await editTask(value, text)
		return
	}

	// Toggle task
	if (flags.has('t') && value) {
		await toggleTask(value)
		return
	}

	// Complete task
	if (flags.has('c') && value) {
		await completeTask(value)
		return
	}

	// Add note
	if (flags.has('n') && value) {
		const noteText = text ? `${value} ${text}` : value
		await addNote(noteText)
		return
	}

	// Find tasks
	if (flags.has('f') && value) {
		await findTasksByKeyword(value, false)
		return
	}

	// Find tasks recursively
	if (flags.has('fs') && value) {
		await findTasksByKeyword(value, true)
		return
	}

	// Open in editor
	if (flags.has('o')) {
		await openInEditor()
		return
	}

	// Add task (positional arguments joined as text)
	if (text && !flags.size) {
		await addTask(text)
		return
	}

	// No arguments - open TUI
	await runTui()
}

main().catch((err) => {
	console.error('Error:', err.message)
	process.exit(1)
})
