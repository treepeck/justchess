import { g, c } from "../utils/dom"

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

/** @enum {number} */
export const MoveType = /** @type {const} */ ({
	Normal: 0,
	Castling: 1,
	Promotion: 2,
	EnPassant: 3,
})

/** @enum {number} */
export const PromotionFlag = /** @type {const} */ ({
	Knight: 0,
	Bishop: 1,
	Rook: 2,
	Queen: 3,
})

export class Move {
	/**
	 * Index of the destination square.
	 * @type {number}
	 */
	to
	/**
	 *  Index of the origin square.
	 * @type {number}
	 */
	from
	/**
	 * @type {PromotionFlag}
	 */
	promoPiece
	/**
	 *
	 * @type {MoveType}
	 */
	moveType

	/**
	 * Decodes the given integer into the move.
	 * @param {number} raw
	 */
	constructor(raw) {
		this.to = raw & 0x3f
		this.from = (raw >> 6) & 0x3f
		this.promoPiece = (raw >> 12) & 0x3
		this.moveType = (raw >> 14) & 0x3
	}
}

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

/**
 * @param {PieceType} pieceType
 * @returns {string}
 */
function pieceType2String(pieceType) {
	switch (pieceType) {
		case PieceType.WP:
			return "P"
		case PieceType.BP:
			return "p"
		case PieceType.WN:
			return "N"
		case PieceType.BN:
			return "n"
		case PieceType.WB:
			return "B"
		case PieceType.BB:
			return "b"
		case PieceType.WR:
			return "R"
		case PieceType.BR:
			return "r"
		case PieceType.WQ:
			return "Q"
		case PieceType.BQ:
			return "q"
		case PieceType.WK:
			return "K"
		default:
			return "k"
	}
}

/**
 * @param {string} str
 * @returns {PieceType | undefined}
 */
function string2PieceType(str) {
	switch (str) {
		case "P":
			return PieceType.WP
		case "p":
			return PieceType.BP
		case "N":
			return PieceType.WN
		case "n":
			return PieceType.BN
		case "B":
			return PieceType.WB
		case "b":
			return PieceType.BB
		case "R":
			return PieceType.WR
		case "r":
			return PieceType.BR
		case "Q":
			return PieceType.WQ
		case "q":
			return PieceType.BQ
		case "K":
			return PieceType.WK
		case "k":
			return PieceType.BK
	}
}

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
		this.element = /** @type {HTMLDivElement} */ (c("div", "piece"))
		this.element.classList.add(`${pieceType2String(pieceType)}`)
		this.pieceType = pieceType
	}
}

/**
 * @typedef {Object} Position
 * @property {number} x - Horizontal pixel coordinate on the board element.
 * @property {number} y - Vertical pixel coordinate on the board element.
 * @property {Square} square - Index of the square.
 */

/**
 * Function that handles player's moves.
 * @callback MoveHandler
 * @param {number} moveIndex
 * @returns {void}
 */

