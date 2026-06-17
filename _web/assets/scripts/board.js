import { get, make } from "/assets/scripts/dom.js"

/**
 * @readonly
 * @enum {number}
 */
const PieceType = {
	WP: 0,
	WN: 1,
	WB: 2,
	WR: 3,
	WQ: 4,
	WK: 5,
	BP: 6,
	BN: 7,
	BB: 8,
	BR: 9,
	BQ: 10,
	BK: 11,
}

/**
 * @param {string} symbol
 * @returns {PieceType}
 * @throws {Error} Throws an error when invalid symbol is found.
 */
function symbol2PieceType(symbol) {
	// prettier-ignore
	switch (symbol) {
		case "P": return PieceType.WP
		case "N": return PieceType.WN
		case "B": return PieceType.WB
		case "R": return PieceType.WR
		case "Q": return PieceType.WQ
		case "K": return PieceType.WK
		case "p": return PieceType.BP
		case "n": return PieceType.BN
		case "b": return PieceType.BB
		case "r": return PieceType.BR
		case "q": return PieceType.BQ
		case "k": return PieceType.BK
	}
	throw new Error("Invalid piece symbol")
}

class Piece {
	/** @type {PieceType} */
	pieceType
	/** @type {HTMLDivElement} */
	element

	/** @param {string} symbol - FEN symbol of piece. */
	constructor(symbol) {
		this.pieceType = symbol2PieceType(symbol)
		this.element = make("div", "board-piece")
		this.element.classList.add(symbol)
	}
}

/**
 * @typedef DraggedPiece
 * @property {Piece} piece
 * @property {number} from - Originating square.
 */

const files = ["a", "b", "c", "d", "e", "f", "g", "h"]
const ranks = ["1", "2", "3", "4", "5", "6", "7", "8"]

export default class Board {
	/** @type {number} In pixels */
	size
	/** @type {number} In pixels */
	squareSize
	/** @type {HTMLDivElement} */
	element
	/**
	 * True - white orientation; false - black orientation.
	 * @type {boolean}
	 */
	orientation
	/**
	 * @type {Map<number, Piece>}
	 */
	pieces
	/**
	 * @type {DraggedPiece | null}
	 */
	draggedPiece

	constructor() {
		this.element = get("board")
		this.size = this.element.clientWidth
		this.squareSize = this.size / 8
		this.draggedPiece = null
		this.pieces = new Map()
		this.orientation = true

		this.element.onpointerdown = (e) => this.onDrag(e)
		this.element.onpointermove = (e) => this.onMove(e)
		this.element.onpointerup = (e) => this.onDrop(e)

		const observer = new ResizeObserver((entries) => {
			for (const entry of entries) {
				// Set new board size and rerender.
				this.size = entry.contentRect.width
				this.squareSize = this.size / 8
				this.render()
			}
		})
		observer.observe(this.element)

		// Render board coords.
		for (let rank = 0; rank < ranks.length; rank++) {
			const r = /** @type {HTMLDivElement} */ (make("div", "board-coord"))
			r.classList.add("rank")
			r.textContent = ranks[rank]
			this.element.appendChild(r)
		}

		for (let file = 0; file < files.length; file++) {
			const f = /** @type {HTMLDivElement} */ (make("div", "board-coord"))
			f.classList.add("file")
			f.textContent = files[file]
			this.element.appendChild(f)
		}
	}

	/** @param {string} piecePlacement */
	parsePiecePlacement(piecePlacement) {
		const rows = piecePlacement.split("/")

		let square = 0
		for (let i = 7; i >= 0; i--) {
			for (let j = 0; j < rows[i].length; j++) {
				const token = parseInt(rows[i][j])
				if (token) {
					// Valid number encountered - skip empty squares.
					square += token
				} else {
					// Piece symbol encountered.
					const p = new Piece(rows[i][j])
					this.appendPiece(square, p)
					square++
				}
			}
		}
	}

	/**
	 * @param {number} square
	 * @param {Piece} p
	 */
	appendPiece(square, p) {
		this.pieces.set(square, p)
		this.element.appendChild(p.element)
		this.render()
	}

