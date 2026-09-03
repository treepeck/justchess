/**
 * @readonly
 * @enum {number}
 */
export const Piece = {
	WPawn: 0,
	BPawn: 1,
	WKnight: 2,
	BKnight: 3,
	WBishop: 4,
	BBishop: 5,
	WRook: 6,
	BRook: 7,
	WQueen: 8,
	BQueen: 9,
	WKing: 10,
	BKing: 11,
	PieceNone: -1,
}

/**
 * @readonly
 * @enum {number}
 */
export const PromotionFlag = {
	Knight: 0,
	Bishop: 1,
	Rook: 2,
	Queen: 3,
}

/**
 * @readonly
 * @enum {number}
 */
export const Color = {
	White: 0,
	Black: 1,
	Both: 2,
}

/**
 * @readonly
 * @enum {number}
 */
export const MoveType = {
	Normal: 0,
	Castling: 1,
	Promotion: 2,
	EnPassant: 3,
}

/**
 * @typedef {Object} Move
 * @property {number} to - Destination square index.
 * @property {number} from - Source square index.
 * @property {PromotionFlag} promoPiece
 * @property {MoveType} moveType
 */

/**
 * @param {number} raw
 * @returns {Move}
 */
export function decodeMove(raw) {
	return {
		to: raw & 0x3f,
		from: (raw >> 6) & 0x3f,
		promoPiece: (raw >> 12) & 0x3,
		moveType: (raw >> 14) & 0x3,
	}
}

/**
 * @readonly
 * @enum {number}
 */
export const CastlingRights = {
	WhiteShort: 1,
	WhiteLong: 2,
	BlackShort: 4,
	BlackLong: 8,
}

/**
 * @typedef {Object} Position
 * @property {Map<number, Piece>} pieces
 * @property {Color} activeColor
 * @property {CastlingRights} castlingRights
 * @property {number} epTarget
 * @property {number} halfmoveCnt
 * @property {number} fullmoveCnt
 */
