package user

import (
	"time"

	"github.com/google/uuid"
)

// U describes user data.
type U struct {
	Id           uuid.UUID `json:"id"`
	Name         string    `json:"username"`
	BlitzRating  uint      `json:"blitzRating"`
	RapidRating  uint      `json:"rapidRating"`
	BulletRating uint      `json:"bulletRating"`
	GamesCount   uint      `json:"gamesCount"`
	Likes        uint      `json:"likes"`
	RegisteredAt time.Time `json:"registeredAt"`
	IsDeleted    bool      `json:"isDeleted"`
	password     string
}

type CreateUserDTO struct {
	Id uuid.UUID `json:"id"`
}

func (u *U) GetPassword() string {
	return u.password
}
