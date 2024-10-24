package enums

import (
	"encoding/json"
	"errors"
)

type PieceType int

const (
	Pawn PieceType = iota + 1
	Rook
	Knight
	Bishop
	Queen
	King
)

func (p PieceType) String() string {
	switch p {
	case 1:
		return "pawn"
	case 2:
		return "rook"
	case 3:
		return "knight"
	case 4:
		return "bishop"
	case 5:
		return "queen"
	case 6:
		return "king"
	default:
		return "unknown piece"
	}
}

func ParsePiece(piece string) (PieceType, error) {
	switch piece {
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
		return 0, errors.New("unknown piece")
	}
}

func (p PieceType) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

func (p *PieceType) UnmarshalJSON(data []byte) (err error) {
	var piece string
	if err = json.Unmarshal(data, &piece); err != nil {
		return err
	}
	if *p, err = ParsePiece(piece); err != nil {
		return err
	}
	return nil
}
