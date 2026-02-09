import { getOrPanic, create } from "./utils/dom"
import { EventAction } from "./ws/event"
import { Socket } from "./ws/socket"
import Board from "./chess/board"
import { Move } from "./chess/move"

/**
 * Appends move SAN to moves table.
 * @param {string} san
 * @param {number} moveIndex
 */
function appendMoveToTable(san, moveIndex) {
	// Half move index.
	const ply = Math.ceil(moveIndex / 2)

	// Append row to the table after each black move.
	if (moveIndex % 2 !== 0) {
		const row = create("div", "move-table-row", `row${ply}`)
		// Append half-move index to the row.
		const ind = create("div", "move-table-ply")
		ind.textContent = `${ply}.`
		row.appendChild(ind)
		// Append row to the table.
		getOrPanic("moveTable").appendChild(row)
	}

	// Append move to the row.
	const move = create("div", "move-table-san")
	move.textContent = san
	getOrPanic(`row${ply}`).appendChild(move)

	// Scroll table to bottom.
	const table = getOrPanic(`moveTable`)
	table.scrollTo({
		top: table.scrollHeight,
		behavior: "smooth",
	})
}

/**
 * Appends chat message to the DOM.
 * @param {string} msg
 */
function appendChatMessage(msg) {
	const message = create("div", "message")
	message.textContent = msg

	// Append message to chat.
	const container = getOrPanic("chatMessages")
	container.appendChild(message)

	// Scroll chat to bottom.
	container.scrollTo({
		top: container.scrollHeight,
		behavior: "smooth",
	})
}

;(() => {
	// Page guard.
	if (!document.getElementById("gameGuard")) return

	/** @param {import("./chess/move").CompletedMove} move */
	const store = (move) => {
		// Store completed move.
		moves.push(move)
		// @ts-ignore - Call Move constructor to correctly initialize fields.
		board.makeMove(new Move(move.m))
		appendMoveToTable(move.s, moves.length)
	}

	/** @type {import("./ws/socket").EventHandler} */
	const eventHandler = (action, payload) => {
		switch (action) {
			case EventAction.Chat:
				appendChatMessage(payload)
				break
			case EventAction.Conn:
				appendChatMessage(`Player ${payload} joined`)
				break
			case EventAction.Disc:
				appendChatMessage(`Player ${payload} leaved`)
			// Synchronize game state.
			case EventAction.Game:
				/**
				 * Decode payload.
				 * @type {import("./ws/event").GamePayload}
				 */
				const pGame = { ...payload }
				// Update legal moves.
				board.setLegalMoves(pGame.lm)
				for (const move of pGame.m) {
					store(move)
				}
				break
			case EventAction.Move:
				/**
				 * Decode payload.
				 * @type {import("./ws/event").MovePayload}
				 */
				const pMove = { ...payload }
				// Update legal moves.
				board.setLegalMoves(pMove.lm)
				store(pMove.m)
				break
		}
	}

	const socket = new Socket(eventHandler)

	/** @type {import("./chess/board").MoveHandler} */
	const moveHandler = (moveIndex) => {
		socket.sendJSON(EventAction.Move, moveIndex)
	}

	// Render chessboard.
	const el = /** @type {HTMLDivElement} */ (getOrPanic("board"))
	const board = new Board(el, moveHandler)

	// Add board event listeners.
	el.onmousedown = (e) => board.onMouseDown(e)
	el.onmousemove = (e) => board.onMouseMove(e)
	el.onmouseup = (e) => board.onMouseUp(e)

	// Make board responsive.
	const observer = new ResizeObserver((entries) => {
		for (const entry of entries) {
			board.setSize(entry.contentRect.width)
		}
	})
	observer.observe(el)

	// Completed moves storage.
	const moves = /** @type {import("./chess/move").CompletedMove[]} */ ([])

	// Handle chat messages.
	const chat = /** @type {HTMLInputElement} */ (getOrPanic("chatInput"))
	const sendChat = () => {
		if (chat.value.length < 1) return
		socket.sendJSON(EventAction.Chat, chat.value)
		// Reset the input value after submitting the message.
		chat.value = ""
	}
	// Handle chat messages.
	getOrPanic("chatSend").onclick = () => sendChat()
	chat.onkeydown = (ev) => {
		if (ev.key === "Enter") sendChat()
	}
})()
