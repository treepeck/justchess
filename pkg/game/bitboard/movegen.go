package bitboard

import (
	"math/bits"

	"justchess/pkg/game/enums"
)

// The following block on contants defines the bit masks needed to
// correctly calculate possible moves by performing bitwise
// operations on a bitboard.
const (
	notA  uint64 = 0xFEFEFEFEFEFEFEFE // Mask for all files except the A.
	notH  uint64 = 0x7F7F7F7F7F7F7F7F // Mask for all files except the H.
	notAB uint64 = 0xFCFCFCFCFCFCFCFC // Mask for all files except the A and B.
	notGH uint64 = 0x3F3F3F3F3F3F3F3F // Mask for all files except the G and H.
)

var GetLSB = bits.TrailingZeros64

// 0-5: To (destination) square index;
// 6-11: From (origin/source) square index;
// 12-15: Move type.
type Move uint16

func NewMove(to, from int, mt enums.MoveType) Move {
	return Move(to | (from << 6) | int(mt<<12))
}

func (m Move) To() int {
	return int(m) & 0x3F
}

func (m Move) From() int {
	return int(m>>6) & 0x3F
}

func (m Move) Type() enums.MoveType {
	return enums.MoveType(m >> 12 & 0xF)
}

///////////////////////////////////////////////////////////////
//                          KING                             //
///////////////////////////////////////////////////////////////

func genKingMovesPattern(king uint64) (moves uint64) {
	moves |= (king & notH) >> 9 // South east. (-9 squares)
	moves |= king >> 8          // South south. (-8 squares)
	moves |= (king & notA) >> 7 // South west. (-7 squares)
	moves |= (king & notA) >> 1 // West. (-1 square)
	moves |= (king & notH) << 1 // East. (+1 square)
	moves |= (king & notA) << 7 // North west. (+7 squares)
	moves |= king << 8          // North north. (+8 squares)
	moves |= (king & notH) << 9 // North east. (+9 squares)
	return
}

func genKingLegalMoves(from int, allies, enemies, attacked uint64,
	can00, can000 bool) (moves []Move) {
	movesBB := genKingMovesPattern(1<<from) & ^allies & ^attacked
	for i := GetLSB(movesBB); movesBB != 0; i = GetLSB(movesBB) {
		if 1<<i&enemies != 0 {
			moves = append(moves, NewMove(i, from, enums.Capture))
		} else {
			moves = append(moves, NewMove(i, from, enums.Quiet))
		}
		movesBB &= movesBB - 1
	}
	// Castling implementation.
	if can000 && (0xE&allies == 0) && (0xE&attacked == 0) {
		moves = append(moves, NewMove(enums.B1, from, enums.QueenCastle))
	}
	if can00 && (0x60&allies == 0) && (0x60&attacked == 0) {
		moves = append(moves, NewMove(enums.G1, from, enums.KingCastle))
	}
	return
}

///////////////////////////////////////////////////////////////
//                          PAWN                             //
///////////////////////////////////////////////////////////////

// TODO: implement en passant.

func genWhitePawnsAttackPattern(pawns uint64) uint64 {
	return ((pawns & notA) << 7) | ((pawns & notH) << 9)
}

func genBlackPawnsAttackPattern(pawns uint64) uint64 {
	return ((pawns & notA) >> 9) | ((pawns & notH) >> 7)
}

func genWhitePawnPseudoLegalMoves(from int, allies uint64,
	enemies uint64) (moves []Move) {
	var pawn, occupied uint64 = 1 << from, allies | enemies
	to := pawn << 8
	// If the forward square is vacant.
	if occupied&to == 0 {
		if to&0xFF00000000000000 != 0 {
			moves = append(moves, NewMove(GetLSB(to), from, enums.QueenPromo))
		} else {
			moves = append(moves, NewMove(GetLSB(to), from, enums.Quiet))
			if pawn&0xFF00 != 0 && occupied&(to<<8) == 0 {
				moves = append(moves, NewMove(GetLSB(to<<8), from, enums.DoublePawnPush))
			}
		}
	}
	leftCapture, rightCapture := pawn&notA<<7, pawn&notH<<9
	if leftCapture&enemies != 0 {
		moves = append(moves, NewMove(GetLSB(leftCapture), from, enums.Capture))
	}
	if rightCapture&enemies != 0 {
		moves = append(moves, NewMove(GetLSB(rightCapture), from, enums.Capture))
	}
	return
}

