/**
 * Alias for document.getElementById. Throws an error if named element is missing.
 * @param {string} id
 * @returns {HTMLElement}
 * @throws {Error}
 */
export function get(id) {
	const element = document.getElementById(id)
	if (!element) throw new Error(`Element with id \"${id}\" is missing`)
	return element
}

/**
 * Creates a new element by calling document.createElement() and returns it.
 * @param {string} tagName
 * @param {string} [className]
 * @returns {HTMLElement}
 */
export function make(tagName, className) {
	const element = document.createElement(tagName)
	if (className) element.classList.add(className)
	return element
}
