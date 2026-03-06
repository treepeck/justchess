import { create } from "../utils/dom"

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
		this.element = /** @type {HTMLDivElement} */ (create("div", "piece"))
		this.element.classList.add(`${pieceType2String(pieceType)}`)
		this.pieceType = pieceType
	}
}

/**
 * @param {PieceType} pieceType
 * @returns {string}
 */
export function pieceType2String(pieceType) {
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
