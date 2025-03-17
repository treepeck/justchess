package user

import (
	"time"

	"github.com/google/uuid"
)

type Role int

const (
	GUEST int = iota
	USER
)

// User stores all user data.
type User struct {
	Id           uuid.UUID `json:"id"`
	Name         string    `json:"username"`
	RegisteredAt time.Time `json:"registeredAt"`
	Role         Role      `json:"role"`
	password     string    `json:"-"`
}

func NewUser(id uuid.UUID) User {
	return User{
		Id:           id,
		Name:         "Guest-" + id.String()[:8],
		RegisteredAt: time.Now(),
	}
}

func (u User) GetPassword() string {
	return u.password
}
