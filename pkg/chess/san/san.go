package san

import (
	"justchess/pkg/chess/bitboard"
	"justchess/pkg/chess/enums"
	"log"
)

var files = [8]string{"a", "b", "c", "d", "e", "f", "g", "h"}

// Move2SAN encodes the m to its SAN representation.
// Basic moves consists of these parts:
//
//  1. Piece name (see [enums.PieceType]). Empty for pawns;
//  2. Originating (from) file or rank. Optional, used only for disambiguation.
//     If a pawn performs a capture, its originating file is always included;
//  3. Denotation of capture by 'x'. Mandatory for capture moves;
//  4. Destination (to) file and rank;
//  5. Denotation of check by '+'. Omitted when the move is a checkmate;
//  6. Denotation of checkmate by '#'.
//
// King castling and queen castling are encoded as "O-O" and "O-O-O" respectively.
func Move2SAN(m bitboard.Move, pieces [12]uint64, legalMoves []bitboard.Move,
	pt enums.PieceType, isCheck, isMate bool) (sanStr string) {

	if m.Type() == enums.KingCastle {
		return "O-O"
	} else if m.Type() == enums.QueenCastle {
		return "O-O-O"
	}

	sanStr = pt.String()

	// Check if there is an ambiguity.
	if pt > enums.BlackPawn {
		for _, lm := range legalMoves {
			if bitboard.GetPieceOnSquare(1<<lm.From(), pieces) == pt &&
				m.From() != lm.From() && m.To() == lm.To() {
				sanStr += disambiguate(m.From(), lm.From())
				break
			}
		}
	}

	if m.Type() == enums.Capture || m.Type() == enums.EPCapture ||
		m.Type() >= enums.KnightPromoCapture {
		if pt < enums.WhiteKnight {
			sanStr += files[m.From()%8]
		}
		sanStr += "x"
	}

	sanStr += files[m.To()%8] + string(rune(m.To()/8+1+'0'))

	sanStr += formatPromo(m.Type())

	if isCheck && isMate {
		sanStr += "#"
	} else if isCheck {
		sanStr += "+"
	}
	return
}

// disambiguate resolves the ambiguity that arises when multiple pieces of
// the same type can move to the same square.
// Steps to resolve:
//
//  1. If the moving pieces can be distinguished by their originating files,
//     the originating file letter of the moving piece is inserted immediately after
//     the moving piece letter;
//  2. If the moving pieces can be distinguished by their originating ranks,
//     the originating rank digit of the moving piece is inserted immediately after
//     the moving piece letter.
func disambiguate(fromA, fromB int) string {
	if fromA%8 != fromB%8 {
		return files[fromA%8]
	}
	if fromA/8 != fromB/8 {
		return string(rune(fromA/8 + 1 + '0'))
	}
	log.Printf("ERROR: cannot disambiguate move")
	return ""
}

func formatPromo(mt enums.MoveType) string {
	switch mt {
	case enums.KnightPromo, enums.KnightPromoCapture:
		return "=N"
	case enums.BishopPromo, enums.BishopPromoCapture:
		return "=B"
	case enums.RookPromo, enums.RookPromoCapture:
		return "=R"
	case enums.QueenPromo, enums.QueenPromoCapture:
		return "=Q"
	default:
		return ""
	}
}
