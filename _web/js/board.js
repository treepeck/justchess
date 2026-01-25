/**
 * Enum representing chess pieces mapped to integer values.
 * @typedef {number} Piece
 */
const Piece = Object.freeze({
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
	NP: -1, // No piece.
})

export default class Board {
	/**
	 * Context to which the board will be rendered.
	 * @type {CanvasRenderingContext2D}
	 */
	#context
	/**
	 * Canvas size in pixels.
	 * @type {number}
	 */
	#size
	/**
	 * Index of the selected square. -1 means that no square is selected.
	 * @type {number}
	 */
	#selectedSquare
	/**
	 * Spritesheet.
	 * @type {Image}
	 */
	#sheet
	/**
	 * Piece which is currently being dragged. Null is case no piece is being
	 * dragged.
	 * @typedef {Object} DraggedPiece
	 * @property {number} x - The horizontal pixel coordinate on the canvas.
	 * @property {number} y - The vertical pixel coordinate on the canvas.
	 * @property {number} from - Source position of the dragged piece (originating square).
	 * @property {Piece} piece - Piece type.
	 */
	/**
	 * @type {DraggedPiece}
	 */
	#draggedPiece
	/**
	 * Piece placement.
	 * @type {Piece[]}
	 */
	#squares
	/**
	 * Size of an individual square on the canvas in pixels.
	 * @type {number}
	 */
	#square
	/**
	 * Size of an individual piece sprite on the spritesheet.
	 * @constant
	 * @type {number}
	 */
	#piece

	/**
	 * @param {CanvasRenderingContext2D} ctx
	 * @param {Image} sheet
	 */
	constructor(ctx, sheet) {
		this.#context = ctx
		this.#sheet = sheet
		this.#size = 0
		this.#selectedSquare = -1
		this.#draggedPiece = null

		this.#square = this.#size / 8
		this.#piece = 300

		// Initialize default piece placement.
		// prettier-ignore
		this.#squares = [
			Piece.WR, Piece.WN, Piece.WB, Piece.WQ, Piece.WK, Piece.WB, Piece.WN, Piece.WR,
			Piece.WP, Piece.WP, Piece.WP, Piece.WP, Piece.WP, Piece.WP, Piece.WP, Piece.WP,
			Piece.NP, Piece.NP, Piece.NP, Piece.NP, Piece.NP, Piece.NP, Piece.NP, Piece.NP,
			Piece.NP, Piece.NP, Piece.NP, Piece.NP, Piece.NP, Piece.NP, Piece.NP, Piece.NP,
			Piece.NP, Piece.NP, Piece.NP, Piece.NP, Piece.NP, Piece.NP, Piece.NP, Piece.NP,
			Piece.NP, Piece.NP, Piece.NP, Piece.NP, Piece.NP, Piece.NP, Piece.NP, Piece.NP,
			Piece.BP, Piece.BP, Piece.BP, Piece.BP, Piece.BP, Piece.BP, Piece.BP, Piece.BP,
			Piece.BR, Piece.BN, Piece.BB, Piece.BQ, Piece.BK, Piece.BB, Piece.BN, Piece.BR,
		]
	}

	/** Renders the board. */
	#render() {
		for (let rank = 0; rank < 8; rank++) {
			for (let file = 0; file < 8; file++) {
				// Draw board squares.
				this.#context.fillStyle = "#e2d3c4"
				if (
					(rank % 2 !== 0 && file % 2 !== 0) ||
					(rank % 2 === 0 && file % 2 === 0)
				) {
					this.#context.fillStyle = "#8e684b"
				}
				const x = file * this.#square
				const y = this.#size - this.#square - rank * this.#square
				this.#context.fillRect(x, y, this.#square, this.#square)

				const ind = 8 * rank + file
				// Draw selected this.#square.
				if (ind === this.#selectedSquare) {
					this.#context.fillStyle = "green"
					this.#context.fillRect(x, y, this.#square, this.#square)
				}

				// Draw dragged piece.
				if (this.#draggedPiece !== null) {
					this.#context.drawImage(
						this.#sheet,
						Math.floor(this.#draggedPiece.piece / 2) * this.#piece,
						this.#draggedPiece.piece % 2 === 0 ? 0 : this.#piece,
						this.#piece,
						this.#piece,
						this.#draggedPiece.x,
						this.#draggedPiece.y,
						this.#square,
						this.#square
					)
				}

				// Draw pieces.
				if (this.#squares[ind] !== Piece.NP) {
					this.#context.drawImage(
						this.#sheet,
						Math.floor(this.#squares[ind] / 2) * this.#piece,
						this.#squares[ind] % 2 === 0 ? 0 : this.#piece,
						this.#piece,
						this.#piece,
						x,
						y,
						this.#square,
						this.#square
					)
				}
			}
		}
	}

	/**
	 * Rerender the board on a canvas when the size changes.
	 * @param {number} size
	 */
	setSize(size) {
		this.#size = size
		this.#square = this.#size / 8

		// Resize canvas.
		this.#context.canvas.width = this.#size
		this.#context.canvas.height = this.#size

		this.#render()
	}

	/**
	 * Handles player's clicks on the canvas.
	 * @param {MouseEvent} e
	 */
	onMouseDown(e) {
		const rect = this.#context.canvas.getBoundingClientRect()
		const x = e.clientX - rect.left
		const y = e.clientY - rect.top

		const file = Math.floor(x / this.#square)
		const rank = Math.floor((this.#size - y) / this.#square)

		this.#selectedSquare = 8 * rank + file

		// Begin piece drag.  Temporary remove the piece from its square while
		// its being dragged.
		if (this.#squares[this.#selectedSquare] !== Piece.NP) {
			const piece = this.#squares[this.#selectedSquare]
			// Update board state.
			this.#squares[this.#selectedSquare] = Piece.NP
			this.#draggedPiece = {
				x: x - this.#square / 2, // Center piece horizontally.
				y: y - this.#square / 2, // Center piece vertically.
				from: this.#selectedSquare,
				piece: piece,
			}
		}

		this.#render()
	}

	/**
	 * Handles player's mouse movement on the canvas.
	 * @param {MouseEvent} e
	 */
	onMouseMove(e) {
		if (this.#draggedPiece !== null) {
			const isLeftButtonPressed = e.buttons === 1

			if (isLeftButtonPressed) {
				const rect = this.#context.canvas.getBoundingClientRect()
				// Move dragged piece with the cursor.
				this.#draggedPiece.x = e.clientX - rect.left - this.#square / 2
				this.#draggedPiece.y = e.clientY - rect.top - this.#square / 2

				this.#render()
			} else {
				// Return the dragged piece into its originating position.
				this.#squares[this.#selectedSquare] = this.#draggedPiece.piece
				this.#draggedPiece = null
			}
		}
	}

	/**
	 * Handles piece drops or regular moves.
	 * @param {MouseEvent} e
	 */
	onMouseUp(e) {
		const rect = this.#context.canvas.getBoundingClientRect()
		const x = e.clientX - rect.left
		const y = e.clientY - rect.top

		const file = Math.floor(x / this.#square)
		const rank = Math.floor((this.#size - y) / this.#square)

		if (this.#draggedPiece !== null) {
			// End piece drag.
			this.#squares[8 * rank + file] = this.#draggedPiece.piece
			this.#draggedPiece = null

			this.#render()
		}
	}
}
