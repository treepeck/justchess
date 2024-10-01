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

func (k *Knight) Move(pieces map[helpers.Pos]Piece, to helpers.Pos) bool {
	availibleMoves := k.GetAvailibleMoves(pieces)
	for _, pos := range availibleMoves {
		if to.File == pos.File && to.Rank == pos.Rank {
			// move the knight
			pieces[k.Pos] = nil
			pieces[to] = k

			k.Pos = to
			return true
		}
	}

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

func (k *Knight) GetAvailibleMoves(pieces map[helpers.Pos]Piece) []helpers.Pos {
	possibleMoves := []helpers.Pos{
		{File: k.Pos.File + 2, Rank: k.Pos.Rank + 1},
		{File: k.Pos.File + 2, Rank: k.Pos.Rank - 1},
		{File: k.Pos.File - 2, Rank: k.Pos.Rank + 1},
		{File: k.Pos.File - 2, Rank: k.Pos.Rank - 1},
		{File: k.Pos.File - 1, Rank: k.Pos.Rank + 2},
		{File: k.Pos.File + 1, Rank: k.Pos.Rank - 2},
		{File: k.Pos.File - 1, Rank: k.Pos.Rank - 2},
		{File: k.Pos.File + 1, Rank: k.Pos.Rank + 2},
	}

	availibleMoves := make([]helpers.Pos, 0)
	for _, move := range possibleMoves {
		if move.IsInBoard() {
			// if the square is empty or if the enemy piece
			if pieces[move] == nil || pieces[move].GetColor() != k.Color {
				availibleMoves = append(availibleMoves, move)
			}
		}
	}

	return availibleMoves
}
