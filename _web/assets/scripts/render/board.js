import { Color, PieceType } from "/assets/scripts/chess/types.js"

/**
 * @typedef {Object} Piece
 * @property {HTMLDivElement} element
 * @property {pieceType} PieceType
 */

/**
 * @typedef {Object} Coords
 * @property {number} x
 * @property {number} y
 */

/**
 * @param {PieceType} pt
 * @returns {string}
 */
function pieceType2ClassName(pt) {
	switch (pt) {
		case PieceType.WPawn:   return "P"
		case PieceType.WKnight: return "N"
		case PieceType.WBishop: return "B"
		case PieceType.WRook:   return "R"
		case PieceType.WQueen:  return "Q"
		case PieceType.WKing:   return "K"
		case PieceType.BPawn:   return "p"
		case PieceType.BKnight: return "n"
		case PieceType.BBishop: return "b"
		case PieceType.BRook:   return "r"
		case PieceType.BQueen:  return "q"
		case PieceType.BKing:   return "k"
	}
}

/**
 * Manages board rendering.
 * The main idea is to minimize the number of DOM manipulations.
 * Piece movement is implemented via CSS transforms using dynamic variables.
 * The board knows as little as possible about the game state.
 * Its responsibility is to receive a Position and render it to the screen.
 * It also manages the dragging process, although move handling is delegated
 * to callbacks.
 * When a new Position arrives, it analyzes the difference between the current
 * and the new positions, and transforms only the pieces whose placement has
 * changed.
 */
export class Board {
	/**
	 * @type {HTMLDivElement}
	 */
	element
	/**
	 * Maps square index to the piece element that is being rendered.
	 * @type {Map<number, Piece>}
	 */
	elements
	/**
	 * @type {Position}
	 */
	position
	/**
	 * @type {Color}
	 */
	orientation
	/**
	 * @type {number}
	 */
	selectedSquare
	/**
	 * @type {HTMLDivElement}
	 */
	squareElement
	/**
	 * @callback
	 */
	onMove

	/**
	 * @param {Position} position
	 * @param {HTMLDivElement} element
	 * @param {moveHandler} onMove
	 */
	constructor(position, element, onMove) {
		this.position = position
		this.element = element
		this.elements = new Map()
		this.orientation = Color.White
		this.selectedSquare = -1
		this.onMove = onMove

		this.squareElement = document.createElement("div")
		this.squareElement.classList.add("board-square")
		this.squareElement.classList.add("board-selected")
		this.squareElement.style.visibility = "hidden"
		this.element.appendChild(this.squareElement)
	}

	/**
	 * Renders the current position.
	 */
	render() {
		for (const [square, pieceType] of this.position.pieces) {
			let onBoard = this.elements.get(square)
			if (!onBoard) {
				const el = document.createElement("div")
				// Render the piece.
				onBoard = { pieceType: pieceType, element: el }
				el.classList.add("board-square")
				el.classList.add("board-piece")
				el.classList.add(pieceType2ClassName(pieceType))
				this.element.appendChild(el)
				this.elements.set(square, onBoard)
			} else if (pieceType !== onBoard.pieceType) {
				// Change the piece type.
				onBoard.element.classList.remove(pieceType2ClassName(onBoard.pieceType))
				onBoard.element.classList.add(pieceType2ClassName(pieceType))
			}
			// Transform the piece.
			const { x, y } = this.square2Coords(square)
			this.translate(onBoard.element, x, y)
		}
	}

	/**
	 * @param {PointerEvent} e
	 */
	onClick(e) {
		if (e.buttons == 2) return // Ignore right clicks.

		const { square, squareSize, coords } = this.event2Coords(e)

		this.selectedSquare = square
		this.squareElement.style.visibility = "visible"
		const { x, y } = this.square2Coords(square)
		this.translate(this.squareElement, x, y)

		const piece = this.position.pieces.get(square)
		if (piece) {
			// Begin piece drag.
			const { element } = this.elements.get(square)
			this.translate(element, x - squareSize / 2, y - squareSize / 2)
		}
	}

