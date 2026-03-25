import { g } from "./dom"

/**
 * Shows dialog with the specified id.
 * @param {string} id
 */
export default function showDialog(id) {
	const dialog = g(id)
	dialog.classList.add("show")

	const btn = g(`${id}Close`)
	btn.onclick = () => {
		dialog.classList.remove("show")
		// Remove event handler.
		btn.onclick = null
	}

	// Focus close btn for accessibility.
	btn.focus()
}
