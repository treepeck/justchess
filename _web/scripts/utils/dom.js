/**
 * Gets DOM element by ID.
 * @param {string} id
 * @returns {HTMLElement}
 * @throws {Error} When element with specified ID is missing.
 */
export function g(id) {
	const el = document.getElementById(id)
	if (!el) {
		throw new Error(`Element with ID ${id} is missing.`)
	}
	return el
}

/**
 * Created new DOM element.
 * @param {string} tagName
 * @param {string} [className] - Optional.
 * @param {string} [id] - Optional.
 * @returns {HTMLElement}
 */
export function c(tagName, className, id) {
	const el = document.createElement(tagName)
	if (className) {
		el.classList.add(className)
	}
	if (id) {
		el.id = id
	}
	return el
}
