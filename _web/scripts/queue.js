import Notification from "./utils/notification"
import { Socket, EventAction } from "./utils/ws"
;(() => {
	// Page guard.
	const container = document.getElementById("container")
	if (!container || container.dataset.page !== "queue") {
		return
	}

	const _ = new Socket(eventHandler, closeHandler)
	const eventHandler = (action, payload) => {
		switch (action) {
			case EventAction.ClientsCounter:
				clientsCounter.textContent = `Players in queue: ${payload}`
				break

			case EventAction.Redirect:
				// Redirect to game room.
				window.location.href = `${API_URL}/game/${payload}`
				break

			case EventAction.Error:
				notification.create(payload)
				break
		}
	}
	const closeHandler = () => {
		notification.create("Please reload the page to reconnect.")
	}

	const notification = new Notification()

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
		countUp.textContent = `${minutes > 9 ? minutes : `0${minutes}`}:${
			seconds > 9 ? seconds : `0${seconds}`
		}`

		setTimeout(() => countUpHandler(), Math.max(0, interval - delta))
	}
})()
