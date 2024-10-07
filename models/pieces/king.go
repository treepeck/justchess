package pieces

import (
	"chess-api/models/enums"
	"chess-api/models/helpers"
	"log/slog"
)

type King struct {
	Color        enums.Color `json:"color"`
	IsCaptured   bool        `json:"isCaptured"`
	MovesCounter uint        `json:"movesCounter"`
	Pos          helpers.Pos `json:"pos"`
	Name         enums.Piece `json:"name"`
	IsChecked    bool        `json:"isChecked"`
}

func NewKing(color enums.Color, pos helpers.Pos) *King {
	return &King{
		Color:        color,
		IsCaptured:   false,
		MovesCounter: 0,
		Pos:          pos,
		Name:         enums.King,
	}
}

func (k *King) Move(pieces map[helpers.Pos]Piece, move *helpers.Move) bool {
	slog.Debug("King Move")
	return false
}

func (k *King) GetName() enums.Piece {
	return enums.King
}

func (k *King) GetColor() enums.Color {
	return k.Color
}

func (k *King) GetPosition() helpers.Pos {
	return k.Pos
}

func (k *King) GetMovesCounter() uint {
	return k.MovesCounter
}

func (k *King) GetPossibleMoves(pieces map[helpers.Pos]Piece,
) map[helpers.Pos]enums.MoveType {
	// calculate all posible moves for the enemy pieces
	// to prevent moving under attacked square.
	// map is used to store the unique moves only.
	inaccessibleSquares := make(map[helpers.Pos]enums.MoveType)

	for _, piece := range pieces {
		if piece.GetColor() != k.Color {
			possibleMoves := piece.GetPossibleMoves(pieces)
			for pos, moveType := range possibleMoves {
				inaccessibleSquares[pos] = moveType
			}
		}
	}

	_ = []helpers.Pos{
		{File: k.Pos.File - 1, Rank: k.Pos.Rank + 1},
		{File: k.Pos.File, Rank: k.Pos.Rank + 1},
		{File: k.Pos.File + 1, Rank: k.Pos.Rank + 1},
		{File: k.Pos.File - 1, Rank: k.Pos.Rank},
		{File: k.Pos.File + 1, Rank: k.Pos.Rank},
		{File: k.Pos.File - 1, Rank: k.Pos.Rank - 1},
		{File: k.Pos.File, Rank: k.Pos.Rank - 1},
		{File: k.Pos.File + 1, Rank: k.Pos.Rank - 1},
	}

	// possibleMoves := make([]helpers.PossibleMove, 0)
	// for _, pos := range possiblePositions {
	// 	if !enemyPM[pos] {
	// 		if pos.IsInBoard() {
	// 			p := pieces[pos]
	// 			if p == nil || p.GetColor() != k.Color {
	// 				// TODO: handle special moves.
	// 				possibleMoves = append(possibleMoves,
	// 					helpers.NewPossibleMove(enums.Basic, pos),
	// 				)
	// 			} else {
	// 				possibleMoves = append(possibleMoves,
	// 					helpers.NewPossibleMove(enums.Defend, pos),
	// 				)
	// 			}
	// 		}
	// 	}
	// }

	return make(map[helpers.Pos]enums.MoveType)
}
