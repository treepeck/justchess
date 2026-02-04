/** @enum {typeof MoveType[keyof typeof MoveType]} */
export const MoveType = /** @type {const} */ ({
	Normal: 0,
	Castling: 1,
	Promotion: 2,
	EnPassant: 3,
})

/** @enum {typeof PromotionFlag[keyof typeof PromotionFlag]} */
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
		// @ts-expect-error
		this.promoPiece = (raw >> 12) & 0x3
		// @ts-expect-error
		this.moveType = (raw >> 14) & 0x3
	}
}

/**
 * @typedef {Object} CompletedMove
 * @property {string} s - Standard Algebraic Notation of the move.
 * @property {import("../chess/move.js").Move} m - Encoded move.
 * @property {number} t - Remaining time on the player's clock.
 */
