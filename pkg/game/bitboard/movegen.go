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

func genKingAttackedDests(king uint64) (moves uint64) {
	moves = (king & notH) >> 9  // South east. (-9 squares)
	moves |= king >> 8          // South south. (-8 squares)
	moves |= (king & notA) >> 7 // South west. (-7 squares)
	moves |= (king & notA) >> 1 // West. (-1 square)
	moves |= (king & notH) << 1 // East. (+1 square)
	moves |= (king & notA) << 7 // North west. (+7 squares)
	moves |= king << 8          // North north. (+8 squares)
	moves |= (king & notH) << 9 // North east. (+9 squares)
	return
}

func genKingLegalMoves(king, allies, enemies, attacked uint64,
	canOO, canOOO bool, c enums.Color) (moves []Move) {
	kingPos := GetLSB(king)

	// Exclude all attacked and occupied by the allied pieces squares, the king can not move on them.
	movesBB := genKingAttackedDests(king) & ^allies & ^attacked

	for ; movesBB > 0; movesBB &= movesBB - 1 {
		to := GetLSB(movesBB)
		if (1<<to)&enemies != 0 {
			moves = append(moves, NewMove(to, kingPos, enums.Capture))
		} else {
			moves = append(moves, NewMove(to, kingPos, enums.Quiet))
		}
	}

	// Castling implementation.
	if c == enums.White {
		if canOO && (0x60&allies == 0) && (0x70&attacked == 0) {
			moves = append(moves, NewMove(enums.G1, kingPos, enums.KingCastle))
		}
		if canOOO && (0xE&allies == 0) && (0x1E&attacked == 0) {
			moves = append(moves, NewMove(enums.C1, kingPos, enums.QueenCastle))
		}
	} else {
		if canOO && (0x6000000000000000&allies == 0) && (0x7000000000000000&attacked == 0) {
			moves = append(moves, NewMove(enums.G8, kingPos, enums.KingCastle))
		}
		if canOOO && (0xE00000000000000&allies == 0) && (0x1E00000000000000&attacked == 0) {
			moves = append(moves, NewMove(enums.C8, kingPos, enums.QueenCastle))
		}
	}
	return
}

///////////////////////////////////////////////////////////////
//                          PAWN                             //
///////////////////////////////////////////////////////////////

// genPawnsAttackDets simultaneously calculates all attacked by pawns squares on an empty board.
func genPawnsAttackDests(pawns uint64, c enums.Color) uint64 {
	if c == enums.White {
		return ((pawns & notA) << 7) | ((pawns & notH) << 9)
	} else {
		return ((pawns & notA) >> 9) | ((pawns & notH) >> 7)
	}
}

// genPawnsPseudoLegalMoves sequentially calculates pseudo legal moves taking piece placement into account.
func genPawnsPseudoLegalMoves(pawns, allies, enemies uint64, c enums.Color,
	epTarget int) (moves []Move) {
	occupied := allies | enemies

	for ; pawns > 0; pawns &= pawns - 1 {
		pawnFrom := GetLSB(pawns)
		north, doubleNorth := 0, 0
		west, east, pawnBB := uint64(0), uint64(0), uint64(1)<<pawnFrom
		canDoublePush := false

		if c == enums.White {
			north, doubleNorth = pawnFrom+8, pawnFrom+16
			west, east = pawnBB&notA<<7, pawnBB&notH<<9
			canDoublePush = pawnBB&0xFF00 != 0
		} else {
			north, doubleNorth = pawnFrom-8, pawnFrom-16
			west, east = pawnBB&notA>>9, pawnBB&notH>>7
			canDoublePush = pawnBB&0xFF000000000000 != 0
		}

		if (1<<north)&occupied == 0 {
			if (1<<north)&0xFF != 0 || uint64(1<<north)&0xFF00000000000000 != 0 {
				moves = append(moves, NewMove(north, pawnFrom, enums.QueenPromo))
			} else {
				moves = append(moves, NewMove(north, pawnFrom, enums.Quiet))
			}

			if canDoublePush && (1<<doubleNorth)&occupied == 0 {
				moves = append(moves, NewMove(doubleNorth, pawnFrom, enums.DoublePawnPush))
			}
		}

		if west&enemies != 0 {
			if west&0xFF != 0 || west&0xFF00000000000000 != 0 {
				moves = append(moves, NewMove(GetLSB(west), pawnFrom, enums.QueenPromoCapture))
			} else {
				moves = append(moves, NewMove(GetLSB(west), pawnFrom, enums.Capture))
			}
		} else if epTarget != enums.NoSquare && west&(1<<epTarget) != 0 {
			moves = append(moves, NewMove(GetLSB(west), pawnFrom, enums.EPCapture))
		}
		if east&enemies != 0 {
			if east&0xFF != 0 || east&0xFF00000000000000 != 0 {
				moves = append(moves, NewMove(GetLSB(east), pawnFrom, enums.QueenPromoCapture))
			} else {
				moves = append(moves, NewMove(GetLSB(east), pawnFrom, enums.Capture))
			}
		} else if epTarget != enums.NoSquare && east&(1<<epTarget) != 0 {
			moves = append(moves, NewMove(GetLSB(east), pawnFrom, enums.EPCapture))
		}
	}
	return
}

