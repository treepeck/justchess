package enums

type MoveType int

const (
	Basic  MoveType = iota + 1 // move to an empty square
	Defend                     // forbits the enemy king to move on an attacked square
	LongCastling
	ShortCastling
	EnPassant
	Promotion
)

func (mt MoveType) String() string {
	switch mt {
	case Basic:
		return "basic"
	case Defend:
		return "defend"
	case LongCastling:
		return "longCastling"
	case ShortCastling:
		return "shortCastling"
	case EnPassant:
		return "enPassant"
	case Promotion:
		return "promotion"
	default:
		return "unknown move type"
	}
}
