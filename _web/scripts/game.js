import Socket from "./ws/socket"
import { g, c } from "./utils/dom"
import Engine from "./utils/engine"
import { EventKind } from "./ws/event"
import showDialog from "./utils/dialog"
import { Board } from "./components/board"
import { Clock, Color } from "./components/clock"
import { formatTermination, formatResult } from "./utils/state"

/**
 * Appends chat message to the DOM.
 * @param {string} msg
 */
function appendChatMessage(msg) {
	const message = c("div", "message")
	message.textContent = msg

	// Append message to chat.
	const container = g("chat")
	container.appendChild(message)

	// Scroll chat to bottom.
	container.scrollTo({
		top: container.scrollHeight,
		behavior: "smooth",
	})
}

/** @param {number} index */
function highlightMove(index) {
	for (const prev of document.getElementsByClassName("move-table-row-san")) {
		prev.classList.remove("current")
	}
	const move = document.getElementById(`${index}`)
	if (move) {
		move.classList.add("current")
	}
}

/**
 * @callback SanClickHandler
 * @param {number} index
 * @returns {void}
 */

/**
 * Appends move SAN to moves table.
 * @param {string} san
 * @param {number} moveIndex
 * @param {SanClickHandler} sanClickHandler
 */
export function appendMove(san, moveIndex, sanClickHandler) {
	// Half move index.
	const ply = Math.ceil(moveIndex / 2)

	// Append row to the table after each black move.
	if (moveIndex % 2 === 1) {
		const row = c("div", "move-table-row", `row${ply}`)
		// Append half-move index to the row.
		const ind = c("div", "move-table-ply")
		ind.textContent = `${ply}.`
		row.appendChild(ind)
		// Append row to the table.
		g("moves").appendChild(row)
	}

	// Append move to the row.
	const move = c("div", "move-table-row-san", `${moveIndex}`)
	highlightMove(moveIndex)
	move.onclick = () => sanClickHandler(moveIndex)
	move.textContent = san
	g(`row${ply}`).appendChild(move)

	// Scroll table to bottom.
	const table = g(`moves`)
	table.scrollTo({
		top: table.scrollHeight,
		behavior: "smooth",
	})
}

