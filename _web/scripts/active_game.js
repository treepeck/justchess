import {
	formatTermination,
	formatResult,
	Termination,
	Result,
} from "./chess/state"
import { appendMoveToTable, highlightCurrentMove } from "./chess/move"
import { getOrPanic, create } from "./utils/dom"
import { Clock, Color } from "./utils/clock"
import { EventKind } from "./ws/event"
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

	/** @param {import("./chess/move").PlayedMove} move */
	const store = (move) => {
		const piecePlacement = move.f.split(" ")[0]
		// Update position.
		board.parsePiecePlacement(piecePlacement)
		board.fens.push(piecePlacement)
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
	const eventHandler = (Kind, payload) => {
		switch (Kind) {
			case EventKind.Chat:
				appendChatMessage(payload)
				break
			case EventKind.Conn:
				appendChatMessage(`Player ${payload} joins`)
				break
			case EventKind.Disc:
				appendChatMessage(`Player ${payload} leaves`)
				break
			// Synchronize game state.
			case EventKind.Game:
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
				if (pGame.wt && pGame.bt) {
					clock.setTime(Color.White, pGame.wt * 1000)
					clock.setTime(Color.Black, pGame.bt * 1000)
					clock.start()
					clock.color =
						board.currentFen % 2 !== 0 ? Color.Black : Color.White
				}
				break
			case EventKind.End:
				clock.stop()

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
				break
			case EventKind.Move:
				/**
				 * Decode payload.
				 * @type {import("./ws/event").MovePayload}
				 */
				const pMove = { ...payload }
				// Update legal moves.
				board.setLegalMoves(pMove.lm)
				store({ s: pMove.s, f: pMove.f })
				clock.flip()
				break
			case EventKind.OfferDraw:
				// Display draw offer window.
				showDialog("acceptDrawDialog")
				getOrPanic("acceptDrawDialogAccept").onclick = () => {
					getOrPanic("acceptDrawDialog").classList.toggle("show")
					socket.sendJSON(EventKind.AcceptDraw, null)
				}
				getOrPanic("acceptDrawDialogClose").onclick = () => {
					getOrPanic("acceptDrawDialog").classList.toggle("show")
					socket.sendJSON(EventKind.DeclineDraw, null)
				}
				break
		}
	}

	const socket = new Socket(eventHandler)

	/** @type {import("./chess/board").MoveHandler} */
	const moveHandler = (moveIndex) => {
		socket.sendJSON(EventKind.Move, moveIndex)
	}

	// Render chessboard.
	const board = new Board(
		moveHandler,
		getOrPanic("boardContainer").classList.contains("flipped"),
	)

	// Handle chat messages.
	const chat = /** @type {HTMLInputElement} */ (getOrPanic("chatInput"))
	const sendChat = () => {
		if (chat.value.length < 1) return
		socket.sendJSON(EventKind.Chat, chat.value)
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
			socket.sendJSON(EventKind.OfferDraw, null)
		}
	}

	getOrPanic("resignBtn").onclick = () => {
		showDialog("resignDialog")
		getOrPanic("resignDialogAccept").onclick = () => {
			getOrPanic("resignDialog").classList.toggle("show")
			socket.sendJSON(EventKind.Resign, null)
		}
	}

	const setCurrentMove = () => {
		highlightCurrentMove(board.currentFen)
		board.parsePiecePlacement(board.fens[board.currentFen])
	}

	// Go through move history using buttons.
	getOrPanic("nullMoveBtn").onclick = () => {
		board.currentFen = 0
		setCurrentMove()
	}
	getOrPanic("prevMoveBtn").onclick = () => {
		if (board.currentFen == 0) return
		board.currentFen -= 1
		setCurrentMove()
	}
	getOrPanic("nextMoveBtn").onclick = () => {
		if (board.currentFen == board.fens.length - 1) return
		board.currentFen += 1
		setCurrentMove()
	}
	getOrPanic("lastMoveBtn").onclick = () => {
		board.currentFen = board.fens.length - 1
		setCurrentMove()
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
		setCurrentMove()
	}
})()
