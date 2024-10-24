package pieces

import (
	"chess-api/models/game/enums"
	"chess-api/models/game/helpers"
)

// traverse is a helper function that travels the board in a given direction.
// dF - delta file, dR - delta rank, pieces - board, p - piece which takes move,
// pm - map that stores all possible moves for a given piece.
func traverse(dF, dR int, pieces map[helpers.Pos]Piece,
	p Piece, pm map[helpers.Pos]enums.MoveType) {
	file, rank := p.GetPosition().File, p.GetPosition().Rank // initial piece position.
	for {
		file += dF // move to the specified file.
		rank += dR // move to the specified rank.

		nextPos := helpers.NewPos(file, rank)
		if !nextPos.IsInBoard() {
			break
		}
		// if there is a piece, current position will be the last possible
		// move for this direction.
		if piece := pieces[nextPos]; piece != nil {
			if p.GetColor() != piece.GetColor() {
				pm[nextPos] = enums.Basic
			} else {
				pm[nextPos] = enums.Defend // protect allied pieces.
			}
			break
		}
		// add empty square and continue the loop.
		pm[nextPos] = enums.Basic
	}
}
