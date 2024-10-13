package pieces

import (
	"chess-api/models/enums"
	"chess-api/models/helpers"
)

type King struct {
	Color        enums.Color `json:"color"`
	MovesCounter uint        `json:"movesCounter"`
	Pos          helpers.Pos `json:"pos"`
	Name         enums.Piece `json:"name"`
	IsChecked    bool        `json:"isChecked"`
}

func NewKing(color enums.Color, pos helpers.Pos) *King {
	return &King{
		Color:        color,
		MovesCounter: 0,
		Pos:          pos,
		Name:         enums.King,
	}
}

func (k *King) Move(pieces map[helpers.Pos]Piece, move *helpers.Move) bool {
	possibleMoves := k.GetPossibleMoves(pieces)

	pm := possibleMoves[move.To]
	if pm != 0 && pm != enums.Defend {
		if pieces[move.To] != nil {
			move.IsCapture = true
		}

		delete(pieces, move.From)
		pieces[move.To] = k
		k.MovesCounter++
		k.Pos = move.To
		move.MoveType = pm

		// handle castling
		if move.MoveType == enums.LongCastling {
			rookPos := helpers.NewPos(k.Pos.File-2, k.Pos.Rank)
			rook := pieces[rookPos]
			delete(pieces, rookPos)
			newRookPos := helpers.NewPos(k.Pos.File+1, k.Pos.Rank)
			pieces[newRookPos] = rook
			rook.(*Rook).Pos = newRookPos
			rook.(*Rook).MovesCounter++
		} else if move.MoveType == enums.ShortCastling {
			rookPos := helpers.NewPos(k.Pos.File+1, k.Pos.Rank)
			rook := pieces[rookPos]
			delete(pieces, rookPos)
			newRookPos := helpers.NewPos(k.Pos.File-1, k.Pos.Rank)
			pieces[newRookPos] = rook
			rook.(*Rook).Pos = newRookPos
			rook.(*Rook).MovesCounter++
		}
		return true
	}
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
			if piece.GetName() == enums.Pawn {
				possibleMoves := piece.GetPossibleMoves(pieces)
				for pos, moveType := range possibleMoves {
					if moveType != enums.PawnForward {
						inaccessibleSquares[pos] = moveType
					}
				}
			} else if piece.GetName() != enums.King {
				possibleMoves := piece.GetPossibleMoves(pieces)
				for pos, moveType := range possibleMoves {
					inaccessibleSquares[pos] = moveType
				}
			} else {
				enemyKingPossibleMoves := []helpers.Pos{
					{File: piece.GetPosition().File - 1,
						Rank: piece.GetPosition().Rank + 1},
					{File: piece.GetPosition().File,
						Rank: piece.GetPosition().Rank + 1},
					{File: piece.GetPosition().File + 1,
						Rank: piece.GetPosition().Rank + 1},
					{File: piece.GetPosition().File - 1,
						Rank: piece.GetPosition().Rank},
					{File: piece.GetPosition().File + 1,
						Rank: piece.GetPosition().Rank},
					{File: piece.GetPosition().File - 1,
						Rank: piece.GetPosition().Rank - 1},
					{File: piece.GetPosition().File,
						Rank: piece.GetPosition().Rank - 1},
					{File: piece.GetPosition().File + 1,
						Rank: piece.GetPosition().Rank - 1},
				}
				for _, pos := range enemyKingPossibleMoves {
					if pos.IsInBoard() {
						inaccessibleSquares[pos] = enums.Basic
					}
				}
			}
		}
	}

	possiblePositions := []helpers.Pos{
		{File: k.Pos.File - 1, Rank: k.Pos.Rank + 1},
		{File: k.Pos.File, Rank: k.Pos.Rank + 1},
		{File: k.Pos.File + 1, Rank: k.Pos.Rank + 1},
		{File: k.Pos.File - 1, Rank: k.Pos.Rank},
		{File: k.Pos.File + 1, Rank: k.Pos.Rank},
		{File: k.Pos.File - 1, Rank: k.Pos.Rank - 1},
		{File: k.Pos.File, Rank: k.Pos.Rank - 1},
		{File: k.Pos.File + 1, Rank: k.Pos.Rank - 1},
	}

	possibleMoves := make(map[helpers.Pos]enums.MoveType)
	for _, pos := range possiblePositions {
		if inaccessibleSquares[pos] == 0 {
			if pos.IsInBoard() {
				p := pieces[pos]
				if p == nil || p.GetColor() != k.Color {
					possibleMoves[pos] = enums.Basic
				} else {
					possibleMoves[pos] = enums.Defend
				}
			}
		}
	}

	// handle castling
	if !k.IsChecked && k.MovesCounter == 0 {
		// check 0-0
		canShortCastle := true
		for i := 1; i <= 2; i++ {
			pos := helpers.NewPos(k.Pos.File+i, k.Pos.Rank)
			// is the square is vacant and is not under attack
			if pieces[pos] != nil || inaccessibleSquares[pos] != 0 {
				canShortCastle = false
			}
		}
		if canShortCastle {
			rookPos := helpers.NewPos(k.Pos.File+3, k.Pos.Rank)
			p := pieces[rookPos]
			if p != nil && p.GetName() == enums.Rook &&
				p.(*Rook).MovesCounter == 0 {
				possibleMoves[helpers.NewPos(k.Pos.File+2, k.Pos.Rank)] =
					enums.ShortCastling
			}
		}

		// check 0-0-0
		canLongCastle := true
		for i := 1; i <= 3; i++ {
			pos := helpers.NewPos(k.Pos.File-i, k.Pos.Rank)
			// is the square is vacant and is not under attack
			if pieces[pos] != nil || inaccessibleSquares[pos] != 0 {
				canLongCastle = false
			}
		}
		if canLongCastle {
			rookPos := helpers.NewPos(k.Pos.File-4, k.Pos.Rank)
			p := pieces[rookPos]
			if p != nil && p.GetName() == enums.Rook &&
				p.(*Rook).MovesCounter == 0 {
				possibleMoves[helpers.NewPos(k.Pos.File-2, k.Pos.Rank)] =
					enums.LongCastling
			}
		}
	}

	return possibleMoves
}
