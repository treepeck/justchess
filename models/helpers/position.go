package helpers

import (
	"fmt"
)

type Pos struct {
	File int `json:"file"` // 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8
	Rank int `json:"rank"` // 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8
}

func NewPos(file, rank int) Pos {
	return Pos{File: file, Rank: rank}
}

func PosFromInd(i, j int) Pos {
	return Pos{
		File: j,
		Rank: 8 - i,
	}
}

func (p Pos) IsInBoard() bool {
	return (p.File >= 1 && p.File <= 8) && (p.Rank >= 1 && p.Rank <= 8)
}

func (p Pos) String() string {
	file := ""
	switch p.File {
	case 1:
		file = "a"
	case 2:
		file = "b"
	case 3:
		file = "c"
	case 4:
		file = "d"
	case 5:
		file = "e"
	case 6:
		file = "f"
	case 7:
		file = "g"
	case 8:
		file = "h"
	}
	return fmt.Sprintf("%s%d", file, p.Rank)
}
