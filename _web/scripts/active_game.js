import { getOrPanic, create } from "./utils/dom"
import { EventAction } from "./ws/event"
import { Socket } from "./ws/socket"
import Board from "./chess/board"
import { appendMoveToTable, Move } from "./chess/move"

/**
 * Appends chat message to the DOM.
 * @param {string} msg
 */
function appendChatMessage(msg) {
	const message = create("div", "message")
	message.textContent = msg

	// Append message to chat.
	const container = getOrPanic("chat")
	container.appendChild(message)

	// Scroll chat to bottom.
	container.scrollTo({
		top: container.scrollHeight,
		behavior: "smooth",
	})
}

;(() => {
	// Page guard.
	if (!document.getElementById("activeGameGuard")) return

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
				break
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
	const board = new Board(moveHandler)

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
