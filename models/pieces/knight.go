package pieces

import (
	"chess-api/models/enums"
	"chess-api/models/helpers"
)

type Knight struct {
	Color      enums.Color `json:"color"`
	IsCaptured bool        `json:"isCaptured"`
	Pos        helpers.Pos `json:"pos"`
	Name       enums.Piece `json:"name"`
}

func NewKnight(color enums.Color, pos helpers.Pos) *Knight {
	return &Knight{
		Color:      color,
		IsCaptured: false,
		Pos:        pos,
		Name:       enums.Knight,
	}
}

func (k *Knight) Move(pieces map[helpers.Pos]Piece, move *helpers.Move) bool {
	// possibleMoves := k.GetPossibleMoves(pieces)
	// for _, pm := range possibleMoves {
	// 	if pm.To.IsEqual(move.To) && pm.MoveType != enums.Defend {
	// 		// move the knight
	// 		pieces[k.Pos] = nil
	// 		pieces[pm.To] = k

	// 		k.Pos = pm.To
	// 		return true
	// 	}
	// }

	return false
}

func (k *Knight) GetName() enums.Piece {
	return enums.Knight
}

func (k *Knight) GetColor() enums.Color {
	return k.Color
}

func (k *Knight) GetPosition() helpers.Pos {
	return k.Pos
}

func (k *Knight) GetPossibleMoves(pieces map[helpers.Pos]Piece,
) map[helpers.Pos]enums.MoveType {
	possiblePos := []helpers.Pos{
		{File: k.Pos.File + 2, Rank: k.Pos.Rank + 1},
		{File: k.Pos.File + 2, Rank: k.Pos.Rank - 1},
		{File: k.Pos.File - 2, Rank: k.Pos.Rank + 1},
		{File: k.Pos.File - 2, Rank: k.Pos.Rank - 1},
		{File: k.Pos.File - 1, Rank: k.Pos.Rank + 2},
		{File: k.Pos.File + 1, Rank: k.Pos.Rank - 2},
		{File: k.Pos.File - 1, Rank: k.Pos.Rank - 2},
		{File: k.Pos.File + 1, Rank: k.Pos.Rank + 2},
	}

	possibleMoves := make(map[helpers.Pos]enums.MoveType)
	for _, pos := range possiblePos {
		if pos.IsInBoard() {
			// if the square is empty or if the enemy piece
			piece := pieces[pos]
			if piece == nil {
				possibleMoves[pos] = enums.Basic
			} else if piece.GetColor() != k.Color {
				possibleMoves[pos] = enums.Basic
			} else if piece.GetColor() == k.Color {
				possibleMoves[pos] = enums.Defend
			}
		}
	}

	return possibleMoves
}
