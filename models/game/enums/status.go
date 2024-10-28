package enums

import (
	"encoding/json"
)

type Status int

const (
	Aborted   Status = iota // one of the players did not make the first move or cancelled the game.
	Waiting                 // the game doesn't start until both sides connect.
	WhiteWon                // white player won.
	BlackWon                // black player won.
	Draw                    // draw or by agreement or by stalemate or there is not enough pieces to checkmate.
	Continues               // game continues.
)

func (s Status) String() string {
	switch s {
	case 0:
		return "aborted"
	case 1:
		return "waiting"
	case 2:
		return "white_won"
	case 3:
		return "black_won"
	case 4:
		return "draw"
	case 5:
		return "continues"
	default:
		panic("unknown status")
	}
}

func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}
