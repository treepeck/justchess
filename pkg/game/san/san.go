package san

import (
	"justchess/pkg/game/bitboard"
	"justchess/pkg/game/enums"
)

var pieceSymbols = [12]string{"", "", "N", "N", "B", "B", "R", "R", "Q", "Q", "K", "K"}

// Move2SAN converts the move to a Standart Algebraic Notation. Note that checks and
// checkmates are not taken into account and must be added further.
func Move2SAN(m bitboard.Move, pieces [12]uint64,
	lm []bitboard.Move, pt enums.PieceType) (san string) {
	switch m.Type() {
	case enums.KingCastle:
		return "O-O"
	case enums.QueenCastle:
		return "O-O-O"
	case enums.KnightPromo:
		san += square2Str(m.To()) + "=N"
	case enums.BishopPromo:
		san += square2Str(m.To()) + "=B"
	case enums.RookPromo:
		san += square2Str(m.To()) + "=R"
	case enums.QueenPromo:
		san += square2Str(m.To()) + "=Q"
	case enums.KnightPromoCapture:
		san += disambiguate(m.From(), m.To(), pieces, lm, true) + "=N"
	case enums.BishopPromoCapture:
		san += disambiguate(m.From(), m.To(), pieces, lm, true) + "=B"
	case enums.RookPromoCapture:
		san += disambiguate(m.From(), m.To(), pieces, lm, true) + "=R"
	case enums.QueenPromoCapture:
		san += disambiguate(m.From(), m.To(), pieces, lm, true) + "=Q"
	case enums.DoublePawnPush:
		san += square2Str(m.To())
	case enums.Capture, enums.EPCapture:
		san += disambiguate(m.From(), m.To(), pieces, lm, true)
	default:
		san += disambiguate(m.From(), m.To(), pieces, lm, false)
	}
	return pieceSymbols[pt] + san
}

func disambiguate(from, to int, pieces [12]uint64,
	lm []bitboard.Move, isCapture bool) (san string) {
	pt := bitboard.GetPieceTypeFromSquare(from, pieces)
	for _, move := range lm {
		_from, _to := move.From(), move.To()
		if _from != from && _to == to {
			if bitboard.GetPieceTypeFromSquare(move.From(), pieces) == pt {
				_from := move.From()
				if from%8 != _from%8 {
					// Step 1: If the moving pieces can be distinguished by their originating files,
					// the originating file letter of the moving piece is inserted immediately after
					// the moving piece letter.
					san = file2Str(from % 8)
				} else if from/8 != _from/8 {
					// Step 2: If the moving pieces can be distinguished by their originating ranks,
					// the originating rank digit of the moving piece is inserted immediately after
					// the moving piece letter.
					san = string(rune(from/8 + 1 + '0'))
				} else {
					// Step 3: The two character square coordinate of the originating square of the
					// moving piece is inserted immediately after the moving piece letter.
					san = square2Str(from)
				}
				if isCapture {
					san += "x"
				}
				return san + square2Str(to)
			}
		}
	}
	return square2Str(to)
}

func square2Str(square int) string {
	return file2Str(square%8) + string(rune(square/8+1+'0'))
}

func file2Str(file int) string {
	switch file {
	case 0:
		return "a"
	case 1:
		return "b"
	case 2:
		return "c"
	case 3:
		return "d"
	case 4:
		return "e"
	case 5:
		return "f"
	case 6:
		return "g"
	case 7:
		return "h"
	}
	panic("incorrect file")
}
