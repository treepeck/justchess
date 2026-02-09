import { getOrPanic } from "./dom"

/**
 * Shows help dialog with the specified id.
 * @param {string} id
 */
export default function showHelpDialog(id) {
	const dialog = getOrPanic(id)
	dialog.classList.toggle("show")

	const btn = getOrPanic(`${id}Close`)

	btn.onclick = () => {
		dialog.classList.toggle("show")
		// Remove event handler.
		btn.onclick = null
	}

	btn.focus()
}
