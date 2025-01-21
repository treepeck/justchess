package fen

import (
	"justchess/pkg/game/bitboard"
	"justchess/pkg/game/enums"
	"justchess/pkg/game/helpers"
	"strconv"
	"strings"
)

func Bitboard2FEN(bb *bitboard.Bitboard) (FEN string) {
	FEN += serializePiecePlacementData(bb.Pieces)
	FEN += serializeActiveColor(bb.ActiveColor)
	FEN += serializeCastlingRights(bb.CastlingRights)
	FEN += serializeEnPassantTarget(bb.EpTarget)
	FEN += strconv.Itoa(bb.HalfmoveClk) + " "
	FEN += strconv.Itoa(bb.FullmoveClk)
	return
}

func serializePiecePlacementData(pieces [14]uint64) (fenPPF string) {
	mapping := map[enums.PieceType]byte{
		enums.WhitePawn:   'P',
		enums.WhiteKnight: 'N',
		enums.WhiteBishop: 'B',
		enums.WhiteRook:   'R',
		enums.WhiteQueen:  'Q',
		enums.WhiteKing:   'K',
		enums.BlackPawn:   'p',
		enums.BlackKnight: 'n',
		enums.BlackBishop: 'b',
		enums.BlackRook:   'r',
		enums.BlackQueen:  'q',
		enums.BlackKing:   'k',
	}
	var board [64]enums.PieceType
	for i := 2; i < 14; i++ {
		for _, squareIndex := range helpers.GetIndicesFromBitboard(pieces[i]) {
			board[squareIndex] = enums.PieceType(i)
		}
	}

	for i := 7; i >= 0; i-- {
		row := ""
		cnt := 0
		for j := 0; j < 8; j++ {
			if b, ok := mapping[board[j+int(i)*8]]; !ok {
				cnt++
			} else {
				if cnt > 0 {
					row += strconv.Itoa(cnt)
					cnt = 0
				}
				row += string(b)
			}
		}
		if cnt > 0 {
			row += strconv.Itoa(cnt)
		}
		fenPPF += row + "/"
	}
	return fenPPF[0:len(fenPPF)-1] + " " // Remove the last "/".
}

func serializeActiveColor(c enums.Color) (ac string) {
	// White space is needed to be able to split the FEN fields.
	if c == enums.White {
		return "w "
	}
	return "b "
}

func serializeCastlingRights(cr [4]bool) (fenCRF string) {
	if cr[0] {
		fenCRF += "K"
	}
	if cr[1] {
		fenCRF += "Q"
	}
	if cr[2] {
		fenCRF += "k"
	}
	if cr[3] {
		fenCRF += "q"
	}
	if len(fenCRF) == 0 {
		fenCRF += "-"
	}
	return fenCRF + " "
}

func serializeEnPassantTarget(epTarget int) (fenEPF string) {
	if epTarget < 0 {
		return "- "
	}
	// Calculate file and rank.
	files := "abcdefgh"
	fenEPF += string(files[epTarget%8])
	fenEPF += strconv.Itoa((epTarget / 8) + 1)
	return fenEPF + " "
}

func FEN2Bitboard(FEN string) *bitboard.Bitboard {
	fields := strings.Split(FEN, " ")
	pieces := parsePiecePlacementData(fields[0])
	color := parseActiveColor(fields[1])
	castlingRights := parseCastlingRights(fields[2])
	epTarget := parseEnPassantTarget(fields[3])
	halfmoveClk, _ := strconv.Atoi(fields[4])
	fullmoveClk, _ := strconv.Atoi(fields[5])
	return bitboard.NewBitboard(pieces, color, castlingRights,
		epTarget, halfmoveClk, fullmoveClk)
}

// parsePiecePlacementData parses bitboards from FEN`s piece placement field.
func parsePiecePlacementData(fenPPF string) (pieces [14]uint64) {
	mapping := map[byte]enums.PieceType{
		'P': enums.WhitePawn,
		'N': enums.WhiteKnight,
		'B': enums.WhiteBishop,
		'R': enums.WhiteRook,
		'Q': enums.WhiteQueen,
		'K': enums.WhiteKing,
		'p': enums.BlackPawn,
		'n': enums.BlackKnight,
		'b': enums.BlackBishop,
		'r': enums.BlackRook,
		'q': enums.BlackQueen,
		'k': enums.BlackKing,
	}
	squareIndex := 0
	rows := strings.Split(fenPPF, "/")
	for i := len(rows) - 1; i >= 0; i-- {
		for j := 0; j < len(rows[i]); j++ {
			b := rows[i][j]
			switch b {
			case 'P', 'N', 'B', 'R', 'Q', 'K':
				pieces[0] |= uint64(1) << squareIndex
			case 'p', 'n', 'b', 'r', 'q', 'k':
				pieces[1] |= uint64(1) << squareIndex
			default:
				squareIndex += int(b - '0') // Convert byte to it`s int representation.
				continue
			}
			pieces[mapping[b]] |= uint64(1) << squareIndex
			squareIndex++
		}
	}
	return
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
		case 'Q':
			cr[1] = true
		case 'k':
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
