// Page guard.
if (document.body.dataset.page !== "game") {
	return
}

import showNotification from "./notification"
import { EventAction } from "./ws"
import Board from "./board"

// Game id is the last element of the pathname.
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