///////////////////////////////////////////////////////////////
//                          KNIGHT                           //
///////////////////////////////////////////////////////////////

// genKnightsAttackDests simultaneously calculates all attacked by knights squares on an empty board.
func genKnightsAttackDests(knights uint64) (moves uint64) {
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

///////////////////////////////////////////////////////////////
//                          BISHOP                           //
///////////////////////////////////////////////////////////////

func genBishopsAttackDests(bishops, occupied uint64) (moves uint64) {
	// South west diagonal. (-9 squares)
	for i := (bishops & notA) >> 9; i != 0; i = (i & notA) >> 9 {
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
	// North west diagonal. (+7 squares)
	for i := (bishops & notA) << 7; i != 0; i = (i & notA) << 7 {
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
	return moves
}

///////////////////////////////////////////////////////////////
//                          ROOK                             //
///////////////////////////////////////////////////////////////

func genRooksAttackDests(rooks, occupied uint64) (moves uint64) {
	// South vertical. (-8 squares)
	for i := rooks << 8; i != 0; i <<= 8 {
		moves |= i
		if occupied&i != 0 {
			break
		}
	}
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
	return moves
}

///////////////////////////////////////////////////////////////
//                           QUEEN                           //
///////////////////////////////////////////////////////////////

func genQueensAttackedDests(queens, occupied uint64) uint64 {
	return genBishopsAttackDests(queens, occupied) |
		genRooksAttackDests(queens, occupied)
}

///////////////////////////////////////////////////////////////
//                           GENERAL                         //
///////////////////////////////////////////////////////////////

// TODO: make it more performant.
// genPseudoLegalMoves sequentially generates all pseudo-legal moves for a given piece type.
func genPseudoLegalMoves(pt enums.PieceType, bb, allies, enemies uint64) (moves []Move) {
	occupied := allies | enemies
	for ; bb > 0; bb &= bb - 1 {
		piecePos := GetLSB(bb)
		from := uint64(1) << piecePos

		var movesBB uint64
		switch pt {
		case enums.WhiteKnight, enums.BlackKnight:
			movesBB = genKnightsAttackDests(from)

		case enums.WhiteBishop, enums.BlackBishop:
			movesBB = genBishopsAttackDests(from, occupied)

		case enums.WhiteRook, enums.BlackRook:
			movesBB = genRooksAttackDests(from, occupied)

		case enums.WhiteQueen, enums.BlackQueen:
			// Queen moves is just a concatenation of the rook`s and bishop`s moves.
			movesBB = genRooksAttackDests(from, occupied) | genBishopsAttackDests(from, occupied)

		default:
			return
		}

		// Skip occupied by the allies pieces squares.
		movesBB &= ^allies

		for ; movesBB > 0; movesBB &= movesBB - 1 {
			to := GetLSB(movesBB)
			if 1<<to&enemies != 0 {
				moves = append(moves, NewMove(to, piecePos, enums.Capture))
			} else {
				moves = append(moves, NewMove(to, piecePos, enums.Quiet))
			}
		}
	}
	return
}

func GenAttackedSquares(pieces [12]uint64, c enums.Color) uint64 {
	var occupied uint64
	for _, pieceBB := range pieces {
		occupied |= pieceBB
	}
	// Get all attacked squares on a new position.
	return genPawnsAttackDests(pieces[0+c], c) |
		genKnightsAttackDests(pieces[2+c]) |
		genBishopsAttackDests(pieces[4+c], occupied) |
		genRooksAttackDests(pieces[6+c], occupied) |
		genQueensAttackedDests(pieces[8+c], occupied) |
		genKingAttackedDests(pieces[10+c])
}
