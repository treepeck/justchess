import { getOrPanic, create } from "../utils/dom"

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
 * Appends move SAN to moves table.
 * @param {string} san
 * @param {number} moveIndex
 */
export function appendMoveToTable(san, moveIndex) {
	// Half move index.
	const ply = Math.ceil(moveIndex / 2)

	// Append row to the table after each black move.
	if (moveIndex % 2 !== 0) {
		const row = create("div", "moves-row", `row${ply}`)
		// Append half-move index to the row.
		const ind = create("div", "moves-ply")
		ind.textContent = `${ply}.`
		row.appendChild(ind)
		// Append row to the table.
		getOrPanic("moves").appendChild(row)
	}

	// Append move to the row.
	const move = create("div", "moves-san")
	move.textContent = san
	getOrPanic(`row${ply}`).appendChild(move)

	// Scroll table to bottom.
	const table = getOrPanic(`moves`)
	table.scrollTo({
		top: table.scrollHeight,
		behavior: "smooth",
	})
}

/**
 * @typedef {Object} CompletedMove
 * @property {Move} m - Encoded move.
 * @property {number} t - Remaining time on the player's clock.
 * @property {string} s - Standard Algebraic Notation of the move.
 */
