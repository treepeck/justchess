import Notification from "./utils/notification"
import { EventAction } from "./ws"
import Board from "./utils/board"
;(() => {
	// Page guard.
	if (document.getElementsByTagName("main")[0]?.dataset.page !== "game") {
		return
	}

	const notification = new Notification()

	// Game id is the last element of the pathname.
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

			default:
				notification.create("Unknown event recieved from server.")
		}
	}

	const sheet = new Image()
	sheet.src = "/images/sheet.svg"
	sheet.onload = () => {
		const ctx = boardCanvas.getContext("2d")
		const board = new Board(ctx, sheet)

		const observer = new ResizeObserver((entries) => {
			for (const entry of entries) {
				board.setSize(entry.contentRect.width)
			}
		})
		observer.observe(boardCanvas)

		// Add event listeners.
		boardCanvas.addEventListener("mousedown", (e) => {
			board.onMouseDown(e)
		})
		boardCanvas.addEventListener("mousemove", (e) => {
			board.onMouseMove(e)
		})
		boardCanvas.addEventListener("mouseup", (e) => {
			board.onMouseUp(e)
		})
	}
})()
