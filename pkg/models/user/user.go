package user

import (
	"time"

	"github.com/google/uuid"
)

// U stores all user data.
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

type Guest struct {
	Id          uuid.UUID `json:"id"`
	Name        string    `json:"username"`
	AccessToken string    `json:"accessToken"`
}

// NewGuest creates a new guest with random id.
func NewGuest() *Guest {
	id := uuid.New()
	return &Guest{
		Id:          id,
		Name:        "Guest-" + id.String()[0:8],
		AccessToken: "",
	}
}

func (u *U) GetPassword() string {
	return u.password
}
