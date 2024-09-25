package enums

import (
	"encoding/json"
	"errors"
)

type File int

const (
	A File = iota
	B
	C
	D
	E
	F
	G
	H
)

func (f File) String() string {
	switch f {
	case 0:
		return "a"
	case 1:
		return "b"
	case 2:
		return "c"
	case 3:
		return "d"
	case 4:
		return "e"
	case 5:
		return "f"
	case 6:
		return "g"
	case 7:
		return "h"
	default:
		panic("unknown file")
	}
}

func ParseFile(f string) (File, error) {
	switch f {
	case "a":
		return A, nil
	case "b":
		return B, nil
	case "c":
		return C, nil
	case "d":
		return D, nil
	case "e":
		return E, nil
	case "f":
		return F, nil
	case "g":
		return G, nil
	case "h":
		return H, nil
	default:
		return -1, errors.New("unknown file")
	}
}

func (f File) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.String())
}

func (f *File) UnmarshalJSON(data []byte) (err error) {
	var file string
	if err = json.Unmarshal(data, &file); err != nil {
		return err
	}
	if *f, err = ParseFile(file); err != nil {
		return err
	}
	return nil
}
