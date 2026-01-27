export type EditorOption = 'system' | 'vim' | 'nano' | 'default-app'

export interface Settings {
	theme: string
	fullscreen: boolean
	showCompleted: boolean
	editor: EditorOption
}

export const DEFAULT_SETTINGS: Settings = {
	theme: 'default',
	fullscreen: false,
	showCompleted: true,
	editor: 'system',
}

export const EDITOR_OPTIONS: {
	value: EditorOption
	label: string
	hint: string
}[] = [
	{ value: 'system', label: 'System default', hint: 'Uses $EDITOR or $VISUAL' },
	{ value: 'vim', label: 'Vim', hint: 'Open in Vim' },
	{ value: 'nano', label: 'Nano', hint: 'Open in Nano' },
	{
		value: 'default-app',
		label: 'Default app',
		hint: 'Open with system default (Preview, etc.)',
	},
]
