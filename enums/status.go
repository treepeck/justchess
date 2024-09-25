package enums

import (
	"encoding/json"
	"errors"
)

type Status int

const (
	Canceled Status = iota
	Waiting
	WhiteWon
	BlackWon
	Draw
	Continues
)

func (s Status) String() string {
	switch s {
	case 0:
		return "canceled"
	case 1:
		return "waiting"
	case 2:
		return "white_won"
	case 3:
		return "black_won"
	case 4:
		return "draw"
	case 5:
		return "continues"
	default:
		panic("unknown status")
	}
}

func ParseStatus(status string) (Status, error) {
	switch status {
	case "canceled":
		return Canceled, nil
	case "waiting":
		return Waiting, nil
	case "white_won":
		return WhiteWon, nil
	case "black_won":
		return BlackWon, nil
	case "draw":
		return Draw, nil
	case "continues":
		return Continues, nil
	default:
		return -1, errors.New("unknown status")
	}
}

func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *Status) UnmarshalJSON(data []byte) (err error) {
	var status string
	if err = json.Unmarshal(data, &status); err != nil {
		return err
	}
	if *s, err = ParseStatus(status); err != nil {
		return err
	}
	return nil
}
