import { Color } from "../components/clock"
import { Move, PromotionFlag } from "../components/board"

// @ts-expect-error
const stockfishUrl = `${API_URL}/stockfish/stockfish-18-lite-single.js`

/**
 * Examples: e2 - 12; b6 - 41.
 * @param {string} square
 * @returns {number}
 */
function square2Number(square) {
	const file = square[0]
	const rank = parseInt(square[1]) - 1

	switch (file) {
		case "a":
			return 8 * rank
		case "b":
			return 8 * rank + 1
		case "c":
			return 8 * rank + 2
		case "d":
			return 8 * rank + 3
		case "e":
			return 8 * rank + 4
		case "f":
			return 8 * rank + 5
		case "g":
			return 8 * rank + 6
		default:
			return 8 * rank + 7
	}
}

/**
 * @param {string} uci
 * @param {Move[]} legalMoves
 * @returns {number}
 */
function UCI2MoveIndex(uci, legalMoves) {
	const from = square2Number(uci.substring(0, 2))
	const to = square2Number(uci.substring(2, 4))
	const promo = PromotionFlag.Queen

	if (uci.includes("N")) {
		promo = PromotionFlag.Knight
	} else if (uci.includes("B")) {
		promo = PromotionFlag.Bishop
	} else if (uci.includes("R")) {
		promo = PromotionFlag.Rook
	}

	for (let i = 0; i < legalMoves.length; i++) {
		const move = legalMoves[i]
		if (from == move.from && to == move.to && promo == move.promoPiece) {
			return i
		}
	}
	throw new Error("Illegal move from engine")
}

export default class Engine {
	/** @type {Worker} */
	worker
	/** @type {import("../components/board").MoveHandler} */
	onMove
	/** @type {Move[]} */
	legalMoves
	/** @type {Color} */
	color

	/**
	 * @param {import("../components/board").MoveHandler} onMove
	 * @param {Move[]} legalMoves
	 * @param {Color} color
	 */
	constructor(onMove, legalMoves, color) {
		this.onMove = onMove
		this.legalMoves = legalMoves
		this.color = color

		this.worker = new Worker(stockfishUrl)

		this.worker.onmessage = (msg) => this.handleMessage(msg)

		// Initialize uci engine.
		this.worker.postMessage("uci")
	}

	/** @param {MessageEvent} msg */
	handleMessage(msg) {
		const tokens = msg.data.split(" ")
		if (tokens.length < 1) {
			console.error("Invalid message")
			return
		}

		switch (tokens[0]) {
			// First of all, configure the this.worker with the specified parameters.
			case "uciok":
				// this.worker.postMessage(
				// 	`setoption name Threads value ${threads}`,
				// )
				// this.worker.postMessage(`setoption name Hash value ${hashSize}`)
				// this.worker.postMessage("setoption name MultiPV value 1")
				// this.worker.postMessage(
				// 	"setoption name UCI_LimitStrength value true",
				// )
				// this.worker.postMessage("setoption name UCI_Elo value 2000")
				this.worker.postMessage("ucinewgame")
				break

			// Finally, process best moves found by engine.
			case "bestmove":
				const uci = tokens[1]
				this.onMove(UCI2MoveIndex(uci, this.legalMoves))
				break
		}
	}

	/** @param {string} fen */
	play(fen) {
		this.worker.postMessage(`position fen ${fen}`)
		setTimeout(() => {
			this.worker.postMessage("go depth 10")
		}, 1500)
	}

	/** @param {Move[]} raw */
	setLegalMoves(raw) {
		this.legalMoves = []
		for (const m of raw) {
			// @ts-expect-error
			this.legalMoves.push(new Move(m))
		}
	}
}
