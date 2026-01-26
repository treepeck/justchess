import Notification from "./utils/notification"
import { EventAction } from "./ws"
;(() => {
	// Page guard.
	if (document.getElementsByTagName("main")[0]?.dataset.page !== "queue") {
		return
	}

	const notification = new Notification()

	// Queue id is the last element of the pathname.
	const id = window.location.pathname.split("/").at(-1)

	// Initialize WebSocket connection.
	const socket = new WebSocket(`http://localhost:3502/ws?id=${id}`)

	socket.onclose = () => {
		notification.create(
			"Connection to the server was lost. Please reload the page."
		)
	}
	socket.onmessage = (raw) => {
		const { action, payload } = JSON.parse(raw.data)

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
				notification.create("Unknown event recieved from server.")
		}
	}

	const interval = 500 // Milliseconds.
	const initial = Date.now()
	let expected = initial + interval
	setTimeout(() => countUp(expected, initial, interval), interval)
})()

// Self-adjusting countup timer.
function countUp(expected, initial, interval) {
	const current = Date.now()
	const delta = current - expected
	if (delta > interval) {
		// Skip missing steps.
		expected += delta
	}

	expected += interval

	let seconds = Math.floor((current - initial) / 1000)
	const minutes = Math.floor(seconds / 60)
	if (minutes > 0) {
		seconds -= 60 * minutes
	}

	countUp.textContent = `${minutes > 9 ? minutes : `0${minutes}`}:${
		seconds > 9 ? seconds : `0${seconds}`
	}`

	setTimeout(
		() => countUp(expected, initial, interval),
		Math.max(0, interval - delta)
	)
}
