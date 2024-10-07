package pieces

import (
	"chess-api/models/enums"
	"chess-api/models/helpers"
)

type Rook struct {
	Color        enums.Color `json:"color"`
	IsCaptured   bool        `json:"isCaptured"`
	MovesCounter uint        `json:"movesCounter"`
	Pos          helpers.Pos `json:"pos"`
	Name         enums.Piece `json:"name"`
}

func NewRook(color enums.Color, pos helpers.Pos) *Rook {
	return &Rook{
		Color:        color,
		IsCaptured:   false,
		MovesCounter: 0,
		Pos:          pos,
		Name:         enums.Rook,
	}
}

func (r *Rook) Move(pieces map[helpers.Pos]Piece, move *helpers.Move) bool {
	// possibleMoves := r.GetPossibleMoves(pieces)
	// for _, pm := range possibleMoves {
	// 	if move.To.IsEqual(pm.To) && pm.MoveType != enums.Defend {
	// 		// move the rook
	// 		pieces[r.Pos] = nil
	// 		pieces[pm.To] = r

	// 		r.Pos = pm.To

	// 		r.MovesCounter++
	// 		return true
	// 	}
	// }
	return false
}

func (r *Rook) GetName() enums.Piece {
	return enums.Rook
}

func (r *Rook) GetColor() enums.Color {
	return r.Color
}

func (r *Rook) GetPosition() helpers.Pos {
	return r.Pos
}

func (r *Rook) GetMovesCounter() uint {
	return r.MovesCounter
}

func (r *Rook) GetPossibleMoves(pieces map[helpers.Pos]Piece,
) map[helpers.Pos]enums.MoveType {
	possibleMoves := make(map[helpers.Pos]enums.MoveType)

	// bottom ranks
	for i := r.Pos.Rank - 1; i >= 1; i-- {
		nextPos := helpers.NewPos(r.Pos.File, i)
		if !nextPos.IsInBoard() {
			break
		} else if p := pieces[nextPos]; p != nil {
			if p.GetColor() != r.Color {
				possibleMoves[nextPos] = enums.Basic
			} else {
				possibleMoves[nextPos] = enums.Defend
			}
			break
		}
		possibleMoves[nextPos] = enums.Basic
	}

	// upper ranks
	for i := r.Pos.Rank + 1; i <= 8; i++ {
		nextPos := helpers.NewPos(r.Pos.File, i)
		if !nextPos.IsInBoard() {
			break
		} else if p := pieces[nextPos]; p != nil {
			if p.GetColor() != r.Color {
				possibleMoves[nextPos] = enums.Basic
			} else {
				possibleMoves[nextPos] = enums.Defend
			}
			break
		}
		possibleMoves[nextPos] = enums.Basic
	}

	// left files
	for i := r.Pos.File - 1; i >= 1; i-- {
		nextPos := helpers.NewPos(i, r.Pos.Rank)
		if !nextPos.IsInBoard() {
			break
		} else if p := pieces[nextPos]; p != nil {
			if p.GetColor() != r.Color {
				possibleMoves[nextPos] = enums.Basic
			} else {
				possibleMoves[nextPos] = enums.Defend
			}
			break
		}
		possibleMoves[nextPos] = enums.Basic
	}

	// right files
	for i := r.Pos.File + 1; i <= 8; i++ {
		nextPos := helpers.NewPos(i, r.Pos.Rank)
		if !nextPos.IsInBoard() {
			break
		} else if p := pieces[nextPos]; p != nil {
			if p.GetColor() != r.Color {
				possibleMoves[nextPos] = enums.Basic
			} else {
				possibleMoves[nextPos] = enums.Defend
			}
			break
		}
		possibleMoves[nextPos] = enums.Basic
	}

	return possibleMoves
}
