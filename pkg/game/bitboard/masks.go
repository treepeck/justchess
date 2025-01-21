package bitboard

// The following block on contants defines the bit masks needed to
// correctly calculate possible moves by performing bitwise
// operations on a bitboard.
const (
	notA  uint64 = 0xFEFEFEFEFEFEFEFE // Mask for all files except the A.
	notH  uint64 = 0x7F7F7F7F7F7F7F7F // Mask for all files except the H.
	notAB uint64 = 0xFCFCFCFCFCFCFCFC // Mask for all files except the A and B.
	notGH uint64 = 0x3F3F3F3F3F3F3F3F // Mask for all files except the G and H.
)
