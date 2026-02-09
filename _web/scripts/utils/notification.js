import { getOrPanic, create } from "./dom"

export default class Notification {
	/**
	 * Notification container.
	 * @type {HTMLDivElement}
	 */
	#container
	/**
	 * How many notifications are currently displayed on the screen.
	 * @type {number}
	 */
	#count

	/**
	 * Appends notification container to the DOM.
	 */
	constructor() {
		this.#container = /** @type {HTMLDivElement} */ (
			create("div", "notification-container")
		)

		// Append notification container to the main container.
		getOrPanic("main").appendChild(this.#container)

		this.#count = 0
	}

	/**
	 * Creates the new notification and appends it to the container.
	 * @param {string} message
	 */
	create(message) {
		const notification = create(
			"div",
			"notification",
			`notification${this.#count}`
		)
		notification.textContent = message

		// Create close button for notification.
		const btn = create("button", "notification-close-button")
		btn.textContent = "X"
		btn.onclick = () => {
			this.#remove(notification.id)
		}

		// Append close button to the nofitication.
		notification.appendChild(btn)

		// Append notification to the container.
		this.#container.appendChild(notification)

		// Focus the close button for accessibility.
		btn.focus()

		this.#count++
	}

	/**
	 * Removes notification with specified id from the container.
	 * @param {string} id
	 */
	#remove(id) {
		const notification = document.getElementById(id)
		if (notification) {
			this.#container.removeChild(notification)
			this.#count--
		}
	}
}
