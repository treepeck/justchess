package bitboard

import (
	"justchess/pkg/game/enums"
	"justchess/pkg/game/helpers"
	"math/bits"
)

// When generating moves, pieces that blocks the direction are taken into account.
// Occupied squares (by allies and enemies) considered as attacked squares to prevent
// the king from capturing the defended pieces.
// The moves returned by gen*PseudoLegalMoves functions must be checked further to became
// legal, since they can expose the allied king to check or did not cover the checked king.

///////////////////////////////////////////////////////////////
//                          KING                             //
///////////////////////////////////////////////////////////////

func genKingMovesPattern(king uint64) uint64 {
	var moves uint64
	moves |= (king & notA) << 7 // North west. (+7 squares)
	moves |= king << 8          // North north. (+8 squares)
	moves |= (king & notH) << 9 // North east. (+9 squares)
	moves |= (king & notA) >> 1 // West. (-1 square)
	moves |= (king & notH) << 1 // East. (+1 square)
	moves |= (king & notA) >> 7 // South west. (-7 squares)
	moves |= king >> 8          // South south. (-8 squares)
	moves |= (king & notH) >> 9 // South east. (-9 squares)
	return moves
}

func genKingLegalMoves(from int, allies, enemies, attacked uint64,
	can00, can000 bool) []helpers.Move {
	moves := genKingMovesPattern(uint64(1)<<from) & ^allies & ^attacked
	legalMoves := make([]helpers.Move, 0)
	for _, i := range helpers.GetIndicesFromBitboard(moves) {
		if uint64(1)<<i&enemies != 0 {
			legalMoves = append(legalMoves, helpers.NewMove(i, from, enums.Capture, enums.King))
		} else {
			legalMoves = append(legalMoves, helpers.NewMove(i, from, enums.Quiet, enums.King))
		}
	}
	// Castling implementation.
	if can000 && (0xE&allies == 0) && (0xE&attacked == 0) {
		legalMoves = append(legalMoves, helpers.NewMove(enums.B1, from, enums.QueenCastle, enums.King))
	}
	if can00 && (0x60&allies == 0) && (0x60&attacked == 0) {
		legalMoves = append(legalMoves, helpers.NewMove(enums.G1, from, enums.KingCastle, enums.King))
	}
	return legalMoves
}

///////////////////////////////////////////////////////////////
//                          PAWN                             //
///////////////////////////////////////////////////////////////

// TODO: Implement en passant.

// genWhitePawnsAttackPattern returns bitboard with attacked squares by white pawns.
func genWhitePawnsAttackPattern(pawns uint64) uint64 {
	return ((pawns & notA) << 7) | ((pawns & notH) << 9)
}

// genBlackPawnsAttackPattern returns bitboard with possible attacked squares by black pawns.
func genBlackPawnsAttackPattern(pawns uint64) uint64 {
	return ((pawns & notA) >> 9) | ((pawns & notH) >> 7)
}

func genWhitePawnPseudoLegalMoves(from int, allies, enemies uint64) []helpers.Move {
	moves := make([]helpers.Move, 0)
	pawn := uint64(1) << from
	if (allies|enemies)&(pawn<<8) == 0 {
		if (pawn<<8)&0xFF00000000000000 != 0 { // If it is 8th rank.
			moves = append(moves, helpers.NewMove(bits.TrailingZeros64(pawn<<8),
				from, enums.Promotion, enums.Pawn))
		} else {
			moves = append(moves, helpers.NewMove(bits.TrailingZeros64(pawn<<8),
				from, enums.Quiet, enums.Pawn))
			// Pawns can perform double forward push from an initial position.
			if pawn&0xFF00 != 0 && (allies|enemies)&(pawn<<16) == 0 {
				moves = append(moves, helpers.NewMove(bits.TrailingZeros64(pawn<<16),
					from, enums.DoublePawnPush, enums.Pawn))
			}
		}
	}
	cm := genWhitePawnsAttackPattern(pawn) // Capture moves.
	return append(moves, helpers.GetMovesFromBitboard(from, (cm&enemies), enemies, enums.Pawn)...)
}

