import Notification from "./utils/notification"
import { EventAction } from "./ws/event"
import BoardCanvas from "./chess/board"
import { Move } from "./chess/move"

/** Manages the state of the active game. */
class Game {
	/**
	 * @type {Move[]}
	 */
	legalMoves
	/**
	 * @type {number}
	 */
	whiteTime
	/**
	 * @type {number}
	 */
	blackTime
	/**
	 * @type {boolean}
	 */
	isWhiteConnected
	/**
	 * @type {boolean}
	 */
	isBlackConnected
	/**
	 * @type {BoardCanvas}
	 */
	board

	/**
	 * @param {import("./ws/event").GamePayload} payload
	 * @param {BoardCanvas} board
	 */
	constructor(payload, board) {
		this.#setLegalMoves(payload.lm)

		for (const completedMove of payload.m) {
			this.#appendMove(completedMove)
		}

		this.whiteTime = payload.wt
		this.blackTime = payload.bt
		this.isWhiteConnected = payload.w
		this.isBlackConnected = payload.b

		this.board = board
	}

	/** @param {import("./ws/event").MovePayload} payload */
	handleMove(payload) {
		this.#setLegalMoves(payload.lm)

		this.#appendMove(payload.m)

		// Update player's clock.
		if (this.legalMoves.length % 2 == 0) {
			this.whiteTime = payload.m.t
		} else {
			this.blackTime = payload.m.t
		}

		this.board.makeMove(payload.m.m)
	}

	/** @param {Move[]} moves */
	#setLegalMoves(moves) {
		this.legalMoves = []
		for (const encoded of moves) {
			this.legalMoves.push(new Move(Number(encoded)))
		}
	}

	/** @param {import("./chess/move").CompletedMove} move */
	#appendMove(move) {
		const table = document.getElementById("tableBody")
		if (!table) return

		const moveDiv = document.createElement("div")
		moveDiv.classList.add("move")
		moveDiv.textContent = move.s
		table.appendChild(moveDiv)
	}
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

	/** @type {HTMLImageElement | null} */
	let sheet = null
	try {
		sheet = await loadSheet()
	} catch (err) {
		console.error(err)
	}
	if (!sheet) return

	const canvas = document.getElementById("boardCanvas")
	if (!canvas || !(canvas instanceof HTMLCanvasElement)) return

	const ctx = canvas.getContext("2d")
	if (!ctx) return

	const board = new BoardCanvas(ctx, sheet, null)
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

	/** @type {Game | null} */
	let game = null

	// @ts-expect-error - API_URL comes from webpack.
	const socket = new WebSocket(`${WS_URL}/ws?id=${id}`)

	const notification = new Notification()
	socket.onclose = () => {
		if (game) {
			notification.create("Please reload the page to reconnect.")
		}
	}

	/** @type {import("./chess/board").MoveCallback} */
	function moveHandler(from, to) {
		if (!game) return false

		for (let i = 0; i < game.legalMoves.length; i++) {
			const move = game.legalMoves[i]
			if (move.from == from && move.to == to) {
				socket.send(JSON.stringify({ a: EventAction.Move, p: i }))
				return true
			}
		}
		return false
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

			case EventAction.Game:
				game = new Game(payload, board)
				board.moveHandler = moveHandler
				break

			case EventAction.Move:
				if (!game) return
				game.handleMove(payload)
				break
		}
	}
}

main()
