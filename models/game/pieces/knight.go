package pieces

import (
	"chess-api/models/game/enums"
	"chess-api/models/game/helpers"
)

type Knight struct {
	Color        enums.Color     `json:"color"`
	Pos          helpers.Pos     `json:"-"`
	Type         enums.PieceType `json:"type"`
	MovesCounter uint            `json:"-"`
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
) []helpers.PossibleMove {
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

	pm := make([]helpers.PossibleMove, 0)
	for _, pos := range possiblePos {
		if !pos.IsInBoard() {
			continue
		}
		piece := pieces[pos]
		if piece == nil || piece.GetColor() != k.Color {
			pm = append(pm, helpers.NewPM(pos, enums.Basic))
		} else {
			pm = append(pm, helpers.NewPM(pos, enums.Defend))
		}
	}
	return pm
}

func (k *Knight) Move(to helpers.Pos) {
	k.Pos = to
	k.MovesCounter++
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

func (k *Knight) GetFEN() string {
	if k.Color == enums.White {
		return "N"
	}
	return "n"
}
