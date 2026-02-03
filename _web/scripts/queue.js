import Notification from "./utils/notification"
import { EventAction } from "./ws/event"
;(() => {
	// Page guard.
	const container = document.getElementById("container")
	if (!container || container.dataset.page !== "queue") return

	const path = window.location.pathname.split("/")
	if (path.length < 2) {
		console.error("Invalid pathname.")
		return
	}
	const id = path[path.length - 1]

	const cnt = document.getElementById("clientsCounter")

	// @ts-expect-error - API_URL comes from webpack.
	const socket = new WebSocket(`${API_URL}/game/${id}`)

	const notification = new Notification()
	socket.onclose = () => {
		notification.create("Please reload the page to reconnect.")
	}
	socket.onerror = () => {
		notification.create("Please reload the page to reconnect.")
	}

	socket.onmessage = (raw) => {
		/** @type {import("./ws/event").Event} */
		const e = JSON.parse(raw.data)
		const action = e.a
		const payload = e.p

		switch (action) {
			case EventAction.Ping:
				// Respond with pong.
				socket.send(JSON.stringify({ a: EventAction.Pong, p: null }))
				const ping = document.getElementById("ping")
				if (ping) ping.textContent = `Ping: ${payload} ms`
				break

			case EventAction.ClientsCounter:
				if (cnt) {
					cnt.textContent = `Players in queue: ${payload}`
				}
				break

			case EventAction.Redirect:
				// Redirect to game room.
				// @ts-expect-error - API_URL comes from webpack.
				window.location.href = `${API_URL}/game/${payload}`
				break

			case EventAction.Error:
				notification.create(payload)
				break
		}
	}

	// Self-adjusting countup timer.
	const timer = document.getElementById("countUp")
	if (!timer) return

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

		timer.textContent = `${minutes > 9 ? minutes : `0${minutes}`}:${
			seconds > 9 ? seconds : `0${seconds}`
		}`

		setTimeout(() => countUpHandler(), Math.max(0, interval - delta))
	}
})()
