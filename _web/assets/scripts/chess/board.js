import { get, make } from "/assets/scripts/dom.js"

/** @enum {number} */
// prettier-ignore
export const Square = /** @type {const} */ ({
	A8: 56, B8: 57, C8: 58, D8: 59, E8: 60, F8: 61, G8: 62, H8: 63,
	A7: 48, B7: 49, C7: 50, D7: 51, E7: 52, F7: 53, G7: 54, H7: 55,
	A6: 40, B6: 41, C6: 42, D6: 43, E6: 44, F6: 45, G6: 46, H6: 47,
	A5: 32, B5: 33, C5: 34, D5: 35, E5: 36, F5: 37, G5: 38, H5: 39,
	A4: 24, B4: 25, C4: 26, D4: 27, E4: 28, F4: 29, G4: 30, H4: 31,
	A3: 16, B3: 17, C3: 18, D3: 19, E3: 20, F3: 21, G3: 22, H3: 23,
	A2:  8, B2:  9, C2: 10, D2: 11, E2: 12, F2: 13, G2: 14, H2: 15,
	A1:  0, B1:  1, C1:  2, D1:  3, E1:  4, F1:  5, G1:  6, H1:  7,
})

/**
 * Enum representing chess pieces.
 * @enum {number}
 */
export const PieceType = /** @type {const} */ ({
	WP: 0,
	BP: 1,
	WN: 2,
	BN: 3,
	WB: 4,
	BB: 5,
	WR: 6,
	BR: 7,
	WQ: 8,
	BQ: 9,
	WK: 10,
	BK: 11,
	NP: -1,
})

const pieceType2fen = /** @type {Map<PieceType, string>} */ (new Map())
pieceType2fen.set(PieceType.WP, "P")
pieceType2fen.set(PieceType.BP, "p")
pieceType2fen.set(PieceType.WN, "N")
pieceType2fen.set(PieceType.BN, "n")
pieceType2fen.set(PieceType.WB, "B")
pieceType2fen.set(PieceType.BB, "b")
pieceType2fen.set(PieceType.WR, "R")
pieceType2fen.set(PieceType.BR, "r")
pieceType2fen.set(PieceType.WQ, "Q")
pieceType2fen.set(PieceType.BQ, "q")
pieceType2fen.set(PieceType.WK, "K")
pieceType2fen.set(PieceType.BK, "k")

const fen2pieceType = /** @type {Map<string, PieceType>} */ (new Map())
fen2pieceType.set("P", PieceType.WP)
fen2pieceType.set("p", PieceType.BP)
fen2pieceType.set("N", PieceType.WN)
fen2pieceType.set("n", PieceType.BN)
fen2pieceType.set("B", PieceType.WB)
fen2pieceType.set("b", PieceType.BB)
fen2pieceType.set("R", PieceType.WR)
fen2pieceType.set("r", PieceType.BR)
fen2pieceType.set("Q", PieceType.WQ)
fen2pieceType.set("q", PieceType.BQ)
fen2pieceType.set("K", PieceType.WK)
fen2pieceType.set("k", PieceType.BK)

export class Piece {
	/**
	 * Reference to the element in which the piece will be rendered.
	 * @type {HTMLDivElement}
	 */
	element
	/**
	 * @type {PieceType}
	 */
	pieceType

	/**
	 * @param {PieceType} pieceType
	 */
	constructor(pieceType) {
		this.element = /** @type {HTMLDivElement} */ (
			make("div", "board-piece")
		)
		this.element.classList.add(pieceType2fen.get(pieceType))
		this.pieceType = pieceType
	}
}

/** Fen string of the initial position. */
const initPlacement = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR"

/**
 * @typedef Position
 * @property {number} x
 * @property {number} y
 * @property {Square} square
 */

/**
 * Wrapper around the HTMLDivElement that renders and manages the chessboard state.
 */
export class Board {
	/**
	 * Reference to the element in which the board will be rendered.
	 * @type {HTMLDivElement}
	 */
	#element
	/**
	 * @type {Map<Square, Piece>}
	 */
	#pieces
	/**
	 * @type {Piece}
	 */
	#draggedPiece
	/** @type {Square} */
	#draggedPieceOrigin
	/**
	 * Size of the board square in pixels.
	 * @type {number}
	 */
	#squareSize

