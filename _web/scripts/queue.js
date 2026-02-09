import { getOrPanic } from "./utils/dom"
import { EventAction } from "./ws/event"
import { Socket } from "./ws/socket"

/** @type {import("./ws/socket").EventHandler} */
function eventHandler(action, payload) {
	switch (action) {
		case EventAction.ClientsCounter:
			// Update clients counter.
			getOrPanic(
				"clientsCounter"
			).textContent = `Players in queue: ${payload}`
			break

		case EventAction.Redirect:
			// Redirect to game room.
			// @ts-expect-error - API_URL comes from webpack.
			window.location.href = `${API_URL}/game/${payload}`
			break

		default:
			throw new Error("Invalid event from server")
	}
}

;(() => {
	// Page guard.
	if (!document.getElementById("queueGuard")) return

	// Initialize connection.
	new Socket(eventHandler)

	// Self-adjusting countup timer.
	const interval = 500 // Milliseconds.
	const initial = Date.now()
	let expected = initial + interval
	setTimeout(() => countUpHandler(), interval)

	const countUpHandler = () => {
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

		getOrPanic("countUpTimer").textContent = `${
			minutes > 9 ? minutes : `0${minutes}`
		}:${seconds > 9 ? seconds : `0${seconds}`}`

		setTimeout(() => countUpHandler(), Math.max(0, interval - delta))
	}
})()
