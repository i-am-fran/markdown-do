export const MENU_FOOTER = '↑↓ navigate • enter select • esc esc quit'

export const LIST_FOOTER =
	'↑↓ navigate • enter select • c complete • d delete • e edit • m move • esc back'

export function getTaskFooter(taskId: number): string {
	return `CLI: mdd -t ${taskId} (toggle) • mdd -e ${taskId} (edit) • mdd -d ${taskId} (delete)`
}
