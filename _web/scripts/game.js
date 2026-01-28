import Notification from "./utils/notification"
import { EventAction } from "./utils/ws"
import BoardCanvas from "./utils/board"

/**
 * @typedef {Object} BoardState
 * @property {number[]} legalMoves
 * @property {number} whiteTime
 * @property {number} blackTime
 * @property {boolean} isWhiteTurn
 */
;(() => {
	// Page guard.
	if (document.getElementsByTagName("main")[0]?.dataset.page !== "game") {
		return
	}

	// Game id is the last element of the pathname.
	const id = window.location.pathname.split("/").at(-1)

	const notification = new Notification()

	// Initialize WebSocket connection.
	const socket = new WebSocket(`http://localhost:3502/ws?id=${id}`)

	socket.onclose = () => {
		notification.create("Please reload the page to reconnect.")
	}
	socket.onmessage = (raw) => {
		const e = JSON.parse(raw.data)
		const action = e.a
		const payload = e.p

		switch (action) {
			case EventAction.Ping:
				// Respond with pong.
				socket.send(JSON.stringify({ a: EventAction.Pong, p: null }))
				ping.textContent = `Ping: ${payload} ms`
				break

			default:
				notification.create("Unknown event recieved from server.")
		}
	}

	// Render board canvas.
	const sheet = new Image()
	sheet.src = "/images/sheet.svg"
	sheet.onload = () => {
		const board = new BoardCanvas(sheet, onMove)
		board.render()

		// Responsive board.
		const observer = new ResizeObserver((entries) => {
			for (const entry of entries) {
				board.setSize(entry.contentRect.width)
			}
		})
		observer.observe(boardCanvas)
	}

	/**
	 * Handles player's move.
	 */
	function onMove(from, to) {
		return false
	}
})()
