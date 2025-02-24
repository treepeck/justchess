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
	Likes        uint      `json:"likes"`
	GamesCount   uint      `json:"gamesCount"`
	BlitzRating  uint      `json:"blitzRating"`
	RapidRating  uint      `json:"rapidRating"`
	BulletRating uint      `json:"bulletRating"`
	RegisteredAt time.Time `json:"registeredAt"`
	IsDeleted    bool      `json:"isDeleted"`
	Role         Role      `json:"role"`
	password     string    `json:"-"`
}

func NewUser(id uuid.UUID) User {
	return User{
		Id:           id,
		Name:         "Guest-" + id.String()[:8],
		BlitzRating:  400,
		RapidRating:  400,
		BulletRating: 400,
		RegisteredAt: time.Now(),
	}
}

func (u User) GetPassword() string {
	return u.password
}
