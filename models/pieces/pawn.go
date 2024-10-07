package pieces

import (
	"chess-api/models/enums"
	"chess-api/models/helpers"
)

type Pawn struct {
	Pos          helpers.Pos `json:"pos"`
	Name         enums.Piece `json:"name"`
	Color        enums.Color `json:"color"`
	MovesCounter uint        `json:"movesCounter"`
	IsEnPassant  bool        `json:"isEnPassant"`
}

func NewPawn(color enums.Color, pos helpers.Pos) *Pawn {
	return &Pawn{
		Pos:          pos,
		Name:         enums.Pawn,
		Color:        color,
		MovesCounter: 0,
		IsEnPassant:  false,
	}
}

func (p *Pawn) Move(pieces map[helpers.Pos]Piece, move *helpers.Move) bool {
	possibleMoves := p.GetPossibleMoves(pieces)

	pm := possibleMoves[move.To]
	if pm != 0 && pm != enums.Defend {
		if pieces[move.To] != nil {
			move.IsCapture = true
		}

		delete(pieces, move.From)
		pieces[move.To] = p
		p.MovesCounter++
		p.Pos = move.To

		if pm == enums.Promotion {
			switch move.PromotionPayload {
			case enums.Knight:
				pieces[p.Pos] = NewKnight(p.Color, move.To)
			case enums.Bishop:
				pieces[p.Pos] = NewBishop(p.Color, move.To)
			case enums.Rook:
				rook := NewRook(p.Color, move.To)
				rook.MovesCounter = p.MovesCounter
				pieces[p.Pos] = rook
			default:
				pieces[p.Pos] = NewQueen(p.Color, move.To)
			}
		} else if pm == enums.EnPassant {
			if p.Color == enums.White {
				delete(pieces, helpers.NewPos(p.Pos.File, p.Pos.Rank-1))
			} else {
				delete(pieces, helpers.NewPos(p.Pos.File, p.Pos.Rank+1))
			}
		}
		return true
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

func (p *Pawn) GetPossibleMoves(pieces map[helpers.Pos]Piece,
) map[helpers.Pos]enums.MoveType {
	// define move direction
	dir := 1
	if p.Color == enums.Black {
		dir = -1
	}

	possibleMoves := make(map[helpers.Pos]enums.MoveType)

	// check if can move forward
	forward := helpers.NewPos(p.Pos.File, p.Pos.Rank+dir)
	if forward.IsInBoard() {
		if pieces[forward] == nil {
			if forward.Rank > 1 && forward.Rank < 8 {
				possibleMoves[forward] = enums.PawnForward
			} else {
				// promition is possible
				possibleMoves[forward] = enums.Promotion
			}
		}

		if p.MovesCounter == 0 {
			doubleForward := helpers.NewPos(p.Pos.File, p.Pos.Rank+dir*2)
			if pieces[doubleForward] == nil {
				// promotion is impossible in a first move
				possibleMoves[doubleForward] = enums.PawnForward
			}
		}
	}

	// left file = -1, right file = +1
	for _, f := range []int{-1, 1} {
		diagonal := helpers.NewPos(p.Pos.File+f, p.Pos.Rank+dir)
		possibleEnPassant := helpers.NewPos(diagonal.File, p.Pos.Rank)

		if diagonal.IsInBoard() {
			// in any case the pawn defends the square
			possibleMoves[diagonal] = enums.Defend

			targetPiece := pieces[diagonal]
			if targetPiece == nil {
				// check en passant case
				ep := pieces[possibleEnPassant]
				// if it is a pawn
				if ep != nil && ep.GetName() == enums.Pawn &&
					ep.GetColor() != p.Color {
					// if it can be captured en passant
					if ep.(*Pawn).IsEnPassant {
						possibleMoves[diagonal] = enums.EnPassant
					}
				}
			} else {
				// if there is an enemy piece
				if targetPiece.GetColor() != p.Color {
					possibleMoves[diagonal] = enums.Basic
					if targetPiece.GetPosition().Rank == 1 ||
						targetPiece.GetPosition().Rank == 8 {
						possibleMoves[diagonal] = enums.Promotion
					}
				}
			}
		}
	}

	return possibleMoves
}
