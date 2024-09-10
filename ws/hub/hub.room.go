package hub

import (
	"chess-api/models"

	"github.com/google/uuid"
)

type hubRoom struct {
	Id      uuid.UUID   `json:"id"`
	Owner   models.User `json:"owner"`
	Control string      `json:"control"`
	Bonus   uint        `json:"bonus"`
}

type createRoomDTO struct {
	Owner   models.User `json:"owner"`
	Control string      `json:"control"`
	Bonus   uint        `json:"bonus"`
}

// Creates a new room.
func newRoom(cr createRoomDTO) *hubRoom {
	return &hubRoom{
		Id:      uuid.New(),
		Owner:   cr.Owner,
		Control: cr.Control,
		Bonus:   cr.Bonus,
	}
}
