import { PieceType } from "./piece"
import { Square } from "./move"

/**
 * Parses the piece placement data of the Forsyth-Edwards Notation string.
 * @param {string} fen
 * @returns {{t: PieceType, s: Square}[]}
 */
export function parsePiecePlacement(fen) {
	if (!fen || fen == "") return []

	const rows = fen.split(" ")[0].split("/")
	const pieces = /** @type {{t: PieceType, s: Square}[]} */ ([])

	/** @type Map<string, PieceType> */
	const mapping = new Map()
	mapping.set("P", PieceType.WP)
	mapping.set("p", PieceType.BP)
	mapping.set("N", PieceType.WN)
	mapping.set("n", PieceType.BN)
	mapping.set("B", PieceType.WB)
	mapping.set("b", PieceType.BB)
	mapping.set("R", PieceType.WR)
	mapping.set("r", PieceType.BR)
	mapping.set("Q", PieceType.WQ)
	mapping.set("q", PieceType.BQ)
	mapping.set("K", PieceType.WK)
	mapping.set("k", PieceType.BK)

	let sqInd = 0
	for (let i = 7; i >= 0; i--) {
		for (let j = 0; j < rows[i].length; j++) {
			let next = mapping.get(rows[i][j])
			if (next !== undefined) {
				pieces.push({ t: next, s: sqInd })
				sqInd++
			} else {
				// Skip empty squares.
				sqInd += parseInt(rows[i][j])
			}
		}
	}
	return pieces
}
