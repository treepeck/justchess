package pieces

import (
	"justchess/pkg/models/game/enums"
	"justchess/pkg/models/game/helpers"
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
) []helpers.PossibleMove {
	pm := make([]helpers.PossibleMove, 0)
	// determine move direction.
	dir := 1
	if p.Color == enums.Black {
		dir = -1
	}

	forward := helpers.NewPos(p.Pos.File, p.Pos.Rank+dir)
	if forward.IsInBoard() && pieces[forward] == nil {
		if forward.Rank > 1 && forward.Rank < 8 {
			pm = append(pm, helpers.NewPM(forward, enums.PawnForward))
		} else {
			pm = append(pm, helpers.NewPM(forward, enums.Promotion))
		}
		if p.MovesCounter == 0 {
			doubleForward := helpers.NewPos(p.Pos.File, p.Pos.Rank+dir*2)
			if pieces[doubleForward] == nil {
				// do not check the promotion since it is impossible on a first move.
				pm = append(pm, helpers.NewPM(doubleForward, enums.PawnForward))
			}
		}
	}
	// check diagonal squares to handle captures.
	for _, f := range []int{-1, 1} {
		diagonal := helpers.NewPos(p.Pos.File+f, p.Pos.Rank+dir)
		if !diagonal.IsInBoard() {
			continue
		}
		enPassantPos := helpers.NewPos(diagonal.File, p.Pos.Rank)

		targetPiece := pieces[diagonal]
		if targetPiece == nil {
			ep := pieces[enPassantPos]
			// if it is a pawn
			if ep != nil && ep.GetType() == enums.Pawn &&
				ep.GetColor() != p.Color {
				// if it can be captured en passant
				if ep.(*Pawn).IsEnPassant {
					pm = append(pm, helpers.NewPM(diagonal, enums.EnPassant))
					continue
				}
			}
			pm = append(pm, helpers.NewPM(diagonal, enums.Defend))
		} else { // if there is an enemy piece.
			if targetPiece.GetColor() != p.Color {
				if targetPiece.GetPosition().Rank == 1 ||
					targetPiece.GetPosition().Rank == 8 {
					pm = append(pm, helpers.NewPM(diagonal, enums.Promotion))
				} else {
					pm = append(pm, helpers.NewPM(diagonal, enums.Basic))
				}
			}
		}
	}
	return pm
}

func (p *Pawn) Move(to helpers.Pos) {
	// if the pawn moves double forward,
	// it can be captured en passant on the next turn.
	if p.MovesCounter == 0 {
		if (p.Pos.Rank == 2 || p.Pos.Rank == 7) &&
			(to.Rank == 4 || to.Rank == 5) {
			p.IsEnPassant = true
		}
	}

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

func (p *Pawn) GetFEN() string {
	if p.Color == enums.White {
		return "P"
	}
	return "p"
}
