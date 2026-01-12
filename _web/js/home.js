import showNotification from "/js/notification.js"
import { EventAction } from "/js/ws.js"

// Initialize WebSocket connection.
const socket = new WebSocket("http://localhost:3502/ws?rid=home")

socket.onclose = () => {
	showNotification(
		"Connection to the server was lost. Please reload the page."
	)
}
socket.onerror = () => {
	showNotification(
		"Connection to the server was lost. Please reload the page."
	)
}
socket.onmessage = (raw) => {
	const msg = JSON.parse(raw.data)
	handleEvent(msg.a, msg.p)
}

function handleEvent(action, payload) {
	switch (action) {
		case EventAction.Ping:
			socket.send(JSON.stringify({ a: EventAction.Pong, p: null }))
			ping.textContent = "Ping: " + payload + " ms"
			break

		default:
			showNotification("Unknown event recieved from server.")
	}
}

function joinMatchmaking(cellId) {
	let payload = { c: 1, b: 0 }

	switch (cellId) {
		case 2:
			payload = { c: 2, b: 1 }
			break
		case 3:
			payload = { c: 3, b: 0 }
			break
		case 4:
			payload = { c: 3, b: 2 }
			break
		case 5:
			payload = { c: 5, b: 0 }
			break
		case 6:
			payload = { c: 5, b: 2 }
			break
		case 7:
			payload = { c: 10, b: 0 }
			break
		case 8:
			payload = { c: 10, b: 10 }
			break
		case 9:
			payload = { c: 15, b: 10 }
			break
	}

	socket.send(JSON.stringify({ a: EventAction.JoinMatchmaking, p: payload }))

	// TODO: show matchmaking window.
}

for (let i = 1; i <= 9; i++) {
	const cell = document.getElementById(`cell${i}`)
	cell.addEventListener("click", () => {
		joinMatchmaking(i)
	})
}

// Show and hide help window.
helpText.addEventListener("click", () => {
	helpWindow.classList.toggle("show")

	// Focus close button.
	closeHelp.focus()
})

closeHelp.addEventListener("click", () => {
	helpWindow.classList.toggle("show")
})
