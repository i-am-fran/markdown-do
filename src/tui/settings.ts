import * as p from '@clack/prompts'
import pc from 'picocolors'
import { EDITOR_OPTIONS, type EditorOption } from '../config/defaults.js'
import {
	getSettings,
	resetSettings,
	updateSettings,
} from '../config/settings.js'
import { LIST_FOOTER } from './hints.js'
import { isCancel, selectWithFooter } from './select.js'

export async function showSettingsMenu(): Promise<void> {
	while (true) {
		const settings = getSettings()

		console.log()

		const editorLabel =
			EDITOR_OPTIONS.find((e) => e.value === settings.editor)?.label ||
			'System default'

		const action = await selectWithFooter({
			message: 'Settings',
			options: [
				{
					value: 'editor' as const,
					label: `Editor: ${pc.cyan(editorLabel)}`,
				},
				{
					value: 'fullscreen' as const,
					label: `Fullscreen mode: ${settings.fullscreen ? pc.green('on') : pc.dim('off')}`,
				},
				{
					value: 'showCompleted' as const,
					label: `Show completed: ${settings.showCompleted ? pc.green('on') : pc.dim('off')}`,
				},
				{
					value: 'theme' as const,
					label: `Theme ${pc.dim('(coming soon)')}`,
					disabled: true,
				},
				{
					value: 'reset' as const,
					label: pc.yellow('Reset to defaults'),
				},
				{
					value: 'back' as const,
					label: pc.dim('Back'),
				},
			],
			footer: LIST_FOOTER,
		})

		if (isCancel(action) || action === 'back') break

		switch (action) {
			case 'editor':
				await selectEditor()
				break

			case 'fullscreen':
				await updateSettings({ fullscreen: !settings.fullscreen })
				p.log.success(
					`Fullscreen mode ${!settings.fullscreen ? 'enabled' : 'disabled'}`,
				)
				break

			case 'showCompleted':
				await updateSettings({ showCompleted: !settings.showCompleted })
				p.log.success(
					`Show completed ${!settings.showCompleted ? 'enabled' : 'disabled'}`,
				)
				break

			case 'reset': {
				const shouldReset = await p.confirm({
					message: 'Reset all settings to defaults?',
				})
				if (shouldReset === true) {
					await resetSettings()
					p.log.success('Settings reset')
				}
				break
			}
		}
	}
}

async function selectEditor(): Promise<void> {
	const settings = getSettings()

	const editor = await selectWithFooter({
		message: 'Select editor:',
		options: EDITOR_OPTIONS.map((opt) => ({
			value: opt.value,
			label:
				opt.value === settings.editor
					? `${opt.label} ${pc.green('(current)')}`
					: opt.label,
			hint: opt.hint,
		})),
		footer: LIST_FOOTER,
	})

	if (!isCancel(editor)) {
		await updateSettings({ editor: editor as EditorOption })
		p.log.success(
			`Editor set to ${EDITOR_OPTIONS.find((e) => e.value === editor)?.label}`,
		)
	}
}
