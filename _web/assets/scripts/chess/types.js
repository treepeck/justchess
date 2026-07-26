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
 * @readonly
 * @enum {number}
 */
export const MoveType = {
	Normal:    0,
	Castling:  1,
	Promotion: 2,
	EnPassant: 3
}

/**
 * @readonly
 * @enum {number}
 */
export const PromotionFlag = {
	Knight: 0,
	Bishop: 1,
	Rook:   2,
	Queen:  3
}

/**
 * @readonly
 * @enum {number}
 */
export const Result = {
	Unknown:  0,
	WhiteWon: 1,
	BlackWon: 2,
	Draw:     3
}

/**
 * @readonly
 * @enum {number}
 */
export const Termination = {
	Unterminated:         0,
	Abandoned:            1,
	Checkmate:            2,
	Stalemate:            3,
	InsufficientMaterial: 4,
	FiftyMoves:           5,
	ThreefoldRepetition:  6,
	Resignation:          7,
	Agreement:            8,
	TimeForfeit:          9
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
