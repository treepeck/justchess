/**
 * @param {string} id
 * @throws {Error}
 * @returns {HTMLElement}
 */
export function getOrPanic(id) {
	const el = document.getElementById(id)
	if (!el) throw new Error(`Missing element ${id}.`)
	return el
}

/**
 * @param {string} tagName
 * @param {string} className
 * @param {string} [id] - Optional.
 * @returns {HTMLElement}
 */
export function create(tagName, className, id) {
	const el = document.createElement(tagName)
	if (id) {
		el.id = id
	}
	if (className.length > 0) {
		el.classList.add(className)
	}

	return el
}