	/**
	 * @param {PointerEvent} e
	 */
	onDrag(e) {
		if (e.buttons == 2) return // Ignore right clicks.

		const p = this.elements.get(this.selectedSquare)
		if (!p) return

		let square, squareSize, coords
		try {
			let res = this.event2Coords(e)
			square = res.square
			squareSize = res.squareSize
			coords = res.coords
			// Center the dragged piece.
			coords.x -= squareSize / 2
			coords.y -= squareSize / 2
		} catch (e) {
			// Return the piece to it's initial position and reset the selected square.
			coords = this.square2Coords(this.selectedSquare)
			this.selectedSquare = -1
			this.squareElement.style.visibility = "hidden"
		} finally {
			this.translate(p.element, coords.x, coords.y)
		}
	}

	/**
	 * @param {PointerEvent} e
	 */
	onDrop(e) {
		if (e.buttons == 2) return // Ignore right clicks.

		const p = this.elements.get(this.selectedSquare)
		if (!p) return

		let square, squareSize, coords
		try {
			let res = this.event2Coords(e)
			square = res.square
			coords = this.square2Coords(square)
			this.onMove()
		} catch (e) {
			// Return the piece to it's initial position and reset the selected square.
			coords = this.square2Coords(this.selectedSquare)
		} finally {
			this.translate(p.element, coords.x, coords.y)
			this.selectedSquare = -1
			this.squareElement.style.visibility = "hidden"
		}
	}

	/**
	 * @param {PointerEvent} e
	 */
	onRightClick(e) {
		e.stopPropagation()
	}

	/**
	 * @param {HTMLDivElement} element
	 * @param {number} x
	 * @param {number} y
	 */
	translate(element, x, y) {
		element.style.setProperty("--x", `${x}px`)
		element.style.setProperty("--y", `${y}px`)
	}

	/**
	 * @param {number} square
	 * @returns {Coords}
	 */
	square2Coords(square) {
		const size = this.element.clientWidth
		const squareSize = size / 8
		const offset = size - squareSize

		const file = square % 8
		const rank = Math.floor(square / 8)

		const x = this.orientation == Color.White
			? file * squareSize
			: offset - file * squareSize
		const y = this.orientation == Color.White
			? offset - rank * squareSize
			: rank * squareSize

		return { x: x, y: y }
	}

	/**
	 * @param {PointerEvent} e
	 * @returns {{
	 *   square:     number,
	 *   squareSize: number,
	 *   coords:     Coords,
	 * }}
	 * @throws {Error} when the event coords are outside of the board element.
	 */
	event2Coords(e) {
		const size = this.element.clientWidth
		const squareSize = size / 8
		const offset = size - squareSize

		const rect = this.element.getBoundingClientRect(e)

		const x = this.orientation == Color.White
			? e.clientX - rect.left
			: rect.right - e.clientX
		const y = this.orientation == Color.White
			?  e.clientY - rect.top
			: rect.bottom - e.clientY

		if (x < 0 || x > size || y < 0 || y > size) throw new Error("event is outside of the board")

		const file = Math.floor(x / squareSize)
		const rank = Math.floor((size - y) / squareSize)
		const square = rank * 8 + file

		return {
			square: square,
			squareSize: squareSize,
			coords: { x: x, y: y },
		}
	}

	observeResize() {
		const observer = new ResizeObserver((entries) => {
			for (const entry of entries) {
				this.render()
			}
		})
		observer.observe(this.element)
	}

	registerEventHandlers() {
		this.element.onpointerdown = (e) => { this.onClick(e) }
		this.element.onpointermove = (e) => { this.onDrag(e) }
		this.element.onpointerup = (e) => { this.onDrop(e) }
		this.element.oncontextmenu = (e) => { this.onRightClick(e) }
	}
}
