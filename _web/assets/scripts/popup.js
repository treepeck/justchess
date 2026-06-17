import { get, make } from "/assets/scripts/dom.js"

/**
 * Creates a new popup (notification dialog with information text).
 * @param {string} text
 */
export function createPopup(text) {
	const dialog = /** @type {HTMLDialogElement} */ (make("dialog", "popup"))
	const destroy = () => {
		dialog.close()
		dialog.remove()
	}
	// Create a close button.
	dialog.textContent = text
	const btn = make("button", "popup-close")
	btn.textContent = "X"
	btn.onclick = destroy
	dialog.appendChild(btn)
	// TODO: code repetiton with dialog.js
	dialog.onclick = (e) => {
		const rect = dialog.getBoundingClientRect()

		const isInsideDialog =
			rect.left <= e.clientX &&
			e.clientX <= rect.left + rect.width &&
			rect.top <= e.clientY &&
			e.clientY <= rect.top + rect.height

		if (!isInsideDialog) destroy()
	}
	// Append dialog to main container.
	get("main").appendChild(dialog)
	dialog.showModal()
}