	constructor() {
		this.#element = /** @type {HTMLDivElement} */ (get("board"))
		this.#squareSize = 62.5

		const observer = new ResizeObserver((entries) => {
			for (const entry of entries) {
				this.#squareSize = entry.contentRect.width / 8
				this.render()
			}
		})
		observer.observe(this.#element)

		this.#element.onpointerdown = (e) => this.#pieceDragStart(e)
		this.#element.onpointermove = (e) => this.#pieceDragMove(e)
		this.#element.onpointerup = (e) => this.#pieceDrop(e)

		this.#pieces = new Map()
		this.parsePiecePlacement(initPlacement)
	}

	/**
	 * Renders the current board state.
	 */
	render() {
		this.clear()

		for (let rank = 0; rank < 8; rank++) {
			for (let file = 0; file < 8; file++) {
				// Render the chess squares.
				let color = "white"
				const sqInd = rank * 8 + file

				if (rank % 2 !== 0) {
					if (sqInd % 2 !== 0) {
						color = "black"
					}
				} else {
					if (sqInd % 2 === 0) {
						color = "black"
					}
				}

				const square = /** @type {HTMLDivElement} */ (make("div"))
				square.id = sqInd
				square.style.position = "absolute"
				square.style.width = `${this.#squareSize}px`
				square.style.height = `${this.#squareSize}px`
				square.style.left = `${this.#squareSize * file}px`
				square.style.bottom = `${this.#squareSize * rank}px`
				square.style.backgroundColor = color

				this.#element.appendChild(square)

				// Render all pieces.
				const p = this.#pieces.get(sqInd)
				if (p) {
					const boardSize = this.#squareSize * 7

					p.element.style.setProperty(
						"--x",
						`${boardSize - file * this.#squareSize}px`,
					)
					p.element.style.setProperty(
						"--y",
						`${boardSize - rank * this.#squareSize}px`,
					)

					this.#element.appendChild(p.element)
				}

				// Render dragged piece.
				if (this.#draggedPiece) {
					this.#element.appendChild(this.#draggedPiece.element)
				}
			}
		}
	}

	/**
	 * Parses the Fen string and updates the position state.
	 * @param {string} fen
	 */
	parsePiecePlacement(fen) {
		// Remove previous pieces from board.
		this.#pieces.clear()

		const rows = fen.split("/")
		let sqInd = 0
		for (let i = 7; i >= 0; i--) {
			for (let j = 0; j < rows[i].length; j++) {
				let next = fen2pieceType.get(rows[i][j])
				if (next !== undefined) {
					const p = new Piece(next)
					this.#pieces.set(sqInd, p)
					sqInd++
				} else {
					// Skip empty squares.
					sqInd += parseInt(rows[i][j])
				}
			}
		}
	}

	/** @param {PointerEvent} e */
	#pieceDragStart(e) {
		const { x, y, square } = this.eventPosition(e)

		console.log(square, x, y)

		const p = this.#pieces.get(square)
		if (!p) {
			return
		}
		// Begin piece drag.
		this.#draggedPieceOrigin = square
		this.#draggedPiece = p

		// Remove piece from board while it's being dragged.
		this.#pieces.delete(square)

		this.render()
	}

	/**
	 * Updates dragged piece position.
	 * @param {PointerEvent} e
	 */
	#pieceDragMove(e) {
		if (!this.#draggedPiece) return

		let { x, y, square } = this.eventPosition(e)
		const boardSize = this.#squareSize * 8

		// Center the piece to the cursor position.
		// x -= this.#squareSize / 2
		// y -= this.#squareSize / 2

		this.#draggedPiece.element.style.setProperty(
			"--x",
			`${boardSize - (square % 8) * this.#squareSize}px`,
		)
		p.element.style.setProperty(
			"--y",
			`${boardSize - Math.floor(rank / 8) * this.#squareSize}px`,
		)

		this.render()
	}

	/** @param {PointerEvent} e */
	#pieceDrop(e) {
		// TODO(artem): actual chess logic, send move to the server.
		const { x, y, square } = this.eventPosition(e)

		this.#pieces.set(square, this.#draggedPiece)

		this.#draggedPiece = null
		this.#draggedPieceOrigin = -1

		this.render()
	}

	/**
	 * Deletes all pieces and elements from board.
	 */
	clear() {
		this.#element.textContent = ""
	}

	/**
	 * @param {PointerEvent} e
	 * @returns {Position}
	 */
	eventPosition(e) {
		const rect = this.#element.getBoundingClientRect()

		const boardSize = this.#squareSize * 8

		let x = e.clientX - rect.left
		let y = boardSize - (e.clientY - rect.top)

		const file = Math.floor(x / this.#squareSize)
		const rank = Math.floor(y / this.#squareSize)

		return {
			x: x,
			y: y,
			square: 8 * rank + file,
		}
	}
}

const board = new Board()
