import showNotification from "/js/notification.js"
import { EventAction } from "/js/ws.js"

// Initialize WebSocket connection.
const socket = new WebSocket("http://localhost:3502/ws?rid=home")

socket.onclose = () => { showNotification("Connection to the server was lost. Please reload the page.") }
socket.onerror = () => { showNotification("Connection to the server was lost. Please reload the page.") }
socket.onmessage = (raw) => {
	const msg = JSON.parse(raw.data)
	handleEvent(msg.a, msg.p)
}

function handleEvent(action, payload) {
	switch (action) {
		case EventAction.Ping:
			socket.send(JSON.stringify({a: EventAction.Pong, p: null}))
			ping.textContent = "Ping: " + payload + " ms"
			break;

		default:
			showNotification("Unknown event recieved from server.")
	}
}

// Show and hide help window.
helpText.addEventListener("click", () => {
	helpWindow.classList.toggle("show")

	// Focus close button.
	closeHelp.focus()
})

closeHelp.addEventListener("click", () => { helpWindow.classList.toggle("show") })