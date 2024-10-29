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

func (k *King) GetPossibleMoves(pieces map[helpers.Pos]Piece,
) map[helpers.Pos]enums.MoveType {
	// calculate all posible moves for enemy pieces
	// to prevent moving under attacked squares.
	// map is used to store the unique moves only.
	is := getInaccessibleSquares(pieces, k.Color)

	possibleMoves := make(map[helpers.Pos]enums.MoveType)
	// checkSquare is a nested function that checks is the specified square
	// vacant.
	checkSquare := func(dF, dR int) { // delta file, delta rank.
		file, rank := k.Pos.File+dF, k.Pos.Rank+dR

		pos := helpers.NewPos(file, rank)
		if is[pos] == 0 { // vacant square, is not under enemy attack.
			if pos.IsInBoard() {
				p := pieces[pos]
				if p == nil || p.GetColor() != k.Color {
					possibleMoves[pos] = enums.Basic
				} else {
					possibleMoves[pos] = enums.Defend
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

	// if the king is checked, it can not castle
	if is[k.Pos] != 0 {
		return possibleMoves
	}

	// handleCastling is a nested function that checks if a king can castle.
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
			if pieces[pos] != nil || is[pos] != 0 {
				return
			}
		}

		// check the rook
		r := pieces[rookPos]
		if r != nil && r.GetType() == enums.Rook &&
			r.GetMovesCounter() == 0 {
			possibleMoves[helpers.NewPos(k.Pos.File+(2*dF), k.Pos.Rank)] = ct
		}
	}
	handleCastling(enums.ShortCastling, 2, 1)
	handleCastling(enums.LongCastling, 3, -1)
	return possibleMoves
}

func getInaccessibleSquares(pieces map[helpers.Pos]Piece, side enums.Color,
) map[helpers.Pos]enums.MoveType {
	is := make(map[helpers.Pos]enums.MoveType)

	for _, piece := range pieces {
		if piece.GetColor() != side {
			switch piece.GetType() {
			// pawn moves are processed separately since pawns
			// cannot attack front squares.
			case enums.Pawn:
				pm := piece.GetPossibleMoves(pieces)
				for pos, moveType := range pm {
					if moveType != enums.PawnForward {
						is[pos] = moveType
					}
				}
			// piece.GetPossibleMoves cannot be called here, otherwise enless loop will occur:
			// king.GetPossibleMoves -> enemyKing.GetPossibleMoves -> ...
			case enums.King:
				getEnemyKingPossibleMoves(piece.(*King), is)
			// all other pieces.
			default:
				pm := piece.GetPossibleMoves(pieces)
				for pos, moveType := range pm {
					is[pos] = moveType
				}
			}
		}
	}
	return is
}

func getEnemyKingPossibleMoves(k *King, pm map[helpers.Pos]enums.MoveType) {
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

	for _, pos := range possiblePositions {
		if pos.IsInBoard() {
			pm[pos] = enums.Basic
		}
	}
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
