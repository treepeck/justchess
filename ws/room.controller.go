package ws

import (
	"chess-api/models/enums"
	"log/slog"
	"strconv"
	"sync"

	"github.com/google/uuid"
)

type RoomController struct {
	sync.Mutex
	Rooms map[*Room]bool `json:"rooms"`
}

func NewController() *RoomController {
	return &RoomController{
		Rooms: make(map[*Room]bool),
	}
}

func (m *RoomController) AddRoom(r *Room) {
	m.Lock()
	defer m.Unlock()

	fn := slog.String("func", "room.AddRoom")
	m.Rooms[r] = true
	slog.Info("room added", fn, slog.Int("count", len(m.Rooms)))
}

func (rc *RoomController) RemoveRoom(r *Room) {
	rc.Lock()
	defer rc.Unlock()

	fn := slog.String("func", "room.RemoveRoom")
	if _, ok := rc.Rooms[r]; ok {
		delete(rc.Rooms, r)
		slog.Info("room removed", fn, slog.Int("count", len(rc.Rooms)))
	}
}

func (rc *RoomController) GetAll() (rooms []*Room) {
	for r := range rc.Rooms {
		rooms = append(rooms, r)
	}
	return
}

func (rc *RoomController) FindAvailible() (rooms []*Room) {
	for r := range rc.Rooms {
		if r.Game.Status == enums.Waiting {
			rooms = append(rooms, r)
		}
	}
	return
}

func (rc *RoomController) FindById(id uuid.UUID) *Room {
	for r := range rc.Rooms {
		if r.Id == id {
			return r
		}
	}
	return nil
}

func (rc *RoomController) FindByOwnerId(id uuid.UUID) *Room {
	for r := range rc.Rooms {
		if r.Owner.Id == id {
			return r
		}
	}
	return nil
}

func (rc *RoomController) FilterRooms(gr GetRoomDTO) (rooms []*Room) {
	for r := range rc.Rooms {
		if gr.Bonus != "all" {
			bonus, err := strconv.ParseUint(gr.Bonus, 10, 32)
			if err == nil && r.Game.Bonus != uint(bonus) {
				continue
			}
		}
		if gr.Control != "all" {
			control, err := enums.ParseControl(gr.Control)
			if err == nil && r.Game.Control != control {
				continue
			}
		}
		rooms = append(rooms, r)
	}
	return
}
