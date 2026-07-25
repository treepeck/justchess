import { CastlingRights, Color, PieceType } from "/assets/scripts/chess/types.js"

/**
 * Maps each board square to its string representation.
 * @type {string[]}
 */
const square2String = [
	"a1", "b1", "c1", "d1", "e1", "f1", "g1", "h1",
	"a2", "b2", "c2", "d2", "e2", "f2", "g2", "h2",
	"a3", "b3", "c3", "d3", "e3", "f3", "g3", "h3",
	"a4", "b4", "c4", "d4", "e4", "f4", "g4", "h4",
	"a5", "b5", "c5", "d5", "e5", "f5", "g5", "h5",
	"a6", "b6", "c6", "d6", "e6", "f6", "g6", "h6",
	"a7", "b7", "c7", "d7", "e7", "f7", "g7", "h7",
	"a8", "b8", "c8", "d8", "e8", "f8", "g8", "h8"
]

/**
 * Port of the [judo/repo/chego/fen.go].ParseFen to JS.
 * @param {string} fen
 * @returns {Position}
 */
export function parseFen(fen) {
	const p = {}

	// Separate FEN fields.
	const fields = fen.split(" ", 6)

	// Parse piece placement.
	p.pieces = parsePiecePlacement(fields[0])

	// Parse active color.
	p.activeColor = fields[1] == "b" ? Color.Black : Color.White

	// Parse castling rights.
	p.castlingRights = 0
	for (let i = 0; i < fields[2].length; i++) {
		switch (fields[2][i]) {
		case 'K':
			p.castlingRights |= CastlingRights.WhiteShort
			break
		case 'Q':
			p.castlingRights |= CastlingRights.WhiteLong
			break
		case 'k':
			p.castlingRights |= CastlingRights.BlackShort
			break
		case 'q':
			p.castlingRights |= CastlingRights.BlackLong
		}
	}

	// Parse en passant target square.
	p.epTarget = -1
	for (let i = 0; i < square2String.length; i++) {
		if (square2String[i] == fields[3]) {
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
 * Port of the [judo/repo/chego/fen.go].ParseBitboards to JS.
 * @param {string} piecePlacement
 * @return {Map<number, PieceType>}
 * @throws {Error} The provided string must be valid FEN part.
 */
function parsePiecePlacement(piecePlacement) {
	const pieces = new Map()

	let square = 56

	// Piece placement data describes each rank beginning from the eigth.
	for (let i = 0; i < piecePlacement.length; i++) {
		const char = piecePlacement[i]

		if (char == '/') { // Rank separator.
			square -= 16
			// Number of consecutive empty squares.
		} else if (char >= '1' && char <= '8') {
			// Convert byte to the integer it represents.
			square += parseInt(char, 10)
		} else { // There is piece on a square.
			let piece = PieceType.WPawn
			// Manual switch construction is ~3x faster than map approach.
			switch (char) {
			case 'N':
				piece = PieceType.WKnight
				break
			case 'B':
				piece = PieceType.WBishop
				break
			case 'R':
				piece = PieceType.WRook
				break
			case 'Q':
				piece = PieceType.WQueen
				break
			case 'K':
				piece = PieceType.WKing
				break
			case 'p':
				piece = PieceType.BPawn
				break
			case 'n':
				piece = PieceType.BKnight
				break
			case 'b':
				piece = PieceType.BBishop
				break
			case 'r':
				piece = PieceType.BRook
				break
			case 'q':
				piece = PieceType.BQueen
				break
			case 'k':
				piece = PieceType.BKing
				break
			}
			pieces.set(square, piece)
			square++
		}
	}
	return pieces
}
