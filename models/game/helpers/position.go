package helpers

import (
	"chess-api/models/game/enums"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

type Pos struct {
	File int // 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8
	Rank int // 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8
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

func (p Pos) IsEqual(other Pos) bool {
	return p.File == other.File && p.Rank == other.Rank
}

func (p Pos) String() string {
	file := ""
	switch p.File {
	case enums.A:
		file = "a"
	case enums.B:
		file = "b"
	case enums.C:
		file = "c"
	case enums.D:
		file = "d"
	case enums.E:
		file = "e"
	case enums.F:
		file = "f"
	case enums.G:
		file = "g"
	case enums.H:
		file = "h"
	}
	return fmt.Sprintf("%s%d", file, p.Rank)
}

func ParsePos(ps string) (p Pos, err error) {
	switch string(ps[0]) { // parse file
	case "a":
		p.File = enums.A
	case "b":
		p.File = enums.B
	case "c":
		p.File = enums.C
	case "d":
		p.File = enums.D
	case "e":
		p.File = enums.E
	case "f":
		p.File = enums.F
	case "g":
		p.File = enums.G
	case "h":
		p.File = enums.H
	default:
		return p, errors.New("unknown file")
	}
	p.Rank, err = strconv.Atoi(string(ps[1]))
	return
}

func (p Pos) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

func (p *Pos) UnmarshalJSON(data []byte) (err error) {
	var pos string
	if err = json.Unmarshal(data, &pos); err != nil {
		return err
	}
	if *p, err = ParsePos(pos); err != nil {
		return err
	}
	return nil
}
