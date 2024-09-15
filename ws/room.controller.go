package ws

import (
	"log/slog"
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
		if r.IsAvailible {
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

func (rc *RoomController) CloseRoom(id uuid.UUID) {
	for r := range rc.Rooms {
		if r.Id == id {
			r.IsAvailible = false
		}
	}
}
