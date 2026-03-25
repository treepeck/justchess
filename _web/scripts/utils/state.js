/**
 * Enum representing the game termination reasons.
 * @enum {number}
 */
export const Termination = /** @type {const} */ ({
	Unterminated: 0,
	Abandoned: 1,
	Checkmate: 2,
	Stalemate: 3,
	InsufficientMaterial: 4,
	FiftyMoves: 5,
	ThreefoldRepetition: 6,
	Resignation: 7,
	Agreement: 8,
	TimeForfeit: 9,
})

/**
 * Enum representing the game result.
 * @enum {number}
 */
export const Result = /** @type {const} */ ({
	Unknown: 0,
	WhiteWon: 1,
	BlackWon: 2,
	Draw: 3,
})

/**
 * @param {Termination} t
 * @returns {string}
 */
export function formatTermination(t) {
	switch (t) {
		case Termination.Abandoned:
			return "game abandoned"
		case Termination.Agreement:
			return "by agreement"
		case Termination.Checkmate:
			return "by checkmate"
		case Termination.FiftyMoves:
			return "by fifty moves rule"
		case Termination.TimeForfeit:
			return "by time forfeit"
		case Termination.InsufficientMaterial:
			return "by insufficient material"
		case Termination.ThreefoldRepetition:
			return "by threefold repetition"
		case Termination.Resignation:
			return "by resignation"
		case Termination.Stalemate:
			return "by stalemate"
		default:
			return "unterminated"
	}
}

/**
 * @param {Result} r
 * @returns {string}
 */
export function formatResult(r) {
	switch (r) {
		case Result.WhiteWon:
			return "White won"
		case Result.BlackWon:
			return "Black won"
		case Result.Draw:
			return "Draw"
		default:
			return "Uknown result"
	}
}
