import { existsSync, mkdirSync, readFileSync, writeFileSync } from 'node:fs'
import { homedir } from 'node:os'
import { join } from 'node:path'
import { DEFAULT_SETTINGS, type Settings } from './defaults.js'

const CONFIG_DIR = join(homedir(), '.config', 'markdowndo')
const CONFIG_FILE = join(CONFIG_DIR, 'config.json')

let cachedSettings: Settings | null = null

function ensureConfigDir(): void {
	if (!existsSync(CONFIG_DIR)) {
		mkdirSync(CONFIG_DIR, { recursive: true })
	}
}

export function getSettings(): Settings {
	if (cachedSettings) return cachedSettings

	try {
		if (existsSync(CONFIG_FILE)) {
			const content = readFileSync(CONFIG_FILE, 'utf-8')
			cachedSettings = { ...DEFAULT_SETTINGS, ...JSON.parse(content) }
		} else {
			cachedSettings = { ...DEFAULT_SETTINGS }
		}
	} catch {
		cachedSettings = { ...DEFAULT_SETTINGS }
	}

	return cachedSettings
}

async function saveSettings(settings: Settings): Promise<void> {
	ensureConfigDir()
	writeFileSync(CONFIG_FILE, JSON.stringify(settings, null, 2))
	cachedSettings = settings
}

export async function updateSettings(
	partial: Partial<Settings>,
): Promise<void> {
	const current = getSettings()
	const updated = { ...current, ...partial }
	await saveSettings(updated)
}

export async function resetSettings(): Promise<void> {
	await saveSettings({ ...DEFAULT_SETTINGS })
}