func genBlackPawnPseudoLegalMoves(from int, allies,
	enemies uint64) (moves []Move) {
	var pawn, occupied uint64 = 1 << from, allies | enemies
	to := pawn >> 8
	// If the forward square is vacant.
	if occupied&to == 0 {
		if uint64(to)&0xFF != 0 {
			moves = append(moves, NewMove(GetLSB(to), from, enums.QueenPromo))
		} else {
			moves = append(moves, NewMove(GetLSB(to), from, enums.Quiet))
			if pawn&0xFF000000000000 != 0 && occupied&(to>>8) == 0 {
				moves = append(moves, NewMove(GetLSB(to>>8), from, enums.DoublePawnPush))
			}
		}
	}
	leftCapture, rightCapture := pawn&notA>>9, pawn&notH>>7
	if leftCapture&enemies != 0 {
		moves = append(moves, NewMove(GetLSB(leftCapture), from, enums.Capture))
	}
	if rightCapture&enemies != 0 {
		moves = append(moves, NewMove(GetLSB(rightCapture), from, enums.Capture))
	}
	return
}

///////////////////////////////////////////////////////////////
//                          KNIGHT                           //
///////////////////////////////////////////////////////////////

func genKnightsMovePattern(knights uint64) (moves uint64) {
	moves |= (knights & notA) >> 17  // South south west. (-17 squares)
	moves |= (knights & notH) >> 15  // South south east. (-15 squares)
	moves |= (knights & notAB) >> 10 // South west west. (-10 squares)
	moves |= (knights & notGH) >> 6  // South east east. (-6 squares)
	moves |= (knights & notAB) << 6  // Noth west west. (+6 squares)
	moves |= (knights & notGH) << 10 // North east east. (+10 squares)
	moves |= (knights & notA) << 15  // North north west. (+15 squares)
	moves |= (knights & notH) << 17  // North north east (+17 squares)
	return moves
}

func genKnightPseudoLegalMoves(from int, allies, enemies uint64) (moves []Move) {
	movesBB := genKnightsMovePattern(1<<from) & ^allies
	for i := GetLSB(movesBB); movesBB != 0; i = GetLSB(movesBB) {
		if 1<<i&enemies != 0 {
			moves = append(moves, NewMove(i, from, enums.Capture))
		} else {
			moves = append(moves, NewMove(i, from, enums.Quiet))
		}
		movesBB &= movesBB - 1
	}
	return
}

///////////////////////////////////////////////////////////////
//                          BISHOP                           //
///////////////////////////////////////////////////////////////

func genBishopsMovePattern(bishops, occupied uint64) (moves uint64) {
	// North west diagonal. (+7 squares)
	for i := (bishops & notA) << 7; i != 0; i = (i & notA) << 7 {
		moves |= i
		if occupied&i != 0 {
			break
		}
	}
	// South west diagonal. (-9 squares)
	for i := (bishops & notA) >> 9; i != 0; i = (i & notA) >> 9 {
		moves |= i
		if occupied&i != 0 {
			break
		}
	}
	// North east diagonal. (+9 squares)
	for i := (bishops & notH) << 9; i != 0; i = (i & notH) << 9 {
		moves |= i
		if occupied&i != 0 {
			break
		}
	}
	// South east diagonal. (-7 squares)
	for i := (bishops & notH) >> 7; i != 0; i = (i & notH) >> 7 {
		moves |= i
		if occupied&i != 0 {
			break
		}
	}
	return moves
}

func genBishopPseudoLegalMoves(from int, allies, enemies uint64) (moves []Move) {
	movesBB := genBishopsMovePattern(1<<from, allies|enemies) & ^allies
	for i := GetLSB(movesBB); movesBB != 0; i = GetLSB(movesBB) {
		if 1<<i&enemies != 0 {
			moves = append(moves, NewMove(i, from, enums.Capture))
		} else {
			moves = append(moves, NewMove(i, from, enums.Quiet))
		}
		movesBB &= movesBB - 1
	}
	return
}

///////////////////////////////////////////////////////////////
//                          ROOK                             //
///////////////////////////////////////////////////////////////

