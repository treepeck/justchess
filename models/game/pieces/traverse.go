package pieces

import (
	"chess-api/models/game/enums"
	"chess-api/models/game/helpers"
)

// traverse travels the board in a given direction.
// dF - delta file, dR - delta rank, pieces - board, p - piece which takes move.
func traverse(dF, dR int, pieces map[helpers.Pos]Piece,
	p Piece) []helpers.PossibleMove {
	pm := make([]helpers.PossibleMove, 0)
	file, rank := p.GetPosition().File, p.GetPosition().Rank // initial piece position.
	for {
		file += dF // move to the specified file.
		rank += dR // move to the specified rank.

		nextPos := helpers.NewPos(file, rank)
		if !nextPos.IsInBoard() {
			break
		}
		// if there is a piece, current position will be the last possible
		// move in this direction.
		if piece := pieces[nextPos]; piece != nil {
			if p.GetColor() != piece.GetColor() {
				pm = append(pm, helpers.NewPM(nextPos, enums.Basic))
			} else {
				pm = append(pm, helpers.NewPM(nextPos, enums.Defend))
			}
			break
		}
		// add empty square and continue the loop.
		pm = append(pm, helpers.NewPM(nextPos, enums.Basic))
	}
	return pm
}
