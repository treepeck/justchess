import { formatTime } from "./utils/clock"
import { getOrPanic } from "./utils/dom"
import { EventKind } from "./ws/event"
import { Socket } from "./ws/socket"
import { request } from "./utils/http"
import showDialog from "./utils/dialog"

/** @type {import("./ws/socket").EventHandler} */
function eventHandler(Kind, payload) {
	switch (Kind) {
		case EventKind.ClientsCounter:
			// Update clients counter.
			getOrPanic("clientsCounter").textContent =
				`Players in queue: ${payload}`

			if (payload < 2) {
				getOrPanic("playVsEngine").onclick = () => {
					request("/play-vs-engine", "POST", null)
				}
				showDialog("emptyQueueDialog")
			}
			break

		case EventKind.Redirect:
			// Redirect to game room.
			// @ts-expect-error - API_URL comes from webpack.
			window.location.href = `${API_URL}/${payload}`
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
	const interval = 1000 // Milliseconds.
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
		formatTime("countUpTimer", Math.floor(current - initial))

		setTimeout(() => countUpHandler(), Math.max(0, interval - delta))
	}
})()
