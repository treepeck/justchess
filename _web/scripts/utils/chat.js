/**
 * Appends the message to the DOM.
 * @param {string} message
 */
export default function addMessage(message) {
	const msgDiv = document.createElement("div")
	msgDiv.classList.add("message")
	msgDiv.textContent = message
	messageContainer.appendChild(msgDiv)

	// Reset the input value after submitting the message.
	chatInput.value = ""
}
