package pieces

import (
	"chess-api/models/game/enums"
	"chess-api/models/game/helpers"
)

type Pawn struct {
	Pos          helpers.Pos     `json:"-"`
	Type         enums.PieceType `json:"type"`
	Color        enums.Color     `json:"color"`
	MovesCounter uint            `json:"-"`
	IsEnPassant  bool            `json:"-"`
}

func NewPawn(color enums.Color, pos helpers.Pos) *Pawn {
	return &Pawn{
		Pos:          pos,
		Type:         enums.Pawn,
		Color:        color,
		MovesCounter: 0,
		IsEnPassant:  false,
	}
}

func (p *Pawn) GetPossibleMoves(pieces map[helpers.Pos]Piece,
) map[helpers.Pos]enums.MoveType {
	pm := make(map[helpers.Pos]enums.MoveType)

	// determine move direction
	dir := 1
	if p.Color == enums.Black {
		dir = -1
	}

	// check if can move forward
	forward := helpers.NewPos(p.Pos.File, p.Pos.Rank+dir)
	if forward.IsInBoard() {
		if pieces[forward] == nil {
			if forward.Rank > 1 && forward.Rank < 8 {
				pm[forward] = enums.PawnForward
			} else {
				// promition is possible
				pm[forward] = enums.Promotion
			}
			if p.MovesCounter == 0 {
				doubleForward := helpers.NewPos(p.Pos.File, p.Pos.Rank+dir*2)
				if pieces[doubleForward] == nil {
					// promotion is impossible in a first move
					pm[doubleForward] = enums.PawnForward
				}
			}
		}
	}

	// left file = -1, right file = +1
	for _, f := range []int{-1, 1} {
		diagonal := helpers.NewPos(p.Pos.File+f, p.Pos.Rank+dir)
		possibleEnPassant := helpers.NewPos(diagonal.File, p.Pos.Rank)

		if diagonal.IsInBoard() {
			// in any case the pawn defends the square
			pm[diagonal] = enums.Defend

			targetPiece := pieces[diagonal]
			if targetPiece == nil {
				// check en passant case
				ep := pieces[possibleEnPassant]
				// if it is a pawn
				if ep != nil && ep.GetType() == enums.Pawn &&
					ep.GetColor() != p.Color {
					// if it can be captured en passant
					if ep.(*Pawn).IsEnPassant {
						pm[diagonal] = enums.EnPassant
					}
				}
			} else {
				// if there is an enemy piece
				if targetPiece.GetColor() != p.Color {
					pm[diagonal] = enums.Basic
					if targetPiece.GetPosition().Rank == 1 ||
						targetPiece.GetPosition().Rank == 8 {
						pm[diagonal] = enums.Promotion
					}
				}
			}
		}
	}
	return pm
}

func (p *Pawn) Move(to helpers.Pos) {
	p.Pos = to
	p.MovesCounter++
}

func (p *Pawn) GetMovesCounter() uint {
	return p.MovesCounter
}

func (p *Pawn) SetMovesCounter(mc uint) {
	p.MovesCounter = mc
}

func (p *Pawn) GetType() enums.PieceType {
	return enums.Pawn
}

func (p *Pawn) GetColor() enums.Color {
	return p.Color
}

func (p *Pawn) GetPosition() helpers.Pos {
	return p.Pos
}
