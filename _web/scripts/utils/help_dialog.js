import { getElement } from "./dom"

/**
 * Shows help dialog with the specified id.
 * @param {string} id
 */
export default function showHelpDialog(id) {
	const dialog = getElement(id)
	dialog.classList.toggle("show")

	const btn = getElement(`${id}CloseButton`)

	btn.onclick = () => {
		dialog.classList.toggle("show")
		// Remove event handler.
		btn.onclick = null
	}

	btn.focus()
}
