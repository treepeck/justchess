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

func (r *Rook) Move(pieces map[helpers.Pos]Piece, to helpers.Pos) bool {
	availibleMoves := r.GetAvailibleMoves(pieces)
	for _, pos := range availibleMoves {
		if to.File == pos.File && to.Rank == pos.Rank {
			// move the rook
			pieces[r.Pos] = nil
			pieces[to] = r

			r.Pos = to

			r.MovesCounter++
			return true
		}
	}
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

func (r *Rook) GetAvailibleMoves(pieces map[helpers.Pos]Piece) []helpers.Pos {
	availibleMoves := make([]helpers.Pos, 0)

	// bottom ranks
	for i := r.Pos.Rank - 1; i >= 1; i-- {
		nextMove := helpers.NewPos(r.Pos.File, i)
		if !nextMove.IsInBoard() {
			break
		} else if p := pieces[nextMove]; p != nil {
			if p.GetColor() != r.Color {
				availibleMoves = append(availibleMoves, nextMove)
			}
			break
		}
		availibleMoves = append(availibleMoves, nextMove)
	}

	// upper ranks
	for i := r.Pos.Rank + 1; i <= 8; i++ {
		nextMove := helpers.NewPos(r.Pos.File, i)
		if !nextMove.IsInBoard() {
			break
		} else if p := pieces[nextMove]; p != nil {
			if p.GetColor() != r.Color {
				availibleMoves = append(availibleMoves, nextMove)
			}
			break
		}
		availibleMoves = append(availibleMoves, nextMove)
	}

	// left files
	for i := r.Pos.File - 1; i >= 1; i-- {
		nextMove := helpers.NewPos(i, r.Pos.Rank)
		if !nextMove.IsInBoard() {
			break
		} else if p := pieces[nextMove]; p != nil {
			if p.GetColor() != r.Color {
				availibleMoves = append(availibleMoves, nextMove)
			}
			break
		}
		availibleMoves = append(availibleMoves, nextMove)
	}

	// right files
	for i := r.Pos.File + 1; i <= 8; i++ {
		nextMove := helpers.NewPos(i, r.Pos.Rank)
		if !nextMove.IsInBoard() {
			break
		} else if p := pieces[nextMove]; p != nil {
			if p.GetColor() != r.Color {
				availibleMoves = append(availibleMoves, nextMove)
			}
			break
		}
		availibleMoves = append(availibleMoves, nextMove)
	}

	return availibleMoves
}
