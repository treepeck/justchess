package pieces

import (
	"chess-api/models/enums"
	"chess-api/models/helpers"
)

type Bishop struct {
	Color      enums.Color `json:"color"`
	IsCaptured bool        `json:"isCaptured"`
	Pos        helpers.Pos `json:"pos"`
	Name       enums.Piece `json:"name"`
}

func NewBishop(color enums.Color, pos helpers.Pos) *Bishop {
	return &Bishop{
		Color:      color,
		IsCaptured: false,
		Pos:        pos,
		Name:       enums.Bishop,
	}
}

func (b *Bishop) Move(pieces map[helpers.Pos]Piece, to helpers.Pos) bool {
	availibleMoves := b.GetAvailibleMoves(pieces)
	for _, pos := range availibleMoves {
		if to.File == pos.File && to.Rank == pos.Rank {
			// move the bishop
			pieces[b.Pos] = nil
			pieces[to] = b
			b.Pos = to

			return true
		}
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

func (b *Bishop) GetAvailibleMoves(pieces map[helpers.Pos]Piece) []helpers.Pos {
	availibleMoves := make([]helpers.Pos, 0)

	rank := b.Pos.Rank
	for i := b.Pos.File - 1; i >= 1; i-- {
		nextMove := helpers.NewPos(i, rank+1)
		rank++

		p := pieces[nextMove]
		if p == nil {
			availibleMoves = append(availibleMoves, nextMove)
			continue
		} else if p.GetColor() != b.Color {
			availibleMoves = append(availibleMoves, nextMove)
		}
		break
	}

	rank = b.Pos.Rank
	for i := b.Pos.File - 1; i >= 1; i-- {
		nextMove := helpers.NewPos(i, rank-1)
		rank--

		p := pieces[nextMove]
		if p == nil {
			availibleMoves = append(availibleMoves, nextMove)
			continue
		} else if p.GetColor() != b.Color {
			availibleMoves = append(availibleMoves, nextMove)
		}
		break
	}

	rank = b.Pos.Rank
	for i := b.Pos.File + 1; i <= 8; i++ {
		nextMove := helpers.NewPos(i, rank+1)
		rank++

		p := pieces[nextMove]
		if p == nil {
			availibleMoves = append(availibleMoves, nextMove)
			continue
		} else if p.GetColor() != b.Color {
			availibleMoves = append(availibleMoves, nextMove)
		}
		break
	}

	rank = b.Pos.Rank
	for i := b.Pos.File + 1; i <= 8; i++ {
		nextMove := helpers.NewPos(i, rank-1)
		rank--

		p := pieces[nextMove]
		if p == nil {
			availibleMoves = append(availibleMoves, nextMove)
			continue
		} else if p.GetColor() != b.Color {
			availibleMoves = append(availibleMoves, nextMove)
		}
		break
	}

	return availibleMoves
}
