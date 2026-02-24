import {
	formatTermination,
	formatResult,
	Termination,
	Result,
} from "./chess/state"
import { appendMoveToTable, highlightCurrentMove } from "./chess/move"
import { getOrPanic, create } from "./utils/dom"
import { Clock, Color } from "./utils/clock"
import { EventAction } from "./ws/event"
import showDialog from "./utils/dialog"
import { Socket } from "./ws/socket"
import Board from "./chess/board"

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
		// Update position.
		board.parsePiecePlacement(move.f)
		board.fens.push(move.f)
		board.currentFen = board.fens.length - 1

		appendMoveToTable(move.s, board.currentFen, (index) => {
			board.currentFen = index
			highlightCurrentMove(index)
			board.parsePiecePlacement(board.fens[index])
		})

		highlightCurrentMove(board.currentFen)
	}

	const clock = new Clock(5 * 60 * 1000, false, Color.White, 1000)

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
				// Set player's clock.
				clock.setTime(Color.White, pGame.wt * 1000)
				clock.setTime(Color.Black, pGame.bt * 1000)
				clock.color =
					board.currentFen % 2 !== 0 ? Color.Black : Color.White
				if (pGame.t == Termination.Unterminated) {
					clock.start()
				} else {
					getOrPanic("endgameDialogResult").textContent =
						formatResult(pGame.r)
					getOrPanic("endgameDialogTermination").textContent =
						formatTermination(pGame.t)
					showDialog("endgameDialog")
				}
				break
			case EventAction.End:
				/**
				 * Decode payload.
				 * @type {import("./ws/event").EndPayload}
				 */
				const pEnd = { ...payload }

				getOrPanic("endgameDialogResult").textContent = formatResult(
					pEnd.r,
				)
				getOrPanic("endgameDialogTermination").textContent =
					formatTermination(pEnd.t)
				showDialog("endgameDialog")
				clock.stop()
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
				// Update player's clock.
				if (board.currentFen % 2 !== 0) {
					clock.setTime(Color.White, pMove.t * 1000)
				} else {
					clock.setTime(Color.Black, pMove.t * 1000)
				}
				clock.flip()
				break
			case EventAction.OfferDraw:
				// Display draw offer window.
				showDialog("acceptDrawDialog")
				getOrPanic("acceptDrawDialogAccept").onclick = () => {
					getOrPanic("acceptDrawDialog").classList.toggle("show")
					socket.sendJSON(EventAction.AcceptDraw, null)
				}
				getOrPanic("acceptDrawDialogClose").onclick = () => {
					getOrPanic("acceptDrawDialog").classList.toggle("show")
					socket.sendJSON(EventAction.DeclineDraw, null)
				}
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

	getOrPanic("offerDrawBtn").onclick = () => {
		showDialog("offerDrawDialog")
		getOrPanic("offerDrawDialogAccept").onclick = () => {
			getOrPanic("offerDrawDialog").classList.toggle("show")
			socket.sendJSON(EventAction.OfferDraw, null)
		}
	}

	getOrPanic("resignBtn").onclick = () => {
		showDialog("resignDialog")
		getOrPanic("resignDialogAccept").onclick = () => {
			getOrPanic("resignDialog").classList.toggle("show")
			socket.sendJSON(EventAction.Resign, null)
		}
	}

	// Go through move history using keyboard.
	document.onkeydown = (e) => {
		switch (e.key) {
			case "ArrowUp":
				board.currentFen = board.fens.length - 1
				break
			case "ArrowRight":
				if (board.currentFen == board.fens.length - 1) return
				board.currentFen += 1
				break
			case "ArrowDown":
				board.currentFen = 0
				break
			case "ArrowLeft":
				if (board.currentFen == 0) return
				board.currentFen -= 1
				break
		}
		highlightCurrentMove(board.currentFen)
		board.parsePiecePlacement(board.fens[board.currentFen])
	}
})()
