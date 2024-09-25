package enums

import (
	"encoding/json"
	"errors"
)

type Control int

const (
	Bullet Control = iota
	Blitz
	Rapid
)

func (c Control) String() string {
	switch c {
	case 0:
		return "bullet"
	case 1:
		return "blitz"
	case 2:
		return "rapid"
	default:
		panic("unknown control")
	}
}

func ParseControl(control string) (Control, error) {
	switch control {
	case "bullet":
		return Bullet, nil
	case "blitz":
		return Blitz, nil
	case "rapid":
		return Rapid, nil
	default:
		return -1, errors.New("unknown control")
	}
}

func (c Control) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

func (c *Control) UnmarshalJSON(data []byte) (err error) {
	var control string
	if err = json.Unmarshal(data, &control); err != nil {
		return err
	}
	if *c, err = ParseControl(control); err != nil {
		return err
	}
	return nil
}