func genRooksMovePattern(rooks, occupied uint64) (moves uint64) {
	// West horizontal. (-1 square)
	for i := (rooks & notA) >> 1; i != 0; i = (i & notA) >> 1 {
		moves |= i
		if occupied&i != 0 {
			break
		}
	}
	// East horizontal. (+1 square)
	for i := (rooks & notH) << 1; i != 0; i = (i & notH) << 1 {
		moves |= i
		if occupied&i != 0 {
			break
		}
	}
	// North vertical. (+8 squares)
	for i := rooks >> 8; i != 0; i >>= 8 {
		moves |= i
		if occupied&i != 0 {
			break
		}
	}
	// South vertical. (-8 squares)
	for i := rooks << 8; i != 0; i <<= 8 {
		moves |= i
		if occupied&i != 0 {
			break
		}
	}
	return moves
}

func genRookPseudoLegalMoves(from int, allies, enemies uint64) (moves []Move) {
	movesBB := genRooksMovePattern(1<<from, allies|enemies) & ^allies
	for i := GetLSB(movesBB); movesBB != 0; i = GetLSB(movesBB) {
		if 1<<i&enemies != 0 {
			moves = append(moves, NewMove(i, from, enums.Capture))
		} else {
			moves = append(moves, NewMove(i, from, enums.Quiet))
		}
		movesBB &= movesBB - 1
	}
	return
}

///////////////////////////////////////////////////////////////
//                           QUEEN                           //
///////////////////////////////////////////////////////////////

// genQueensMovePattern simultaneously calculates all squares the queens can move to.
func genQueensMovePattern(queens, occupied uint64) uint64 {
	return genBishopsMovePattern(queens, occupied) |
		genRooksMovePattern(queens, occupied)
}

func genQueenPseudoLegalMoves(from int, allies, enemies uint64) (moves []Move) {
	movesBB := genQueensMovePattern(1<<from, allies|enemies) & ^allies
	for i := GetLSB(movesBB); movesBB != 0; i = GetLSB(movesBB) {
		if 1<<i&enemies != 0 {
			moves = append(moves, NewMove(i, from, enums.Capture))
		} else {
			moves = append(moves, NewMove(i, from, enums.Quiet))
		}
		movesBB &= movesBB - 1
	}
	return
}

///////////////////////////////////////////////////////////////
//                           GENERAL                         //
///////////////////////////////////////////////////////////////

func GenAttackedSquares(piece, occupied uint64, pt enums.PieceType) uint64 {
	switch pt {
	case enums.WhitePawn:
		return genWhitePawnsAttackPattern(piece)
	case enums.BlackPawn:
		return genBlackPawnsAttackPattern(piece)
	case enums.WhiteKnight, enums.BlackKnight:
		return genKnightsMovePattern(piece)
	case enums.WhiteBishop, enums.BlackBishop:
		return genBishopsMovePattern(piece, occupied)
	case enums.WhiteRook, enums.BlackRook:
		return genRooksMovePattern(piece, occupied)
	case enums.WhiteQueen, enums.BlackQueen:
		return genQueensMovePattern(piece, occupied)
	default:
		return 0
	}
}

func genAttackedSquaresBySide(pieces [6]uint64,
	occupied uint64, c enums.Color) (attacked uint64) {
	if c == enums.White {
		attacked |= genWhitePawnsAttackPattern(pieces[0])
	} else {
		attacked |= genBlackPawnsAttackPattern(pieces[0])
	}
	attacked |= genKnightsMovePattern(pieces[1])
	attacked |= genBishopsMovePattern(pieces[2], occupied)
	attacked |= genRooksMovePattern(pieces[3], occupied)
	attacked |= genQueensMovePattern(pieces[4], occupied)
	attacked |= genKingMovesPattern(pieces[5])
	return
}

func genPseudoLegalMoves(pt enums.PieceType, from int,
	allies, enemies uint64) []Move {
	switch pt {
	case enums.WhitePawn:
		return genWhitePawnPseudoLegalMoves(from, allies, enemies)
	case enums.BlackPawn:
		return genBlackPawnPseudoLegalMoves(from, allies, enemies)
	case enums.WhiteKnight, enums.BlackKnight:
		return genKnightPseudoLegalMoves(from, allies, enemies)
	case enums.WhiteBishop, enums.BlackBishop:
		return genBishopPseudoLegalMoves(from, allies, enemies)
	case enums.WhiteRook, enums.BlackRook:
		return genRookPseudoLegalMoves(from, allies, enemies)
	case enums.WhiteQueen, enums.BlackQueen:
		return genQueenPseudoLegalMoves(from, allies, enemies)
	}
	panic("incorrect piece type")
}
