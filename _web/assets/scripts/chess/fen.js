import { Piece, Color, CastlingRights } from "/assets/scripts/chess/types.js"

export const initPos =
	"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

// prettier-ignore
const square2String = [
	"a1", "b1", "c1", "d1", "e1", "f1", "g1", "h1",
	"a2", "b2", "c2", "d2", "e2", "f2", "g2", "h2",
	"a3", "b3", "c3", "d3", "e3", "f3", "g3", "h3",
	"a4", "b4", "c4", "d4", "e4", "f4", "g4", "h4",
	"a5", "b5", "c5", "d5", "e5", "f5", "g5", "h5",
	"a6", "b6", "c6", "d6", "e6", "f6", "g6", "h6",
	"a7", "b7", "c7", "d7", "e7", "f7", "g7", "h7",
	"a8", "b8", "c8", "d8", "e8", "f8", "g8", "h8",
]

/**
 * Parses piece the FEN string. It's the caller's responsibility to validate the input.
 * @param {string} fen
 * @returns {import("/assets/scripts/chess/types.js").Position}
 */
export function parseFen(fen) {
	const p =
		/** @type {import("/assets/scripts/chess/types.js").Position} */ ({})

	// Separate FEN fields.
	const fields = fen.split(" ", 6)

	// Parse piece placement.
	p.pieces = parsePiecePlacement(fields[0])

	// Parse active color.
	if (fields[1] === "b") {
		p.activeColor = Color.Black
	} else {
		p.activeColor = Color.White
	}

	// Parse castling rights.
	for (let i = 0; i < fields[2].length; i++) {
		switch (fields[2][i]) {
			case "K":
				p.castlingRights |= CastlingRights.WhiteShort
				break
			case "Q":
				p.castlingRights |= CastlingRights.WhiteLong
				break
			case "k":
				p.castlingRights |= CastlingRights.BlackShort
				break
			case "q":
				p.castlingRights |= CastlingRights.BlackLong
				break
		}
	}

	// Parse en passant target square.
	for (let i = 0; i < square2String.length; i++) {
		if (square2String[i] === fields[3]) {
			p.epTarget = i
		}
	}

	// Parse halfmove counter.
	p.halfmoveCnt = parseInt(fields[4])

	// Parse fullmove counter.
	p.fullmoveCnt = parseInt(fields[5])

	return p
}

/**
 * Parses piece the piece placement part of the FEN string.
 * It's the caller's responsibility to validate the input.
 * @param {string} piecePlacement
 * @returns {Map<number, Piece>}
 */
export function parsePiecePlacement(piecePlacement) {
	const pieces = new Map()

	let square = 56

	// Piece placement data describes each rank beginning from the eigth.
	for (let i = 0; i < piecePlacement.length; i++) {
		const char = piecePlacement[i]

		if (char === "/") {
			// Rank separator.
			square -= 16
			// Number of consecutive empty squares.
		} else if (char >= "1" && char <= "8") {
			square += parseInt(char)
		} else {
			// There is piece on a square.
			let piece = Piece.WPawn
			// Manual switch construction is ~3x faster than map approach.
			switch (char) {
				case "N":
					piece = Piece.WKnight
					break
				case "B":
					piece = Piece.WBishop
					break
				case "R":
					piece = Piece.WRook
					break
				case "Q":
					piece = Piece.WQueen
					break
				case "K":
					piece = Piece.WKing
					break
				case "p":
					piece = Piece.BPawn
					break
				case "n":
					piece = Piece.BKnight
					break
				case "b":
					piece = Piece.BBishop
					break
				case "r":
					piece = Piece.BRook
					break
				case "q":
					piece = Piece.BQueen
					break
				case "k":
					piece = Piece.BKing
					break
			}
			pieces.set(square, piece)
			square++
		}
	}

	return pieces
}
