// import { Piece } from "/assets/scripts/chess/types.js"
import { initPos, parseFen } from "/assets/scripts/chess/fen.js"

// prettier-ignore
const classNames = ["P", "p", "N", "n", "B", "b", "R", "r", "Q", "q", "K", "k"]

/**
 * Board handles rendering of the 3D board. It manages drag&drop, square selection,
 * animations, etc. However, it doesn't know anything about chess logic. The 3D
 * effect is achieved by using CSS 3D transforms. That is every part of the board
 * scene is a 2D flat positioned in document's 3D space.
 */
export class Board {
	/**
	 * Front element of the board, to which all UI elements are rendered.
	 * @type {HTMLDivElement}
	 */
	front
	/**
	 * @type {import("/assets/scripts/chess/types.js").Position}
	 */
	position

	constructor() {
		this.front = document.getElementById("front")

		this.appendRanks()
		this.appendFiles()

		this.position = parseFen(initPos)
		this.appendPieces()
	}

	/**
	 * Responsively translates an element on the front. Not all types of elements
	 * are positioned the same way. For instance, files and ranks are positioned
	 * @param {HTMLDivElement} element
	 * @param {number} square
	 */
	translate(element, square) {
		const file = square % 8
		const rank = Math.floor(square / 8)

		const size = this.front.clientWidth
		const squareSize = size / 8
		// offset is needed to calculate the correct Y coordinate
		// since the 0,0 point in located in top left corner.
		const offset = size - squareSize

		const x = file * squareSize
		const y = offset - rank * squareSize

		element.style.setProperty("--square", `${square}`)
		element.style.setProperty("--x", `${x}px`)
		element.style.setProperty("--y", `${y}px`)
	}

	/**
	 * Appends rank elements to front and positions them.
	 */
	appendRanks() {
		const ranks = ["1", "2", "3", "4", "5", "6", "7", "8"]
		for (let i = 0; i < ranks.length; i++) {
			const rankDiv = document.createElement("div")
			rankDiv.textContent = ranks[i]
			rankDiv.classList.add("board-object")
			rankDiv.classList.add("board-rank")
			this.front.appendChild(rankDiv)
			this.translate(rankDiv, i * 8)
		}
	}

	/**
	 * Appends files elements to front and positions them.
	 */
	appendFiles() {
		const files = ["a", "b", "c", "d", "e", "f", "g", "h"]
		for (let i = 0; i < files.length; i++) {
			const fileDiv = document.createElement("div")
			fileDiv.textContent = files[i]
			fileDiv.classList.add("board-object")
			fileDiv.classList.add("board-file")
			this.front.appendChild(fileDiv)
			this.translate(fileDiv, i)
		}
	}

	appendPieces() {
		for (const [square, piece] of this.position.pieces) {
			const pieceDiv = document.createElement("div")
			pieceDiv.classList.add("board-piece")
			pieceDiv.classList.add(classNames[piece])
			this.front.appendChild(pieceDiv)
			this.translate(pieceDiv, square)
		}
	}

	/**
	 * Makes board responsive by observing the size changes of the front and repositioning elements.
	 */
	registerResizeObserver() {
		const observer = new ResizeObserver(() => {
			// Reposition all elements.
			for (const element of this.front.children) {
				const square = parseInt(
					element.style.getPropertyValue("--square"),
				)
				this.translate(element, square)
			}
		})
		observer.observe(this.front)
	}
}
