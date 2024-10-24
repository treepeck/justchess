package pieces

import (
	"chess-api/models/game/enums"
	"chess-api/models/game/helpers"
)

type Knight struct {
	Color        enums.Color     `json:"color"`
	Pos          helpers.Pos     `json:"pos"`
	Type         enums.PieceType `json:"type"`
	MovesCounter uint            `json:"movesCounter"`
}

func NewKnight(color enums.Color, pos helpers.Pos) *Knight {
	return &Knight{
		Color:        color,
		Pos:          pos,
		Type:         enums.Knight,
		MovesCounter: 0,
	}
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

func (k *Knight) Move(to helpers.Pos) {
	k.Pos = to
}

func (k *Knight) GetType() enums.PieceType {
	return enums.Knight
}

func (k *Knight) GetMovesCounter() uint {
	return k.MovesCounter
}

func (k *Knight) SetMovesCounter(mc uint) {
	k.MovesCounter = mc
}

func (k *Knight) GetColor() enums.Color {
	return k.Color
}

func (k *Knight) GetPosition() helpers.Pos {
	return k.Pos
}
