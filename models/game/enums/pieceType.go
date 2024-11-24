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
		return "♙"
	case 2:
		return "♖"
	case 3:
		return "♘"
	case 4:
		return "♗"
	case 5:
		return "♕"
	case 6:
		return "♔"
	default:
		return ""
	}
}

func ParsePiece(piece string) PieceType {
	switch piece {
	case "♙":
		return Pawn
	case "♖":
		return Rook
	case "♘":
		return Knight
	case "♗":
		return Bishop
	case "♕":
		return Queen
	case "♔":
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
