package enums

// Color represents a color of a chessboard square, player or a piece.
// There are 2 valid colors - white and black.
// To inverse the color, perform the XOR.
type Color int

const (
	White Color = iota
	Black
)
