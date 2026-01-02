import { Piece, DraggedPiece } from "/public/js/types.js"

// Board size in pixels.
const BOARD  = 560
// Square size in pixels.
const SQUARE = BOARD / 8
// Source image piece size in pixels.
const PIECE = 90

class Board {
	constructor(ctx, sheet) {
		// Canvas rendering content.
		this.ctx = ctx
		// Sprite sheet.
		this.sheet = sheet

		this.marked = []
		this.selected = Piece.NP
		this.draggedPiece = null
		this.squares = [
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

	draw() {
		for (let rank = 0; rank < 8; rank++) {
			for (let file = 0; file < 8; file++) {
				// Draw board squares.
				this.ctx.fillStyle = "#b5936e"
				if ((rank % 2 !== 0 && file % 2 !== 0) ||
					(rank % 2 === 0 && file % 2 === 0)) {
					this.ctx.fillStyle = "#41312f"
				}

				const x = file * SQUARE
				const y = BOARD - SQUARE - rank * SQUARE
				this.ctx.fillRect(x, y, SQUARE, SQUARE)

				const ind = 8 * rank + file
				// Draw selected square.
				if (ind === this.selectedSquare) {
					this.ctx.fillStyle = "green"
					this.ctx.fillRect(x, y, SQUARE, SQUARE)
				}

				// Draw dragged piece.
				if (this.draggedPiece) {
					this.ctx.drawImage(
						this.sheet,
						Math.floor(this.draggedPiece.piece / 2) * PIECE,
						this.draggedPiece.piece % 2 === 0 ? 0 : PIECE,
						PIECE, PIECE,
						this.draggedPiece.x, this.draggedPiece.y,
						SQUARE, SQUARE
					)
				}

				// Draw pieces.
				if (this.squares[ind] !== Piece.NP) {
					this.ctx.drawImage(
						this.sheet,
						Math.floor(this.squares[ind] / 2) * PIECE,
						this.squares[ind] % 2 === 0 ? 0 : PIECE,
						PIECE, PIECE,
						x, y,
						SQUARE, SQUARE
					)
				}
			}
		}
	}

	onMouseDown(e) {
		const rect = e.target.getBoundingClientRect()

		const x = e.clientX - rect.left
		const y = e.clientY - rect.top

		const file = Math.floor(x / SQUARE)
		const rank = Math.floor((BOARD - y) / SQUARE)

		this.selected = 8 * rank + file

		// Begin piece drag.
		if (this.squares[this.selected] !== Piece.NP) {
			this.draggedPiece = new DraggedPiece(
				x - SQUARE / 2, // Center dragged piece horizontally.
				y - SQUARE / 2, // Center dragged piece vertically.
				this.squares[this.selected],
				this.selected,
			)
			// Remove the piece from its originating square while it's being dragged.
			this.squares[this.selected] = Piece.NP
		}

		// Redraw the board.
		this.draw()
	}

	onMouseMove(e) {
		// Update dragged piece position.
		if (this.draggedPiece) {
			const rect = e.target.getBoundingClientRect()
			this.draggedPiece.x = (e.clientX - rect.left) - SQUARE / 2
			this.draggedPiece.y = (e.clientY - rect.top) - SQUARE / 2

			// Redraw the board.
			this.draw()
		}
	}

	onMouseUp(e) {
		// Handle piece drop.
		if (this.draggedPiece) {
			const rect = e.target.getBoundingClientRect()

			const x = e.clientX - rect.left
			const y = e.clientY - rect.top

			const file = Math.floor(x / SQUARE)
			const rank = Math.floor((BOARD - y) / SQUARE)

			this.squares[8*rank+file] = this.draggedPiece.piece
			this.draggedPiece = null
			this.selected = -1

			// Redraw the board.
			this.draw()
		}
	}
}

const sheet = new Image()
sheet.onload = () => {
	const canvas = document.getElementById("board")
	const ctx = canvas.getContext("2d")

	const board = new Board(ctx, sheet)
	board.draw()

	// Add event listeners.
	canvas.addEventListener("mousedown", e => board.onMouseDown(e))
	canvas.addEventListener("mousemove", e => board.onMouseMove(e))
	canvas.addEventListener("mouseup", e => board.onMouseUp(e))
}
sheet.src = "/public/img/sheet.png"