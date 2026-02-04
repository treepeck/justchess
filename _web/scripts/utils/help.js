import { getElement } from "./dom"

export default class HelpWindow {
	/**
	 * Shows help window with the specified id.
	 * @param {string} id
	 */
	static show(id) {
		getElement(id).classList.toggle("show")
		const btn = getElement(`close${id}`)
		btn.onclick = () => {
			HelpWindow.#hide(id)
		}

		// Focus close button for accessibility.
		btn.focus()
	}

	/**
	 * Hides help window with the specified id.
	 * @param {string} id
	 */
	static #hide(id) {
		getElement("id").classList.toggle("show")
		getElement(`close${id}`).onclick = null
	}
}