func genBlackPawnPseudoLegalMoves(from int, allies, enemies uint64) []helpers.Move {
	moves := make([]helpers.Move, 0)
	pawn := uint64(1) << from
	if (allies|enemies)&(pawn>>8) == 0 {
		if (pawn>>8)&0x00000000000000FF != 0 { // If it is 1th rank.
			moves = append(moves, helpers.NewMove(bits.TrailingZeros64(pawn>>8),
				from, enums.Promotion, enums.Pawn))
		} else {
			moves = append(moves, helpers.NewMove(bits.TrailingZeros64(pawn>>8),
				from, enums.Quiet, enums.Pawn))
			// Pawns can perform double backward push from an initial position.
			if pawn&0xFF000000000000 != 0 && (allies|enemies)&(pawn>>16) == 0 {
				moves = append(moves, helpers.NewMove(bits.TrailingZeros64(pawn>>16),
					from, enums.DoublePawnPush, enums.Pawn))
			}
		}
	}
	cm := genBlackPawnsAttackPattern(pawn) // Capture moves.
	return append(moves, helpers.GetMovesFromBitboard(from, (cm&enemies), enemies, enums.Pawn)...)
}

///////////////////////////////////////////////////////////////
//                          KNIGHT                           //
///////////////////////////////////////////////////////////////

func genKnightsMovePattern(knights uint64) uint64 {
	var moves uint64
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

func genKnightPseudoLegalMoves(from int, allies, enemies uint64) []helpers.Move {
	moves := genKnightsMovePattern(uint64(1)<<from) & ^allies
	return helpers.GetMovesFromBitboard(from, moves, moves&enemies, enums.Knight) // Moves & enemies = capture moves.
}

///////////////////////////////////////////////////////////////
//                          BISHOP                           //
///////////////////////////////////////////////////////////////

func genBishopsMovePattern(bishops, occupied uint64) uint64 {
	var moves uint64
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

func genBishopPseudoLegalMoves(from int, allies, enemies uint64) []helpers.Move {
	moves := genBishopsMovePattern(uint64(1)<<from, allies|enemies) & ^allies
	moves &= ^allies // Exclude the squares occupied by allied pieces.
	return helpers.GetMovesFromBitboard(from, moves, enemies, enums.Bishop)
}

///////////////////////////////////////////////////////////////
//                          ROOK                             //
///////////////////////////////////////////////////////////////

func genRooksMovePattern(rooks, occupied uint64) uint64 {
	var moves uint64
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

func genRookPseudoLegalMoves(from int, allies, enemies uint64) []helpers.Move {
	moves := genRooksMovePattern(uint64(1)<<from, allies|enemies)
	moves &= ^allies // Exclude the allied pieces.
	return helpers.GetMovesFromBitboard(from, moves, enemies, enums.Rook)
}

///////////////////////////////////////////////////////////////
//                           QUEEN                           //
///////////////////////////////////////////////////////////////

// genQueensMovePattern simultaneously calculates all squares the queens can move to.
func genQueensMovePattern(queens, occupied uint64) uint64 {
	return genBishopsMovePattern(queens, occupied) |
		genRooksMovePattern(queens, occupied)
}

func genQueenPseudoLegalMoves(from int, allies, enemies uint64) []helpers.Move {
	moves := genQueensMovePattern(uint64(1)<<from, allies|enemies)
	moves &= ^allies
	return helpers.GetMovesFromBitboard(from, moves, enemies, enums.Queen)
}

///////////////////////////////////////////////////////////////
//                           GENERAL                         //
///////////////////////////////////////////////////////////////

func genAttackedSquares(c enums.Color, allies []uint64,
	occupied uint64) (attacked uint64) {
	if c == enums.White {
		attacked |= genWhitePawnsAttackPattern(allies[0])
	} else {
		attacked |= genBlackPawnsAttackPattern(allies[0])
	}
	attacked |= genKnightsMovePattern(allies[1])
	attacked |= genBishopsMovePattern(allies[2], occupied)
	attacked |= genRooksMovePattern(allies[3], occupied)
	attacked |= genQueensMovePattern(allies[4], occupied)
	attacked |= genKingMovesPattern(allies[5])
	return
}

func genPseudoLegalMoves(pt enums.PieceType, c enums.Color,
	from int, allies, enemies uint64) []helpers.Move {
	switch pt {
	case enums.Pawn:
		if c == enums.White {
			return genWhitePawnPseudoLegalMoves(from, allies, enemies)
		}
		return genBlackPawnPseudoLegalMoves(from, allies, enemies)
	case enums.Knight:
		return genKnightPseudoLegalMoves(from, allies, enemies)
	case enums.Bishop:
		return genBishopPseudoLegalMoves(from, allies, enemies)
	case enums.Rook:
		return genRookPseudoLegalMoves(from, allies, enemies)
	case enums.Queen:
		return genQueenPseudoLegalMoves(from, allies, enemies)
	}
	panic("incorrect piece type")
}
