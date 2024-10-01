package enums

import (
	"encoding/json"
	"errors"
)

type Color int

const (
	White Color = iota
	Black
)

func (c Color) String() string {
	switch c {
	case 0:
		return "white"
	case 1:
		return "black"
	default:
		panic("unknown color")
	}
}

func ParseColor(color string) (Color, error) {
	switch color {
	case "white":
		return White, nil
	case "black":
		return Black, nil
	default:
		return -1, errors.New("unknown color")
	}
}

func (c Color) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

func (c *Color) UnmarshalJSON(data []byte) (err error) {
	var color string
	if err = json.Unmarshal(data, &color); err != nil {
		return err
	}
	if *c, err = ParseColor(color); err != nil {
		return err
	}
	return nil
}