;(() => {
	const parts = window.location.pathname.split("/")
	if (parts.length < 3 || (parts[1] != "rated" && parts[1] != "engine"))
		return

	const isTerminated = /** @type {boolean} */ (
		JSON.parse(
			// @ts-expect-error
			document.getElementsByClassName("game-layout")[0].dataset
				.terminated,
		)
	)

	/** @type {import("./components/board").MoveHandler} */
	const moveHandler = (moveIndex) => {
		socket?.sendJSON(EventKind.Move, moveIndex)
	}

	let engine = /** @type {Engine | null} */ (null)
	if (parts[1] == "engine" && !isTerminated) {
		// @ts-expect-error
		const playerColor = parseInt(g("board").dataset.color)
		engine = new Engine(moveHandler, [], 1 ^ playerColor)
	}

	const store = (/** @type {import("./ws/event").PlayedMove} */ move) => {
		const piecePlacement = move.f.split(" ")[0]
		// Update position.
		board.parsePiecePlacement(piecePlacement)
		board.fens.push(piecePlacement)
		board.currentFen = board.fens.length - 1

		appendMove(move.s, board.currentFen, (index) => {
			board.currentFen = index
			highlightMove(index)
			board.parsePiecePlacement(board.fens[index])
		})

		highlightMove(board.currentFen)
	}

	/** @type {import("./ws/socket").EventHandler} */
	const eventHandler = (Kind, payload) => {
		switch (Kind) {
			case EventKind.Chat:
				appendChatMessage(payload)
				break
			case EventKind.Conn:
				appendChatMessage(`${payload} joins the game`)
				break
			case EventKind.Disc:
				appendChatMessage(`${payload} leaves the game`)
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
				if (engine) engine.setLegalMoves(pGame.lm)
				for (const move of pGame.m) {
					store(move)
				}
				if (pGame.wt && pGame.bt && clock) {
					clock.setTime(Color.White, pGame.wt * 1000)
					clock.setTime(Color.Black, pGame.bt * 1000)
					clock.start()
					clock.color =
						board.currentFen % 2 !== 0 ? Color.Black : Color.White
				}

				console.log(engine, engine?.color, board.currentFen % 2)
				if (engine && engine.color == board.currentFen % 2) {
					if (pGame.m.length > 0) {
						// @ts-expect-error
						engine.play(pGame.m.at(-1).f)
					} else {
						engine.play(
							"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
						)
					}
				}
				break
			case EventKind.End:
				clock?.stop()

				/**
				 * Decode payload.
				 * @type {import("./ws/event").EndPayload}
				 */
				const pEnd = { ...payload }
				g("endgameDialogResult").textContent = formatResult(pEnd.r)
				g("endgameDialogTermination").textContent = formatTermination(
					pEnd.t,
				)
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
				if (engine) engine.setLegalMoves(pMove.lm)
				store({ s: pMove.s, f: pMove.f })
				clock?.flip()

				if (engine && engine.color == board.currentFen % 2) {
					engine.play(pMove.f)
				}
				break
			case EventKind.OfferDraw:
				// Display draw offer window.
				showDialog("acceptDrawDialog")
				g("acceptDrawDialogAccept").onclick = () => {
					g("acceptDrawDialog").classList.toggle("show")
					socket?.sendJSON(EventKind.AcceptDraw, null)
				}
				g("acceptDrawDialogClose").onclick = () => {
					g("acceptDrawDialog").classList.toggle("show")
					socket?.sendJSON(EventKind.DeclineDraw, null)
				}
				break
		}
	}

	let socket = /** @type {Socket | null} */ (null)
	if (!isTerminated) {
		socket = new Socket(eventHandler)
	}

	const layout = document.getElementsByClassName("board-layout")[0]
	const board = new Board(moveHandler, layout.classList.contains("flipped"))

	/** @type {{whiteTime: number, blackTime: number}[]} */
	const times = []
	let clock = /** @type {Clock | null} */ (null)
	const control = document.getElementById("whiteClock")
	if (control) {
		const t = parseInt(control.textContent) * 1000
		clock = new Clock(t, false, Color.White, 1000)
		times.push({ whiteTime: t, blackTime: t })
	}

	for (const row of g("moves").getElementsByClassName("move-table-row")) {
		for (const san of row.getElementsByClassName("move-table-row-san")) {
			// @ts-expect-error
			const fen = san.dataset.fen
			// Update position.
			board.parsePiecePlacement(fen)
			board.fens.push(fen)
			board.currentFen = board.fens.length - 1

			const index = parseInt(san.id)
			highlightMove(index)

			// @ts-expect-error
			const timeDiff = san.dataset.timediff
			if (!timeDiff || !clock) {
				continue
			}
			let wt = clock.whiteTime
			let bt = clock.blackTime
			if (index % 2 == 0) {
				wt += parseInt(timeDiff) * 1000
				clock.setTime(Color.White, wt)
			} else {
				bt += parseInt(timeDiff) * 1000
				clock.setTime(Color.Black, bt)
			}
			times.push({
				whiteTime: wt,
				blackTime: bt,
			})
		}
	}
	// Scroll table to bottom.
	const table = g(`moves`)
	table.scrollTo({
		top: table.scrollHeight,
		behavior: "smooth",
	})

	const chat = /** @type {HTMLInputElement | null} */ (
		document.getElementById("chatInput")
	)
	if (chat) {
		const send = () => {
			if (chat.value.length < 1) return
			socket?.sendJSON(EventKind.Chat, chat.value)
			chat.value = ""
		}
		g("chatSend").onclick = send
		chat.onkeydown = (ev) => {
			if (ev.key === "Enter") send()
		}
	}

	if (!isTerminated && chat) {
		g("offerDrawBtn").onclick = () => {
			showDialog("offerDrawDialog")
			g("offerDrawDialogAccept").onclick = () => {
				g("offerDrawDialog").classList.toggle("show")
				socket?.sendJSON(EventKind.OfferDraw, null)
			}
		}

		g("resignBtn").onclick = () => {
			showDialog("resignDialog")
			g("resignDialogAccept").onclick = () => {
				g("resignDialog").classList.toggle("show")
				socket?.sendJSON(EventKind.Resign, null)
			}
		}
	}

	const setCurrentMove = () => {
		highlightMove(board.currentFen - 1)
		board.parsePiecePlacement(board.fens[board.currentFen])

		if (clock) {
			clock.setTime(Color.White, times[board.currentFen].whiteTime)
			clock.setTime(Color.Black, times[board.currentFen].blackTime)
		}
	}

	// Go through move history using buttons.
	g("nullMoveBtn").onclick = () => {
		board.currentFen = 0
		setCurrentMove()
	}
	g("prevMoveBtn").onclick = () => {
		if (board.currentFen == 0) return
		board.currentFen -= 1
		setCurrentMove()
	}
	g("nextMoveBtn").onclick = () => {
		if (board.currentFen == board.fens.length - 1) return
		board.currentFen += 1
		setCurrentMove()
	}
	g("lastMoveBtn").onclick = () => {
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
			default:
				return
		}
		setCurrentMove()
	}

	if (isTerminated) {
		const result = g("endgameDialogResult")
		result.textContent = formatResult(parseInt(result.textContent))

		const term = g("endgameDialogTermination")
		term.textContent = formatTermination(parseInt(term.textContent))
		showDialog("endgameDialog")
	}
})()
