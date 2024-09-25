package helpers

import (
	"chess-api/enums"
	"fmt"
)

type Position struct {
	File enums.File `json:"file"`
	Rank int        `json:"rank"` // 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8
}

func PosFromInd(i, j int) Position {
	var file enums.File
	switch j {
	case 0:
		file = enums.A
	case 1:
		file = enums.B
	case 2:
		file = enums.C
	case 3:
		file = enums.D
	case 4:
		file = enums.E
	case 5:
		file = enums.F
	case 6:
		file = enums.G
	case 7:
		file = enums.H
	default:
		panic("unknown file")
	}
	return Position{
		File: file,
		Rank: 8 - i,
	}
}

func (p Position) IsInBoard() bool {
	return (p.File >= 0 && p.File <= 7) && (p.Rank >= 1 && p.Rank <= 8)
}

func (p Position) String() string {
	return fmt.Sprintf("%s%d", p.File.String(), p.Rank)
}
