import { SelectPrompt, isCancel } from '@clack/core'
import pc from 'picocolors'

const S_BAR = '│'
const S_BAR_END = '└'
const S_RADIO_ACTIVE = '●'
const S_RADIO_INACTIVE = '○'
const S_STEP_ACTIVE = '◆'
const S_STEP_CANCEL = '■'
const S_STEP_SUBMIT = '◇'

interface SelectOption<T> {
	value: T
	label: string
	hint?: string
	disabled?: boolean
}

interface SelectConfig<T> {
	message: string
	options: SelectOption<T>[]
	footer?: string
	initialValue?: T
	hotkeys?: Record<string, string>
}

export interface HotkeyResult<T> {
	action: string
	value: T
}

function symbol(state: string) {
	switch (state) {
		case 'initial':
		case 'active':
			return pc.cyan(S_STEP_ACTIVE)
		case 'cancel':
			return pc.red(S_STEP_CANCEL)
		case 'submit':
			return pc.green(S_STEP_SUBMIT)
		default:
			return pc.cyan(S_STEP_ACTIVE)
	}
}

function renderOption<T>(option: SelectOption<T>, state: string) {
	const label = option.label ?? String(option.value)
	switch (state) {
		case 'disabled':
			return `${pc.dim(S_RADIO_INACTIVE)} ${pc.dim(label)}${option.hint ? ` ${pc.dim(`(${option.hint ?? 'disabled'})`)}` : ''}`
		case 'selected':
			return pc.dim(label)
		case 'active':
			return `${pc.green(S_RADIO_ACTIVE)} ${label}${option.hint ? ` ${pc.dim(`(${option.hint})`)}` : ''}`
		case 'cancelled':
			return pc.strikethrough(pc.dim(label))
		default:
			return `${pc.dim(S_RADIO_INACTIVE)} ${label}`
	}
}

export async function selectWithFooter<T>(
	config: SelectConfig<T>,
): Promise<T | symbol | HotkeyResult<T>> {
	return new Promise((resolve) => {
		const prompt = new SelectPrompt({
			options: config.options,
			initialValue: config.initialValue,
			render() {
				const title = `${symbol(this.state)}  ${config.message}`
				const bar = pc.cyan(S_BAR)
				const barGray = pc.gray(S_BAR)
				const barEnd = pc.cyan(S_BAR_END)

				switch (this.state) {
					case 'submit': {
						const selected = renderOption(
							config.options[this.cursor],
							'selected',
						)
						return `${title}\n${barGray}  ${selected}`
					}
					case 'cancel': {
						const cancelled = renderOption(
							config.options[this.cursor],
							'cancelled',
						)
						return `${title}\n${barGray}  ${cancelled}\n${barGray}`
					}
					default: {
						const opts = config.options
							.map((opt, i) =>
								renderOption(
									opt,
									opt.disabled
										? 'disabled'
										: i === this.cursor
											? 'active'
											: 'inactive',
								),
							)
							.join(`\n${bar}  `)

						const footer = config.footer
							? `\n${barEnd}  ${pc.dim(config.footer)}`
							: `\n${barEnd}`

						return `${title}\n${bar}  ${opts}${footer}\n`
					}
				}
			},
		})

		// Handle custom hotkeys
		if (config.hotkeys) {
			prompt.on('key', (key) => {
				if (key && config.hotkeys && key in config.hotkeys) {
					const currentOption = config.options[prompt.cursor]
					// Don't trigger hotkeys on disabled options
					if (currentOption && !currentOption.disabled) {
						// Force close the prompt using type assertion for internal state
						const promptInternal = prompt as unknown as {
							state: string
							close: () => void
						}
						promptInternal.state = 'submit'
						promptInternal.close()
						resolve({
							action: config.hotkeys[key],
							value: currentOption.value,
						})
					}
				}
			})
		}

		prompt.prompt().then((result) => {
			resolve(result)
		})
	})
}

export function isHotkeyResult<T>(
	value: T | symbol | HotkeyResult<T>,
): value is HotkeyResult<T> {
	return (
		typeof value === 'object' &&
		value !== null &&
		'action' in value &&
		'value' in value
	)
}

export { isCancel }
