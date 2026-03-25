import { g, c } from "../utils/dom"
/**
 * Simple notification popup system.
 */
export default class MessageSystem {
	/**
	 * Message container.
	 * @type {HTMLDivElement}
	 */
	#container

	constructor() {
		this.#container = /** @type {HTMLDivElement} */ (
			c("div", "message-container")
		)

		// Append message container to the main element.
		g("main").appendChild(this.#container)
	}

	/**
	 * Creates the new message and appends it to the container.
	 * @param {string} text
	 */
	create(text) {
		const message = c(
			"div",
			"message",
			`message${this.#container.childNodes.length}`,
		)
		message.classList.add("flex")
		message.textContent = text
		// Create close button for notification.
		const btn = c("button", "message-close-button")
		btn.classList.add("button")
		btn.textContent = "X"
		btn.onclick = () => this.#container.removeChild(g(message.id))
		// Append close button to the message.
		message.appendChild(btn)
		// Append message to the container.
		this.#container.appendChild(message)
		// Focus the close button for accessibility.
		btn.focus()
	}
}
