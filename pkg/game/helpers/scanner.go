package helpers

import (
	"justchess/pkg/game/enums"
	"math/bits"
)

// GetIndicesFromBitboard returns the indices of all set bits in the bitboard.
func GetIndicesFromBitboard(bb uint64) []int {
	indices := make([]int, 0)
	for bb != 0 {
		i := bits.TrailingZeros64(bb)
		indices = append(indices, i)
		// Clear the LSB.
		bb &= bb - 1
	}
	return indices
}

// GetMovesFromBitboard encodes the slice of moves from a bitboard.
func GetMovesFromBitboard(from int, moves, enemies uint64, pt enums.PieceType) []Move {
	mSlice := make([]Move, 0)
	for _, i := range GetIndicesFromBitboard(moves) {
		moveType := enums.Quiet
		if uint64(1)<<i&enemies != 0 {
			moveType = enums.Capture
		}
		mSlice = append(mSlice, NewMove(i, from, moveType, pt))
	}
	return mSlice
}
