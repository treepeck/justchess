package enums

import (
	"encoding/json"
	"errors"
)

type Piece int

const (
	Pawn Piece = iota
	Rook
	Knight
	Bishop
	Queen
	King
)

func (p Piece) String() string {
	switch p {
	case 0:
		return "pawn"
	case 1:
		return "rook"
	case 2:
		return "knight"
	case 3:
		return "bishop"
	case 4:
		return "queen"
	case 5:
		return "king"
	default:
		panic("unknown piece")
	}
}

func ParsePiece(control string) (Piece, error) {
	switch control {
	case "pawn":
		return Pawn, nil
	case "rook":
		return Rook, nil
	case "knight":
		return Knight, nil
	case "bishop":
		return Bishop, nil
	case "queen":
		return Queen, nil
	case "king":
		return King, nil
	default:
		return -1, errors.New("unknown piece")
	}
}

func (p Piece) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

func (p *Piece) UnmarshalJSON(data []byte) (err error) {
	var piece string
	if err = json.Unmarshal(data, &piece); err != nil {
		return err
	}
	if *p, err = ParsePiece(piece); err != nil {
		return err
	}
	return nil
}