	/** @param {PointerEvent} e */
	onDrag(e) {
		const { x, y, square } = this.parseEventCoordinates(e)

		const p = this.pieces.get(square)
		if (!p) return

		// Begin piece drag.
		this.draggedPiece = { piece: p, from: square }

		// Temporary remove the dragged piece from the board.
		this.pieces.delete(square)

		// Position dragged piece.
		p.element.style.setProperty("--x", `${x - this.squareSize / 2}px`)
		p.element.style.setProperty("--y", `${y - this.squareSize / 2}px`)
		p.element.style.zIndex = 2
	}

	/** @param {PointerEvent} e */
	onMove(e) {
		if (!this.draggedPiece) return

		const { x, y, square } = this.parseEventCoordinates(e)

		if (x < 0 || y < 0 || x > this.size || y > this.size) {
			this.appendPiece(this.draggedPiece.from, this.draggedPiece.piece)
			this.draggedPiece.piece.element.style.zIndex = 1
			this.draggedPiece = null
			return
		}

		const element = this.draggedPiece.piece.element
		// Position dragged piece.
		element.style.setProperty("--x", `${x - this.squareSize / 2}px`)
		element.style.setProperty("--y", `${y - this.squareSize / 2}px`)
	}

	/**
	 * Actually handle chess logic.
	 * @param {PointerEvent} e
	 */
	onDrop(e) {
		if (!this.draggedPiece) return

		let { square } = this.parseEventCoordinates(e)
		if (square < 0 || square > 63) {
			square = this.draggedPiece.from
		}

		this.appendPiece(square, this.draggedPiece.piece)
		this.draggedPiece.piece.element.style.zIndex = 1
		this.draggedPiece = null
	}

	/**
	 * @param {PointerEvent} e
	 * @return {{
	 *   x: number,
	 *   y: number,
	 *   rank: number,
	 * 	 file: number,
	 *   square: number,
	 * }}
	 */
	parseEventCoordinates(e) {
		const rect = this.element.getBoundingClientRect()

		const x = e.clientX - rect.left
		const y = e.clientY - rect.top

		let rank = Math.floor((this.size - y) / this.squareSize)
		let file = Math.floor(x / this.squareSize)
		if (!this.orientation) {
			rank = Math.floor(y / this.squareSize)
			file = Math.floor((this.size - x) / this.squareSize)
		}

		return {
			x: x,
			y: y,
			rank: rank,
			file: file,
			square: rank * 8 + file,
		}
	}

	flip() {
		this.orientation = !this.orientation
		this.render()
	}

	render() {
		const offset = this.size - this.squareSize

		// Reposition pieces.
		for (const [square, p] of this.pieces.entries()) {
			const rank = Math.floor(square / 8)
			const file = square % 8

			if (this.orientation) {
				p.element.style.setProperty(
					"--x",
					`${file * this.squareSize}px`,
				)
				p.element.style.setProperty(
					"--y",
					`${offset - rank * this.squareSize}px`,
				)
			} else {
				p.element.style.setProperty(
					"--x",
					`${offset - file * this.squareSize}px`,
				)
				p.element.style.setProperty(
					"--y",
					`${rank * this.squareSize}px`,
				)
			}
		}

		// Reposition board coordinates.
		for (const coord of this.element.getElementsByClassName(
			"board-coord",
		)) {
			const fontSize = this.size * 0.03
			coord.style.fontSize = `${fontSize}px`

			if (coord.classList.contains("rank")) {
				const rank = parseInt(coord.textContent) - 1
				coord.style.setProperty("--x", `${0}px`)

				const y = this.orientation
					? offset - rank * this.squareSize
					: rank * this.squareSize
				coord.style.setProperty("--y", `${y}px`)
			} else {
				let file = 0
				for (let i = 1; i < files.length; i++) {
					if (files[i] == coord.textContent) {
						file = i
						break
					}
				}

				const x = this.orientation
					? file * this.squareSize
					: this.size - (file + 1) * this.squareSize

				coord.style.setProperty("--x", `${x}px`)
				coord.style.setProperty(
					"--y",
					`${this.size - this.size * 0.045}px`,
				)
			}
		}
	}
}
