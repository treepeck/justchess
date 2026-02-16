import {
	Move,
	MoveType,
	PromotionFlag,
	Square,
	highlightCurrentMove,
} from "./move"
import { create, getOrPanic } from "../utils/dom"
import { Piece, PieceType, pieceType2String } from "./piece"

/**
 * @typedef {Object} Position
 * @property {number} x - Horizontal pixel coordinate on the board element.
 * @property {number} y - Vertical pixel coordinate on the board element.
 * @property {Square} square - Index of the square.
 */

/**
 * Function that handles the player's move.
 * @callback MoveHandler
 * @param {number} moveIndex
 * @returns {void}
 */

/** Fen string of the initial position. */
const initPlacement = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR"

/**
 * Map of piece symbols to piece types to parse fen strings.
 * @type {Map<string, PieceType>}
 */
const mapping = new Map()
mapping.set("P", PieceType.WP)
mapping.set("p", PieceType.BP)
mapping.set("N", PieceType.WN)
mapping.set("n", PieceType.BN)
mapping.set("B", PieceType.WB)
mapping.set("b", PieceType.BB)
mapping.set("R", PieceType.WR)
mapping.set("r", PieceType.BR)
mapping.set("Q", PieceType.WQ)
mapping.set("q", PieceType.BQ)
mapping.set("K", PieceType.WK)
mapping.set("k", PieceType.BK)

/**
 * Wrapper around the HTML element that renders and manages the chessboard state.
 */
export default class Board {
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
	 * @param {MoveHandler} moveHandler
	 */
	constructor(moveHandler) {
		this.#element = /** @type {HTMLDivElement} */ (getOrPanic("board"))
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

		// Go through move history using keyboard.
		document.onkeydown = (e) => {
			switch (e.key) {
				case "ArrowUp":
					this.currentFen = this.fens.length - 1
					break
				case "ArrowRight":
					if (this.currentFen == this.fens.length - 1) return
					this.currentFen += 1
					break
				case "ArrowDown":
					this.currentFen = 0
					break
				case "ArrowLeft":
					if (this.currentFen == 0) return
					this.currentFen -= 1
					break
			}
			highlightCurrentMove(this.currentFen)
			this.parsePiecePlacement(this.fens[this.currentFen])
		}
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
				let next = mapping.get(rows[i][j])
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
							move.to
						)
						break
					case PromotionFlag.Bishop:
						this.#appendPiece(
							new Piece(isWhite ? PieceType.WB : PieceType.BB),
							move.to
						)
						break
					case PromotionFlag.Rook:
						this.#appendPiece(
							new Piece(isWhite ? PieceType.WR : PieceType.BR),
							move.to
						)
						break
					case PromotionFlag.Queen:
						this.#appendPiece(
							new Piece(isWhite ? PieceType.WQ : PieceType.BQ),
							move.to
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
			const dp = getOrPanic("draggedPiece")
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
					this.#element.removeChild(getOrPanic("selectedSquare"))
					break
				}
			}

			// Remove dragged piece element from the board.
			const el = getOrPanic("draggedPiece")
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
		const dialog = create("div", "", "promotionDialog")
		dialog.onclick = () => {
			this.#element.removeChild(getOrPanic("promotionDialog"))
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
				}px`
			)
		}

		getOrPanic("board").appendChild(dialog)
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

		const selected = create("div", "", "selectedSquare")

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

		const x = e.clientX - rect.left
		const y = e.clientY - rect.top

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
