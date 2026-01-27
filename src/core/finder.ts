import { existsSync } from 'node:fs'
import { dirname, join } from 'node:path'
import fg from 'fast-glob'

const TODO_PATTERNS = ['**/TODO*.md', '**/todo*.md']
const DEFAULT_TODO_FILE = 'TODO.md'

export interface FoundFile {
	path: string
	relativePath: string
}

export async function findTodoFiles(
	cwd: string = process.cwd(),
	recursive = false,
): Promise<FoundFile[]> {
	const pattern = recursive ? TODO_PATTERNS : ['TODO*.md', 'todo*.md']

	const files = await fg(pattern, {
		cwd,
		absolute: true,
		caseSensitiveMatch: false,
		ignore: ['**/node_modules/**', '**/.git/**'],
	})

	return files.map((path) => ({
		path,
		relativePath: path.replace(`${cwd}/`, ''),
	}))
}

export async function findDefaultTodoFile(
	cwd: string = process.cwd(),
): Promise<string> {
	const defaultPath = join(cwd, DEFAULT_TODO_FILE)

	if (existsSync(defaultPath)) {
		return defaultPath
	}

	// Look for any TODO*.md in current directory
	const files = await findTodoFiles(cwd, false)
	if (files.length > 0) {
		return files[0].path
	}

	// Return default path (will be created if needed)
	return defaultPath
}

export async function findTodoFilesInSubdirs(
	cwd: string = process.cwd(),
): Promise<Map<string, FoundFile[]>> {
	const files = await findTodoFiles(cwd, true)
	const byDir = new Map<string, FoundFile[]>()

	for (const file of files) {
		const dir = dirname(file.relativePath)
		const dirKey = dir === '.' ? '(root)' : dir

		const existing = byDir.get(dirKey)
		if (existing) {
			existing.push(file)
		} else {
			byDir.set(dirKey, [file])
		}
	}

	return byDir
}
