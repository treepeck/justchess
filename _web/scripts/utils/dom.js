/**
 * @param {string} id
 * @throws {Error}
 * @returns {HTMLElement}
 */
export function getElement(id) {
	const el = document.getElementById(id)
	if (!el) throw new Error(`Missing element ${id}.`)
	return el
}
