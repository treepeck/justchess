package pieces

import (
	"chess-api/models/game/enums"
	"chess-api/models/game/helpers"
)

type King struct {
	Color        enums.Color     `json:"color"`
	MovesCounter uint            `json:"-"`
	Pos          helpers.Pos     `json:"-"`
	Type         enums.PieceType `json:"type"`
}

func NewKing(color enums.Color, pos helpers.Pos) *King {
	return &King{
		Color:        color,
		MovesCounter: 0,
		Pos:          pos,
		Type:         enums.King,
	}
}

// King`s GetPossibleMoves returns valid moves.
// Thus, there is no need to additionaly check king`s possible moves.
func (k *King) GetPossibleMoves(pieces map[helpers.Pos]Piece,
) []helpers.PossibleMove {
	is := getInaccessibleSquares(pieces, k.Color)

	pm := make([]helpers.PossibleMove, 0)
	// checkSquare checks is the specified square vacant.
	checkSquare := func(dF, dR int) { // delta file, delta rank.
		file, rank := k.Pos.File+dF, k.Pos.Rank+dR

		pos := helpers.NewPos(file, rank)
		if !is[pos] { // vacant square, is not under enemy attack.
			if pos.IsInBoard() {
				p := pieces[pos]
				if p == nil || p.GetColor() != k.Color {
					pm = append(pm, helpers.NewPM(pos, enums.Basic))
				}
			}
		}
	}
	checkSquare(-1, +1) // upper left square.
	checkSquare(0, +1)  // upper square.
	checkSquare(+1, +1) // upper right square.
	checkSquare(-1, 0)  // left square.
	checkSquare(+1, 0)  // right square.
	checkSquare(-1, -1) // lower left square.
	checkSquare(0, -1)  // lower square.
	checkSquare(+1, -1) // lower right square.

	// the king can not castle in check.
	if is[k.Pos] {
		return pm
	}
	// handleCastling checks is a king can castle.
	handleCastling := func(ct enums.MoveType, ss, dF int) {
		var rookPos helpers.Pos
		if ct == enums.ShortCastling {
			rookPos = helpers.NewPos(k.Pos.File+3, k.Pos.Rank) // 0-0
		} else {
			rookPos = helpers.NewPos(k.Pos.File-4, k.Pos.Rank) // 0-0-0
		}
		for i := 1; i <= ss; i++ {
			pos := helpers.NewPos(k.Pos.File+(i*dF), k.Pos.Rank)
			// if the square is not vacant or under attack.
			if pieces[pos] != nil || is[pos] {
				return
			}
		}
		// check the rook
		r := pieces[rookPos]
		if r != nil && r.GetType() == enums.Rook &&
			r.GetMovesCounter() == 0 {
			finalPos := helpers.NewPos(k.Pos.File+(2*dF), k.Pos.Rank)
			pm = append(pm, helpers.NewPM(finalPos, ct))
		}
	}
	handleCastling(enums.ShortCastling, 2, 1)
	handleCastling(enums.LongCastling, 3, -1)
	return pm
}

// getInaccessibleSquares calculate all posible moves for enemy pieces
// to forbit the king to move under attacked squares. The map as a return type
// is used to store the unique moves only.
func getInaccessibleSquares(pieces map[helpers.Pos]Piece, side enums.Color,
) map[helpers.Pos]bool {
	is := make(map[helpers.Pos]bool)

	for _, piece := range pieces {
		if piece.GetColor() != side {
			switch piece.GetType() {
			// pawn moves are processed separately since pawns
			// cannot attack front squares.
			case enums.Pawn:
				pm := piece.GetPossibleMoves(pieces)
				for _, m := range pm {
					if m.MoveType != enums.PawnForward {
						is[m.To] = true
					}
				}
			// piece.GetPossibleMoves cannot be called here,
			// otherwise endless loop will occur:
			// king.GetPossibleMoves -> enemyKing.GetPossibleMoves -> ...
			case enums.King:
				for _, pos := range getEnemyKingMovePattern(piece.(*King)) {
					is[pos] = true
				}

			default:
				pm := piece.GetPossibleMoves(pieces)
				for _, m := range pm {
					is[m.To] = true
				}
			}
		}
	}
	return is
}

func getEnemyKingMovePattern(k *King) []helpers.Pos {
	possiblePositions := []helpers.Pos{
		{File: k.Pos.File - 1, Rank: k.Pos.Rank + 1},
		{File: k.Pos.File, Rank: k.Pos.Rank + 1},
		{File: k.Pos.File + 1, Rank: k.Pos.Rank + 1},
		{File: k.Pos.File - 1, Rank: k.Pos.Rank},
		{File: k.Pos.File + 1, Rank: k.Pos.Rank},
		{File: k.Pos.File - 1, Rank: k.Pos.Rank - 1},
		{File: k.Pos.File, Rank: k.Pos.Rank - 1},
		{File: k.Pos.File + 1, Rank: k.Pos.Rank - 1},
	}

	mp := make([]helpers.Pos, 0)
	for _, pos := range possiblePositions {
		if pos.IsInBoard() {
			mp = append(mp, pos)
		}
	}
	return mp
}

func (k *King) Move(to helpers.Pos) {
	k.Pos = to
	k.MovesCounter++
}

func (k *King) GetMovesCounter() uint {
	return k.MovesCounter
}

func (k *King) SetMovesCounter(mc uint) {
	k.MovesCounter = mc
}

func (k *King) GetType() enums.PieceType {
	return enums.King
}

func (k *King) GetColor() enums.Color {
	return k.Color
}

func (k *King) GetPosition() helpers.Pos {
	return k.Pos
}

func (k *King) GetFEN() string {
	if k.Color == enums.White {
		return "K"
	}
	return "k"
}
