package pieces

import (
	"chess-api/models/enums"
	"chess-api/models/helpers"
)

type Pawn struct {
	Color        enums.Color `json:"color"`
	MovesCounter uint        `json:"movesCounter"`
	EnPassant    bool        `json:"enPassant"`
	Pos          helpers.Pos `json:"pos"`
	Name         enums.Piece `json:"name"`
	IsCaptured   bool        `json:"isCaptured"`
}

func NewPawn(color enums.Color, pos helpers.Pos) *Pawn {
	return &Pawn{
		Color:        color,
		MovesCounter: 0,
		Pos:          pos,
		Name:         enums.Pawn,
		IsCaptured:   false,
	}
}

func (p *Pawn) Move(pieces map[helpers.Pos]Piece, to helpers.Pos) bool {
	availibleMoves := p.GetAvailibleMoves(pieces)
	for _, pos := range availibleMoves {
		if to.File == pos.File && to.Rank == pos.Rank {
			// needed to determine en passant moves
			beginRank := p.Pos.Rank
			// move the pawn
			pieces[p.Pos] = nil
			pieces[to] = p

			p.Pos = to
			p.MovesCounter++

			// check if en passant move
			if p.MovesCounter > 1 {
				p.EnPassant = false
			} else if p.MovesCounter == 1 {
				if beginRank-p.Pos.Rank == 2 || beginRank-p.Pos.Rank == -2 {
					p.EnPassant = true
				}
			}
			return true
		}
	}
	return false
}

func (p *Pawn) GetName() enums.Piece {
	return enums.Pawn
}

func (p *Pawn) GetColor() enums.Color {
	return p.Color
}

func (p *Pawn) GetPosition() helpers.Pos {
	return p.Pos
}

func (p *Pawn) GetMovesCounter() uint {
	return p.MovesCounter
}

func (p *Pawn) GetAvailibleMoves(pieces map[helpers.Pos]Piece) []helpers.Pos {
	// calculate move direction
	dir := 1
	if p.Color == enums.Black {
		dir = -1
	}

	availibleMoves := make([]helpers.Pos, 0)

	// check if can move forward
	forward := helpers.NewPos(p.Pos.File, p.Pos.Rank+dir)
	if forward.IsInBoard() {
		if pieces[forward] == nil {
			availibleMoves = append(availibleMoves, forward)
		}

		if p.MovesCounter == 0 {
			doubleForward := helpers.NewPos(p.Pos.File, p.Pos.Rank+dir*2)
			if pieces[doubleForward] == nil {
				availibleMoves = append(availibleMoves, doubleForward)
			}
		}
	}

	// left diagonal -1, right diagonal +1
	for _, df := range []int{-1, 1} {
		diagonal := helpers.NewPos(p.Pos.File+df, p.Pos.Rank+dir)
		enPassantDiagonal := helpers.NewPos(diagonal.File, p.Pos.Rank)

		if diagonal.IsInBoard() {
			targetPiece := pieces[diagonal]

			if targetPiece == nil {
				// check en passant
				enPassant := pieces[enPassantDiagonal]
				if enPassant != nil && enPassant.GetName() == enums.Pawn &&
					enPassant.GetColor() != p.Color {
					if enPassant.(*Pawn).EnPassant {
						availibleMoves = append(availibleMoves, diagonal)
					}
				}
			} else {
				if targetPiece.GetColor() != p.Color {
					availibleMoves = append(availibleMoves, diagonal)
				}
			}
		}
	}

	return availibleMoves
}
