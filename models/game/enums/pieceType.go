package enums

import (
	"encoding/json"
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

func ParsePiece(piece string) PieceType {
	switch piece {
	case "pawn":
		return Pawn
	case "rook":
		return Rook
	case "knight":
		return Knight
	case "bishop":
		return Bishop
	case "queen":
		return Queen
	case "king":
		return King
	default:
		return 0
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
	*p = ParsePiece(piece)
	return nil
}
