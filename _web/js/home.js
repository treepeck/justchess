import showNotification from "/js/notification.js"

// WebSocket connection.

const socket = new WebSocket("http://localhost:3502/ws?rid=home")

socket.onopen = () => {
}

socket.onclose = () => {
	showNotification("Connection to the server was lost. Please reload the page.")
}

socket.onerror = () => {
	showNotification("Connection to the server was lost. Please reload the page.")
}

socket.onmessage = (raw) => {
	const msg = JSON.parse(raw.data)
}

// Show and hide help window.
helpText.addEventListener("click", () => {
	helpWindow.classList.toggle("show")

	// Focus close button.
	closeHelp.focus()
})

closeHelp.addEventListener("click", () => { helpWindow.classList.toggle("show") })