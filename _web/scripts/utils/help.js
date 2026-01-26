export default class HelpWindow {
	/**
	 * Shows help window with the specified id.
	 * @param {string} id
	 */
	static show(id) {
		const help = document.getElementById(id)
		if (help) {
			help.classList.toggle("show")

			const btn = document.getElementById(`close${id}`)
			btn.onclick = () => {
				HelpWindow.#hide(id)
			}

			// Focus close button for accessibility.
			btn.focus()
		}
	}

	/**
	 * Hides help window with the specified id.
	 * @param {string} id
	 */
	static #hide(id) {
		const help = document.getElementById(id)
		if (help) {
			help.classList.toggle("show")

			const btn = document.getElementById(`close${id}`)
			btn.onlick = null
		}
	}
}
