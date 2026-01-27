export const MENU_FOOTER = '↑↓ • enter • esc esc quit'

export const LIST_FOOTER =
	'↑↓ • enter • c complete • d delete • e edit • m move • esc back'

export function getTaskFooter(taskId: number): string {
	return `mdd -t ${taskId} • mdd -e ${taskId} • mdd -d ${taskId}`
}
