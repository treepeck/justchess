package enums

// Color represents a color of a chessboard square, player or a piece.
// There are 2 valid colors - white and black.
type Color int

const (
	White Color = iota
	Black
)

func (c Color) Inverse() Color {
	if c == White {
		return Black
	}
	return White
}
