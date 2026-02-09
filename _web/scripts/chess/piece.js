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
function pieceType2String(pieceType) {
	switch (pieceType) {
		case PieceType.WP:
			return "WP"
		case PieceType.BP:
			return "BP"
		case PieceType.WN:
			return "WN"
		case PieceType.BN:
			return "BN"
		case PieceType.WB:
			return "WB"
		case PieceType.BB:
			return "BB"
		case PieceType.WR:
			return "WR"
		case PieceType.BR:
			return "BR"
		case PieceType.WQ:
			return "WQ"
		case PieceType.BQ:
			return "BQ"
		case PieceType.WK:
			return "WK"
		default:
			return "BK"
	}
}
