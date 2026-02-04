import Notification from "./utils/notification"
import { EventAction } from "./ws/event"
import BoardCanvas from "./chess/board"
import { Move } from "./chess/move"
import { getElement } from "./utils/dom"

/** @param {string} san */
function appendMove(san) {
	const moveDiv = document.createElement("div")
	moveDiv.classList.add("move")
	moveDiv.textContent = san

	getElement("tableBody").appendChild(moveDiv)
}

/**
 * Appends the message to the DOM.
 * @param {string} message
 */
function appendMessage(message) {
	const msgDiv = document.createElement("div")
	msgDiv.classList.add("message")
	msgDiv.textContent = message
	getElement("messageContainer").appendChild(msgDiv)
}

/**
 * @returns {Promise<HTMLImageElement>}
 */
function loadSheet() {
	return new Promise((resolve, reject) => {
		const sheet = new Image()
		sheet.onload = () => resolve(sheet)
		sheet.onerror = (err) => reject(err)

		sheet.src = "/images/sheet.svg"
	})
}

async function main() {
	// Page guard.
	const container = document.getElementById("container")
	if (!container || container.dataset.page !== "game") {
		return
	}

	const path = window.location.pathname.split("/")
	if (path.length < 2) {
		console.error("Invalid pathname.")
		return
	}
	const id = path[path.length - 1]

	// Load sprite sheet.
	/** @type {HTMLImageElement | null} */
	let sheet = null
	try {
		sheet = await loadSheet()
	} catch (err) {
		console.error(err)
	}
	if (!sheet) return

	// Render chessboard on the canvas.
	const canvas = /** @type {HTMLCanvasElement} */ (getElement("boardCanvas"))

	const ctx = /** @type {CanvasRenderingContext2D} */ (
		canvas.getContext("2d")
	)

	const board = new BoardCanvas(ctx, sheet, (moveIndex) => {
		socket.send(JSON.stringify({ a: EventAction.Move, p: moveIndex }))
	})
	board.render()

	// Add event listeners.
	canvas.onmousedown = (e) => board.onMouseDown(e)
	canvas.onmousemove = (e) => board.onMouseMove(e)
	canvas.onmouseup = (e) => board.onMouseUp(e)

	// Make board responsive.
	const observer = new ResizeObserver((entries) => {
		for (const entry of entries) {
			board.setSize(entry.contentRect.width)
		}
	})
	observer.observe(canvas)

	// Initialize WebSocket connection.
	// @ts-expect-error - API_URL comes from webpack.
	const socket = new WebSocket(`${WS_URL}/ws?id=${id}`)

	const notification = new Notification()
	socket.onerror = () => {
		notification.create("Please reload the page to reconnect.")
	}

	// Handle chat messages.
	const chat = /** @type {HTMLInputElement} */ (getElement("chatInput"))
	const sendChat = () => {
		if (chat.value.length < 1) return

		appendMessage("You: " + chat.value)

		socket.send(
			JSON.stringify({
				a: EventAction.Chat,
				p: chat.value,
			})
		)
		// Reset the input value after submitting the message.
		chat.value = ""
	}
	// Handle chat messages.
	chat.onclick = () => sendChat()
	chat.onkeydown = (e) => {
		if (e.key === "Enter") sendChat()
	}

	// Handle messages.
	socket.onmessage = (raw) => {
		/** @type {import("./ws/event").Event} */
		const e = JSON.parse(raw.data)
		const action = e.a
		const payload = e.p

		switch (action) {
			case EventAction.Ping:
				// Respond with pong.
				socket.send(JSON.stringify({ a: EventAction.Pong, p: null }))
				getElement("ping").textContent = `Ping: ${payload} ms`
				break

			case EventAction.Chat:
				appendMessage(JSON.parse(payload))
				break

			case EventAction.Game:
				/** @type {import("./ws/event").GamePayload} */
				const pGame = { ...payload }

				board.setLegalMoves(pGame.lm)

				for (const completedMove of pGame.m) {
					// @ts-expect-error
					const move = new Move(completedMove.m)
					board.makeMove(move)
					appendMove(completedMove.s)
				}

				getElement("isWhiteConnected").textContent = payload.w
				getElement("isBlackConnected").textContent = payload.b

				break

			case EventAction.Move:
				/** @type {import("./ws/event").MovePayload} */
				const pMove = { ...payload }

				board.setLegalMoves(pMove.lm)

				appendMove(pMove.m.s)

				// @ts-expect-error
				const move = new Move(pMove.m.m)
				board.makeMove(move)
				break
		}
	}
}

main()
