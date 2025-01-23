package fen

import (
	"justchess/pkg/game/bitboard"
	"justchess/pkg/game/enums"
	"justchess/pkg/game/helpers"
	"strconv"
	"strings"
)

var serializeMapping = [2]map[enums.PieceType]byte{
	{enums.Pawn: 'P',
		enums.Knight: 'N',
		enums.Bishop: 'B',
		enums.Rook:   'R',
		enums.Queen:  'Q',
		enums.King:   'K'},
	{enums.Pawn: 'p',
		enums.Knight: 'n',
		enums.Bishop: 'b',
		enums.Rook:   'r',
		enums.Queen:  'q',
		enums.King:   'k'},
}
var parseMapping = map[byte]enums.PieceType{
	'P': enums.Pawn,
	'N': enums.Knight,
	'B': enums.Bishop,
	'R': enums.Rook,
	'Q': enums.Queen,
	'K': enums.King,
	'p': enums.Pawn,
	'n': enums.Knight,
	'b': enums.Bishop,
	'r': enums.Rook,
	'q': enums.Queen,
	'k': enums.King,
}

func Bitboard2FEN(bb *bitboard.Bitboard) (FEN string) {
	FEN += serializePiecePlacementData(bb.Pieces)
	FEN += serializeActiveColor(bb.ActiveColor)
	FEN += serializeCastlingRights(bb.CastlingRights)
	FEN += serializeEnPassantTarget(bb.EPTarget)
	FEN += strconv.Itoa(bb.HalfmoveCnt) + " "
	FEN += strconv.Itoa(bb.FullmoveCnt)
	return
}

// TODO: make serializePiecePlacementData more performant.
func serializePiecePlacementData(pieces [2][7]uint64) (fenPPF string) {

	var board [64]byte
	for i := 0; i < len(pieces); i++ {
		for j := 1; j < len(pieces[0]); j++ {
			for _, pieceInd := range helpers.GetIndicesFromBitboard(pieces[i][j]) {
				board[pieceInd] = serializeMapping[i][enums.PieceType(j)+1]
			}
		}
	}
	for i := 7; i >= 0; i-- {
		row := ""
		cnt := 0
		for j := 0; j < 8; j++ {
			b := board[j+i*8]
			if b == 0 {
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
func parsePiecePlacementData(fenPPF string) (pieces [2][7]uint64) {
	squareIndex := 0
	rows := strings.Split(fenPPF, "/")
	for i := len(rows) - 1; i >= 0; i-- {
		for j := 0; j < len(rows[i]); j++ {
			b := rows[i][j]
			switch b {
			case 'P', 'N', 'B', 'R', 'Q', 'K':
				pieces[0][0] |= uint64(1) << squareIndex
				pieces[0][parseMapping[b]-1] |= uint64(1) << squareIndex
			case 'p', 'n', 'b', 'r', 'q', 'k':
				pieces[1][0] |= uint64(1) << squareIndex
				pieces[1][parseMapping[b]-1] |= uint64(1) << squareIndex
			default:
				squareIndex += int(b - '0') // Convert byte to it`s int representation.
				continue
			}
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
