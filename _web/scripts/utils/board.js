/**
 * Enum representing chess pieces.
 * @enum {typeof Piece[keyof typeof Piece]}
 */
const Piece = /** @type {const} */ ({
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
 * @typedef {Object} Position
 * @property {number} x - The horizontal pixel coordinate on the canvas.
 * @property {number} y - The vertical pixel coordinate on the canvas.
 * @property {number} square - Index of the square [0-63].
 */

/**
 * Piece which is currently being dragged. Null is case no piece is being
 * dragged.
 * @typedef {Object} DraggedPiece
 * @property {Piece} piece - Piece type.
 * @property {Position} position
 */

/**
 * @callback MoveCallback
 * @param {number} from
 * @param {number} to
 * @returns {boolean}
 */

export default class BoardCanvas {
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
	 * @type {number}
	 */
	#piece
	/**
	 * Callback function which is called when the player performs the move.
	 * @type {MoveCallback}
	 */
	#onMove

	/**
	 * @param {Image} sheet
	 * @param {MoveCallback} onMove
	 */
	constructor(sheet, onMove) {
		// Add event listeners.
		boardCanvas.addEventListener("mousedown", (e) => {
			this.onMouseDown(e)
		})
		boardCanvas.addEventListener("mousemove", (e) => {
			this.onMouseMove(e)
		})
		boardCanvas.addEventListener("mouseup", (e) => {
			this.onMouseUp(e)
		})

		this.#context = boardCanvas.getContext("2d")
		this.#sheet = sheet
		this.#size = 0
		this.#selectedSquare = -1
		this.#draggedPiece = null
		this.#piece = 300
		// Assign onMove callback.
		this.#onMove = onMove

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
	render() {
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
				// Make dragged piece a bit bigger.
				const size = Math.round(this.#square * 1.1)
				if (this.#draggedPiece !== null) {
					this.#context.drawImage(
						this.#sheet,
						Math.floor(this.#draggedPiece.piece / 2) * this.#piece,
						this.#draggedPiece.piece % 2 === 0 ? 0 : this.#piece,
						this.#piece,
						this.#piece,
						this.#draggedPiece.x,
						this.#draggedPiece.y,
						size,
						size
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
		const dpr = window.devicePixelRatio || 1
		if (
			size * dpr >=
			this.#context.canvas.style.getPropertyValue("max-width")
		) {
			this.#size = Math.round(size * dpr)
			this.#context.scale(dpr, dpr)
		} else {
			this.#size = Math.round(size)
		}

		this.#square = Math.round(this.#size / 8)
		this.#context.canvas.width = this.#size
		this.#context.canvas.height = this.#size

		this.render()
	}

	/**
	 * Handles player's clicks on the canvas.
	 * @param {MouseEvent} e
	 */
	onMouseDown(e) {
		const { x, y, square } = this.#getPositionOfEvent(e)

		const selected =
			this.#selectedSquare > -1
				? this.#squares[this.#selectedSquare]
				: Piece.NP

		if (selected !== Piece.NP) {
			// Perform the move and update the position if it was successfull.
			if (
				square !== this.#selectedSquare &&
				this.#onMove(this.#selectedSquare, square)
			) {
				this.#squares[square] = selected
				this.#squares[this.#selectedSquare] = Piece.NP
				this.#selectedSquare = -1
				this.render()
				return
			}
		}

		this.#selectedSquare = square
		const piece = this.#squares[square]
		if (piece !== Piece.NP) {
			// Temporary remove the piece from its square while its being dragged.
			this.#squares[square] = Piece.NP
			// Begin piece drag.
			this.#draggedPiece = {
				x: Math.round(x - (this.#square * 1.1) / 2), // Center piece horizontally.
				y: Math.round(y - (this.#square * 1.1) / 2), // Center piece vertically.
				from: this.#selectedSquare,
				piece: piece,
			}
		}

		this.render()
	}

	/**
	 * Handles player's mouse movement on the canvas.
	 * @param {MouseEvent} e
	 */
	onMouseMove(e) {
		if (this.#draggedPiece !== null) {
			const isLeftButtonPressed = e.buttons === 1

			if (isLeftButtonPressed) {
				const { x, y } = this.#getPositionOfEvent(e)
				// Center and move dragged piece with the cursor.
				this.#draggedPiece.x = x - this.#square / 2
				this.#draggedPiece.y = y - this.#square / 2

				this.render()
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
		// Shortcut: no move to be performed.
		if (!this.#draggedPiece) {
			return
		}

		// End piece drag.
		const piece = this.#draggedPiece.piece
		this.#squares[this.#selectedSquare] = piece
		this.#draggedPiece = null

		const { square } = this.#getPositionOfEvent(e)
		// Shortcut: to and from squares are the same.
		if (square === this.#selectedSquare) {
			this.render()
			return
		}

		// Perform the move and update the position if it was successfull.
		if (this.#onMove(this.#selectedSquare, square)) {
			this.#squares[square] = piece
			this.#squares[this.#selectedSquare] = Piece.NP
			this.#selectedSquare = -1
		}
		this.render()
	}

	/**
	 * @param {MouseEvent} e
	 * @returns {Position}
	 */
	#getPositionOfEvent(e) {
		const rect = this.#context.canvas.getBoundingClientRect()
		const scaleX = this.#context.canvas.width / rect.width
		const scaleY = this.#context.canvas.height / rect.height

		const x = (e.clientX - rect.left) * scaleX
		const y = (e.clientY - rect.top) * scaleY

		const file = Math.floor(x / this.#square)
		const rank = Math.floor((this.#size - y) / this.#square)

		return {
			x: x,
			y: y,
			square: 8 * rank + file,
		}
	}
}
