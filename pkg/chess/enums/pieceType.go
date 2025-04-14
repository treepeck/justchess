package enums

type PieceType int

const (
	WhitePawn PieceType = iota
	BlackPawn
	WhiteKnight
	BlackKnight
	WhiteBishop
	BlackBishop
	WhiteRook
	BlackRook
	WhiteQueen
	BlackQueen
	WhiteKing
	BlackKing
)

func (pt PieceType) String() string {
	switch pt {
	case WhiteKnight, BlackKnight:
		return "N"
	case WhiteBishop, BlackBishop:
		return "B"
	case WhiteRook, BlackRook:
		return "R"
	case WhiteQueen, BlackQueen:
		return "Q"
	case WhiteKing, BlackKing:
		return "K"
	default:
		return ""
	}
}
