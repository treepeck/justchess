import showNotification from "/js/notification.js"
import { EventAction } from "/js/ws.js"

// Queue id is the last element of the pathname.
const id = window.location.pathname.split("/").at(-1)

// Initialize WebSocket connection.
const socket = new WebSocket(`http://localhost:3502/ws?id=${id}`)

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
			// Respond with pong.
			socket.send(JSON.stringify({ a: EventAction.Pong, p: null }))
			ping.textContent = `Ping: ${payload} ms`
			break

		case EventAction.ClientsCounter:
			clientsCounter.textContent = `Players in queue: ${payload}`
			break

		case EventAction.Redirect:
			// Redirect to game room.
			window.location.href = `http://localhost:3502/game/${payload}`
			break

		default:
			showNotification("Unknown event recieved from server.")
	}
}

// Countup timer.

switch (id) {
	case "2":
		timeControl.textContent = "Control: 2 minutes"
		timeBonus.textContent = "Bonus: 1 second"
		break
	case "3":
		timeControl.textContent = "Control: 3 minutes"
		timeBonus.textContent = "Bonus: 0 seconds"
		break
	case "4":
		timeControl.textContent = "Control: 3 minutes"
		timeBonus.textContent = "Bonus: 2 seconds"
		break
	case "5":
		timeControl.textContent = "Control: 5 minutes"
		timeBonus.textContent = "Bonus: 0 seconds"
		break
	case "6":
		timeControl.textContent = "Control: 5 minutes"
		timeBonus.textContent = "Bonus: 2 seconds"
		break
	case "7":
		timeControl.textContent = "Control: 10 minutes"
		timeBonus.textContent = "Bonus: 0 seconds"
		break
	case "8":
		timeControl.textContent = "Control: 10 minutes"
		timeBonus.textContent = "Bonus: 10 seconds"
		break
	case "9":
		timeControl.textContent = "Control: 15 minutes"
		timeBonus.textContent = "Bonus: 10 seconds"
		break
}
