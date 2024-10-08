package pieces

import (
	"chess-api/models/enums"
	"chess-api/models/helpers"
)

type Bishop struct {
	Color enums.Color `json:"color"`
	Pos   helpers.Pos `json:"pos"`
	Name  enums.Piece `json:"name"`
}

func NewBishop(color enums.Color, pos helpers.Pos) *Bishop {
	return &Bishop{
		Color: color,
		Pos:   pos,
		Name:  enums.Bishop,
	}
}

func (b *Bishop) Move(pieces map[helpers.Pos]Piece, move *helpers.Move) bool {
	possibleMoves := b.GetPossibleMoves(pieces)

	pm := possibleMoves[move.To]
	if pm != 0 && pm != enums.Defend {
		if pieces[move.To] != nil {
			move.IsCapture = true
		}

		delete(pieces, move.From)
		pieces[move.To] = b
		b.Pos = move.To
		return true
	}

	return false
}

func (b *Bishop) GetName() enums.Piece {
	return enums.Bishop
}

func (b *Bishop) GetColor() enums.Color {
	return b.Color
}

func (b *Bishop) GetPosition() helpers.Pos {
	return b.Pos
}

func (b *Bishop) GetPossibleMoves(pieces map[helpers.Pos]Piece,
) map[helpers.Pos]enums.MoveType {
	possibleMoves := make(map[helpers.Pos]enums.MoveType)

	rank := b.Pos.Rank
	for i := b.Pos.File - 1; i >= 1; i-- {
		nextPos := helpers.NewPos(i, rank+1)
		rank++

		if nextPos.IsInBoard() {
			p := pieces[nextPos]
			if p == nil {
				possibleMoves[nextPos] = enums.Basic
				continue
			} else if p.GetColor() != b.Color {
				possibleMoves[nextPos] = enums.Basic
			} else {
				possibleMoves[nextPos] = enums.Defend
			}
		}
		break
	}

	rank = b.Pos.Rank
	for i := b.Pos.File - 1; i >= 1; i-- {
		nextPos := helpers.NewPos(i, rank-1)
		rank--

		if nextPos.IsInBoard() {
			p := pieces[nextPos]
			if p == nil {
				possibleMoves[nextPos] = enums.Basic
				continue
			} else if p.GetColor() != b.Color {
				possibleMoves[nextPos] = enums.Basic
			} else {
				possibleMoves[nextPos] = enums.Defend
			}
		}
		break
	}

	rank = b.Pos.Rank
	for i := b.Pos.File + 1; i <= 8; i++ {
		nextPos := helpers.NewPos(i, rank+1)
		rank++

		if nextPos.IsInBoard() {
			p := pieces[nextPos]
			if p == nil {
				possibleMoves[nextPos] = enums.Basic
				continue
			} else if p.GetColor() != b.Color {
				possibleMoves[nextPos] = enums.Basic
			} else {
				possibleMoves[nextPos] = enums.Defend
			}
		}
		break
	}

	rank = b.Pos.Rank
	for i := b.Pos.File + 1; i <= 8; i++ {
		nextPos := helpers.NewPos(i, rank-1)
		rank--

		if nextPos.IsInBoard() {
			p := pieces[nextPos]
			if p == nil {
				possibleMoves[nextPos] = enums.Basic
				continue
			} else if p.GetColor() != b.Color {
				possibleMoves[nextPos] = enums.Basic
			} else {
				possibleMoves[nextPos] = enums.Defend
			}
		}
		break
	}

	return possibleMoves
}