/** Fen string of the initial position. */
const initPlacement = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR"

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
	 * Piece placement.
	 * @type {Map<Square, Piece>}
	 */
	#pieces
	/**
	 * @type {Move[]}
	 */
	#legalMoves
	/**
	 * @type {PieceType}
	 */
	#draggedPiece
	/**
	 * @type {Square}
	 */
	#selectedSquare
	/**
	 * Fen strings of reached positions during the game.
	 * @type {string[]}
	 */
	fens
	/**
	 * Index of the current displayed position.
	 * @type {number}
	 */
	currentFen
	/**
	 * Board element size in pixels.
	 * @type {number}
	 */
	#size
	/**
	 * @type {MoveHandler}
	 */
	#moveHandler
	/**
	 * @type {boolean}
	 */
	#isFlipped

	/**
	 * @param {MoveHandler} moveHandler
	 * @param {boolean} isFlipped
	 */
	constructor(moveHandler, isFlipped) {
		this.#element = /** @type {HTMLDivElement} */ (g("board"))
		this.#pieces = /** @type {Map<Square, Piece>} */ (new Map())

		this.parsePiecePlacement(initPlacement)

		this.#selectedSquare = -1
		this.#draggedPiece = PieceType.NP
		this.#legalMoves = []
		this.#moveHandler = moveHandler

		this.fens = [initPlacement]
		this.currentFen = 0

		// Initialize default board size.
		this.#size = this.#element.offsetWidth

		// Add board event listeners.
		this.#element.onmousedown = (e) => this.#onMouseDown(e)
		this.#element.onmousemove = (e) => this.#onMouseMove(e)
		this.#element.onmouseup = (e) => this.#onMouseUp(e)

		// Make board responsive.
		const observer = new ResizeObserver((entries) => {
			for (const entry of entries) {
				this.#setSize(entry.contentRect.width)
			}
		})
		observer.observe(this.#element)

		g("flipBoardBtn").onclick = () => {
			const layout = document.getElementsByClassName("board-layout")[0]
			layout.classList.toggle("flipped")
			this.#isFlipped = !this.#isFlipped
		}

		this.#isFlipped = isFlipped
	}

	/**
	 * Parses the Fen string and updates the position state.
	 * @param {string} fen
	 */
	parsePiecePlacement(fen) {
		// Remove previous pieces from board.
		this.#pieces.clear()
		while (this.#element.firstChild) {
			this.#element.removeChild(this.#element.firstChild)
		}

		const rows = fen.split("/")
		let sqInd = 0
		for (let i = 7; i >= 0; i--) {
			for (let j = 0; j < rows[i].length; j++) {
				let next = string2PieceType(rows[i][j])
				if (next !== undefined) {
					this.#appendPiece(new Piece(next), sqInd)
					sqInd++
				} else {
					// Skip empty squares.
					sqInd += parseInt(rows[i][j])
				}
			}
		}
	}

	/**
	 * Serializes the current position into a Fen string.
	 * @returns {string}
	 */
	serializePiecePlacement() {
		let fen = ""
		// Convert map to array for more convinient serializing.
		const board = /** @type {PieceType[]} */ ([])
		// Go through all squares.
		for (let i = 0; i < 64; i++) {
			const piece = this.#pieces.get(i)
			// Add piece to board.
			board[i] = piece !== undefined ? piece.pieceType : PieceType.NP
		}

		let emptySquares = 0
		for (let rank = 7; rank >= 0; rank--) {
			for (let file = 0; file < 8; file++) {
				const square = 8 * rank + file

				if (board[square] === PieceType.NP) {
					emptySquares++
				} else {
					if (emptySquares > 0) {
						fen += emptySquares.toString()
						emptySquares = 0
					}
					fen += pieceType2String(board[square])
				}

				// Add rank separators.
				if ((square + 1) % 8 === 0) {
					if (emptySquares > 0) {
						fen += emptySquares.toString()
						emptySquares = 0
					}
					// Don't add separator in the end of string.
					if (square !== 7) {
						fen += "/"
					}
				}
			}
		}
		return fen
	}

	/**
	 * Does not affect reachedFes or currentFen. Simply performs the move.
	 * It's the caller's responsibility to ensure that move is legal.
	 * @param {Move} move
	 */
	makeMove(move) {
		const piece = this.#pieces.get(move.from)
		if (piece === undefined) throw new Error("Piece is not on the board")

		// Remove captured piece if it is present.
		const captured = this.#pieces.get(move.to)
		if (captured !== undefined) {
			this.#pieces.delete(move.to)
			this.#element.removeChild(captured.element)
		}

		switch (move.moveType) {
			case MoveType.Castling:
				// Update rook position.
				switch (move.to) {
					case Square.G1: // White O-O.
						this.#movePiece(Square.H1, Square.F1)
						break

					case Square.C1: // White O-O-O.
						this.#movePiece(Square.A1, Square.D1)
						break

					case Square.G8: // Black O-O.
						this.#movePiece(Square.H8, Square.F8)
						break

					case Square.C8: // Black O-O-O.
						this.#movePiece(Square.A8, Square.D8)
						break
				}
				break

			case MoveType.EnPassant:
				const capturedSquare =
					piece.pieceType === PieceType.WP ? move.to - 8 : move.to + 8
				const captured = this.#pieces.get(capturedSquare)
				if (captured === undefined)
					throw new Error("Piece is not on board")

				this.#pieces.delete(capturedSquare)
				this.#element.removeChild(captured.element)
				break

			case MoveType.Promotion:
				// Remove pawn.
				this.#pieces.delete(move.from)
				this.#element.removeChild(piece.element)

				const isWhite = piece.pieceType % 2 === 0

				// Place promoted piece.
				switch (move.promoPiece) {
					case PromotionFlag.Knight:
						this.#appendPiece(
							new Piece(isWhite ? PieceType.WN : PieceType.BN),
							move.to,
						)
						break
					case PromotionFlag.Bishop:
						this.#appendPiece(
							new Piece(isWhite ? PieceType.WB : PieceType.BB),
							move.to,
						)
						break
					case PromotionFlag.Rook:
						this.#appendPiece(
							new Piece(isWhite ? PieceType.WR : PieceType.BR),
							move.to,
						)
						break
					case PromotionFlag.Queen:
						this.#appendPiece(
							new Piece(isWhite ? PieceType.WQ : PieceType.BQ),
							move.to,
						)
						break
				}
				return
		}

		// Move piece to the destination square.
		this.#movePiece(move.from, move.to)
	}

	/**
	 * Updates legal moves on the board.
	 * @param {Move[]} moves
	 */
	setLegalMoves(moves) {
		this.#legalMoves = []
		for (const encoded of moves) {
			// @ts-expect-error
			this.#legalMoves.push(new Move(encoded))
		}
	}

	/**
	 * Handles player's click on the board element.
	 * @param {MouseEvent} e
	 */
	#onMouseDown(e) {
		const { x, y, square } = this.#getPositionOfEvent(e)

		const prev = document.getElementById("selectedSquare")
		if (prev !== null) {
			const from = this.#selectedSquare

			// Remove previous selected square.
			this.#element.removeChild(prev)
			this.#selectedSquare = -1

			// Call move handler is the move is legal.
			for (let i = 0; i < this.#legalMoves.length; i++) {
				const move = this.#legalMoves[i]
				if (move.from === from && move.to === square) {
					const piece = /** @type {Piece} */ (this.#pieces.get(from))
					if (move.moveType === MoveType.Promotion) {
						const isWhite = piece.pieceType % 2 === 0
						this.#renderPromotionDialog(isWhite, move.to, i)
					} else {
						this.makeMove(move)
						this.#moveHandler(i)
					}
					return
				}
			}
		}

		const piece = this.#pieces.get(square)
		if (piece !== undefined) {
			// Remove piece from board while it's being dragged.
			this.#pieces.delete(square)
			this.#element.removeChild(piece.element)

			// Append dragged piece to the board.
			const dp = new Piece(piece.pieceType)
			dp.element.id = "draggedPiece"
			this.#element.appendChild(dp.element)

			// Position dragged piece under the player's mouse cursor.
			const squareSize = this.#size / 8
			dp.element.style.setProperty("--x", `${x - squareSize / 2}px`)
			dp.element.style.setProperty("--y", `${y - squareSize / 2}px`)

			this.#draggedPiece = piece.pieceType
		}

		// Append new selected square to the board.
		this.#appendSelectedSquare(square)
	}

	/**
	 * Handles player's mouse movements above the board element.
	 * @param {MouseEvent} e
	 */
	#onMouseMove(e) {
		if (this.#draggedPiece !== PieceType.NP) {
			const { x, y } = this.#getPositionOfEvent(e)

			// Position dragged piece under the player's mouse cursor.
			const dp = g("draggedPiece")
			const squareSize = this.#size / 8
			dp.style.setProperty("--x", `${x - squareSize / 2}px`)
			dp.style.setProperty("--y", `${y - squareSize / 2}px`)
		}
	}

	/**
	 * Handles player's mouse release on the board element.
	 * @param {MouseEvent} e
	 */
	#onMouseUp(e) {
		const dp = this.#draggedPiece

		if (dp !== PieceType.NP) {
			// Restore piece placement.
			this.#appendPiece(new Piece(dp), this.#selectedSquare)

			// Call move handler if the move is valid.
			const { square } = this.#getPositionOfEvent(e)
			for (let i = 0; i < this.#legalMoves.length; i++) {
				const move = this.#legalMoves[i]
				if (move.from === this.#selectedSquare && move.to === square) {
					if (move.moveType === MoveType.Promotion) {
						const isWhite = dp % 2 === 0
						this.#renderPromotionDialog(isWhite, move.to, i)
					} else {
						this.makeMove(move)
						this.#moveHandler(i)
					}
					// Reset selected square.
					this.#selectedSquare = -1
					this.#element.removeChild(g("selectedSquare"))
					break
				}
			}

			// Remove dragged piece element from the board.
			const el = g("draggedPiece")
			this.#element.removeChild(el)
			this.#draggedPiece = PieceType.NP
		}
	}

	/**
	 * Update piece positions when the elemen's size changes to make board responsive.
	 * @param {number} size
	 */
	#setSize(size) {
		this.#size = Math.round(size)
		this.#pieces.forEach((p, s) => this.#appendPiece(p, s))
	}

	/**
	 * @param {Square} from
	 * @param {Square} to
	 */
	#movePiece(from, to) {
		const piece = this.#pieces.get(from)
		if (piece === undefined) throw new Error("Piece is not on the board")

		// Update piece position.
		this.#pieces.delete(from)
		this.#pieces.set(to, piece)

		const { x, y } = this.#square2Position(to)
		piece.element.style.setProperty("--x", `${x}px`)
		piece.element.style.setProperty("--y", `${y}px`)
	}

	/**
	 * @param {boolean} isWhite
	 * @param {Square} destination
	 * @param {number} moveIndex
	 */
	#renderPromotionDialog(isWhite, destination, moveIndex) {
		const dialog = c("div", "", "promotionDialog")
		dialog.onclick = () => {
			this.#element.removeChild(g("promotionDialog"))
		}

		const promoPieces = [
			new Piece(isWhite ? PieceType.WN : PieceType.BN),
			new Piece(isWhite ? PieceType.WB : PieceType.BB),
			new Piece(isWhite ? PieceType.WR : PieceType.BR),
			new Piece(isWhite ? PieceType.WQ : PieceType.BQ),
		]
		for (let i = 0; i < promoPieces.length; i++) {
			const piece = promoPieces[i]
			piece.element.classList.add("promotion-choice")

			// Add event listener.
			piece.element.onclick = () => {
				this.makeMove(this.#legalMoves[moveIndex + i])
				this.#moveHandler(moveIndex + i)
			}

			dialog.appendChild(piece.element)

			const squareSize = this.#size / 8
			const file = Math.floor(destination % 8)
			const rank = Math.floor(destination / 8)

			piece.element.style.setProperty("--x", `${file * squareSize}px`)
			piece.element.style.setProperty(
				"--y",
				`${
					this.#size -
					squareSize -
					(isWhite ? rank - i : rank + i) * squareSize
				}px`,
			)
		}

		g("board").appendChild(dialog)
	}

	/**
	 * @param {Piece} piece
	 * @param {Square} square
	 */
	#appendPiece(piece, square) {
		if (this.#pieces.get(square) === undefined) {
			this.#pieces.set(square, piece)
		}

		const pos = this.#square2Position(square)
		piece.element.style.setProperty("--x", `${pos.x}px`)
		piece.element.style.setProperty("--y", `${pos.y}px`)

		this.#element.appendChild(piece.element)
	}

	/**
	 * Appends the selected square element to the board.
	 * @param {Square} square
	 */
	#appendSelectedSquare(square) {
		this.#selectedSquare = square

		const selected = c("div", "", "selectedSquare")

		const pos = this.#square2Position(square)
		selected.style.setProperty("--x", `${pos.x}px`)
		selected.style.setProperty("--y", `${pos.y}px`)

		this.#element.appendChild(selected)
	}

	/**
	 * Returns the position of the triggered mouse event.
	 * @param {MouseEvent} e
	 * @returns {Position}
	 */
	#getPositionOfEvent(e) {
		const squareSize = this.#size / 8

		const rect = this.#element.getBoundingClientRect()

		let x = e.clientX - rect.left
		let y = e.clientY - rect.top

		if (this.#isFlipped) {
			x = rect.right - e.clientX
			y = rect.bottom - e.clientY
		}

		const file = Math.floor(x / squareSize)
		const rank = Math.floor((this.#size - y) / squareSize)

		return {
			x: x,
			y: y,
			square: 8 * rank + file,
		}
	}

	/**
	 * @param {Square} square
	 * @returns {Position}
	 */
	#square2Position(square) {
		const squareSize = this.#size / 8

		const file = Math.floor(square % 8)
		const rank = Math.floor(square / 8)

		return {
			x: file * squareSize,
			y: this.#size - squareSize - rank * squareSize,
			square: square,
		}
	}
}
