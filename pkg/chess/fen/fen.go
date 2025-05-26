package fen

import (
	"math/bits"
	"strconv"
	"strings"

	"justchess/pkg/chess/bitboard"
	"justchess/pkg/chess/enums"
)

const DefaultFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

func Bitboard2FEN(bb *bitboard.Bitboard) string {
	var fenStr strings.Builder

	fenStr.WriteString(serializePiecePlacement(bb.Pieces))
	fenStr.WriteString(serializeActiveColor(bb.ActiveColor))
	fenStr.WriteString(serializeCastlingRights(bb.CastlingRights))
	fenStr.WriteString(serializeEnPassantTarget(bb.EPTarget))
	fenStr.WriteString(strconv.Itoa(bb.HalfmoveCnt) + " ")
	fenStr.WriteString(strconv.Itoa(bb.FullmoveCnt))

	return fenStr.String()
}

var pieceSymbols = [12]byte{
	'P', 'p', 'N', 'n', 'B', 'b',
	'R', 'r', 'Q', 'q', 'K', 'k',
}

func serializePiecePlacement(bitboards [12]uint64) string {
	// Used to add characters to a string without extra mem allocs.
	var piecePlacementData strings.Builder

	var board [8][8]byte

	for pieceType, bitboard := range bitboards {
		// Go through all pieces on a bitboard.
		for ; bitboard > 0; bitboard &= bitboard - 1 {
			squareIndex := bits.TrailingZeros64(bitboard)
			// Add piece on board.
			board[squareIndex/8][squareIndex%8] = pieceSymbols[pieceType]
		}
	}

	var numOfEmptySquares byte

	for rank := 7; rank >= 0; rank-- {
		for file := 0; file < 8; file++ {
			char := board[rank][file]

			if char == 0 { // Empty square.
				numOfEmptySquares++
			} else { // Piece on square.
				if numOfEmptySquares > 0 {
					piecePlacementData.WriteByte('0' + numOfEmptySquares)
					numOfEmptySquares = 0
				}
				piecePlacementData.WriteByte(char)
			}

			// To add rank separators.
			squareIndex := 8*rank + file
			if (squareIndex+1)%8 == 0 {
				if numOfEmptySquares > 0 {
					piecePlacementData.WriteByte('0' + numOfEmptySquares)
					numOfEmptySquares = 0
				}
				// Do not add separator in the end of the string.
				if squareIndex != 7 {
					piecePlacementData.WriteByte('/')
				}
			}
		}
	}
	piecePlacementData.WriteByte(' ')

	return piecePlacementData.String()
}

func serializeActiveColor(c enums.Color) (ac string) {
	// White space is needed to be able to split the FEN fields.
	if c == enums.White {
		return "w "
	}
	return "b "
}

func serializeCastlingRights(cr [4]bool) string {
	var fenCRF strings.Builder

	if cr[0] {
		fenCRF.WriteByte('K')
	}
	if cr[2] {
		fenCRF.WriteByte('Q')
	}
	if cr[1] {
		fenCRF.WriteByte('k')
	}
	if cr[3] {
		fenCRF.WriteByte('q')
	}
	if fenCRF.Len() == 0 {
		fenCRF.WriteByte('-')
	}
	fenCRF.WriteByte(' ')

	return fenCRF.String()
}

func serializeEnPassantTarget(epTarget int) string {
	var fenEPF strings.Builder

	if epTarget < 0 {
		return "- "
	}
	// Calculate file and rank.
	files := "abcdefgh"
	fenEPF.WriteByte(files[epTarget%8])
	fenEPF.WriteString(strconv.Itoa((epTarget / 8) + 1))
	fenEPF.WriteByte(' ')

	return fenEPF.String()
}

func FEN2Bitboard(FEN string) *bitboard.Bitboard {
	fields := strings.Split(FEN, " ")
	pieces := parsePiecePlacement(fields[0])
	color := parseActiveColor(fields[1])
	castlingRights := parseCastlingRights(fields[2])
	epTarget := parseEnPassantTarget(fields[3])
	halfmoveClk, _ := strconv.Atoi(fields[4])
	fullmoveClk, _ := strconv.Atoi(fields[5])
	return bitboard.NewBitboard(pieces, color, castlingRights,
		epTarget, halfmoveClk, fullmoveClk)
}

// parsePiecePlacement parses bitboards from FEN's piece placement field.
func parsePiecePlacement(piecePlacementData string) [12]uint64 {
	var bitboards [12]uint64
	squareIndex := 56

	// Piece placement data describes each rank beginning from the eigth.
	for i := 0; i < len(piecePlacementData); i++ {
		char := piecePlacementData[i]

		if char == '/' { // Rank separator.
			squareIndex -= 16
		} else if char >= '1' && char <= '8' { // Number of consecutive empty squares.
			// Convert byte to the integer it represents.
			squareIndex += int(char - '0')
		} else { // There is piece on a square.
			var pieceType enums.PieceType // enum.PieceWPawn by default.
			// Manual switch construction is ~3x faster than map approach.
			switch char {
			case 'N':
				pieceType = enums.WhiteKnight
			case 'B':
				pieceType = enums.WhiteBishop
			case 'R':
				pieceType = enums.WhiteRook
			case 'Q':
				pieceType = enums.WhiteQueen
			case 'K':
				pieceType = enums.WhiteKing
			case 'p':
				pieceType = enums.BlackPawn
			case 'n':
				pieceType = enums.BlackKnight
			case 'b':
				pieceType = enums.BlackBishop
			case 'r':
				pieceType = enums.BlackRook
			case 'q':
				pieceType = enums.BlackQueen
			case 'k':
				pieceType = enums.BlackKing
			}
			// Set the bit on the bitboard to place a piece.
			bitboards[pieceType] |= 1 << squareIndex
			squareIndex++
		}
	}

	return bitboards
}

func parseActiveColor(fenACF string) enums.Color {
	if fenACF == "w" {
		return enums.White
	}
	return enums.Black
}

func parseCastlingRights(fenCRF string) (cr [4]bool) {
	for i := 0; i < len(fenCRF); i++ {
		switch fenCRF[i] {
		case 'K':
			cr[0] = true
		case 'k':
			cr[1] = true
		case 'Q':
			cr[2] = true
		case 'q':
			cr[3] = true
		}
	}
	return
}

func parseEnPassantTarget(fenEPF string) (index int) {
	if fenEPF == "-" {
		return -1
	}
	// Check file.
	switch fenEPF[0] {
	case 'a':
		index = 0
	case 'b':
		index = 1
	case 'c':
		index = 2
	case 'd':
		index = 3
	case 'e':
		index = 4
	case 'f':
		index = 5
	case 'g':
		index = 6
	case 'h':
		index = 7
	}
	rank := fenEPF[1] - '0'
	return index + int(rank-1)*8
}
