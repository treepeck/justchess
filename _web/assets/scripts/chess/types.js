
/**
 * @readonly
 * @enum {number}
 */
export const PieceType = {
	WPawn:     0,
	BPawn:     1,
	WKnight:   2,
	BKnight:   3,
	WBishop:   4,
	BBishop:   5,
	WRook:     6,
	BRook:     7,
	WQueen:    8,
	BQueen:    9,
	WKing:     10,
	BKing:     11,
	PieceNone: -1
}


/**
 * @readonly
 * @enum {number}
 */
export const CastlingRights = {
	WhiteShort: 1,
	WhiteLong:  2,
	BlackShort: 4,
	BlackLong:  8
}

/**
 * @readonly
 * @enum {number}
 */
export const Color = {
	White: 0,
	Black: 1
}

/**
 * @typedef {Object} Position
 * @property {Map<number, PieceType>} pieces
 * @property {Color} activeColor
 * @property {CastlingRights} castlingRights
 * @property {number} epTarget
 * @property {number} halfmoveCnt
 * @property {number} fullmoveCnt
 */
