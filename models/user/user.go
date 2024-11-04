package user

import (
	"chess-api/jwt_auth"
	"crypto/rand"
	"time"

	"github.com/google/uuid"
)

// U stores all user data.
type U struct {
	Id           uuid.UUID     `json:"id"`
	Name         string        `json:"username"`
	BlitzRating  uint          `json:"blitzRating"`
	RapidRating  uint          `json:"rapidRating"`
	BulletRating uint          `json:"bulletRating"`
	GamesCount   uint          `json:"gamesCount"`
	Likes        uint          `json:"likes"`
	RegisteredAt time.Time     `json:"registeredAt"`
	IsDeleted    bool          `json:"isDeleted"`
	Role         jwt_auth.Role `json:"-"`
	password     string
}

func NewUser(id uuid.UUID) *U {
	b := make([]byte, 32)
	rand.Read(b)
	return &U{
		Id:           id,
		Name:         "Guest-" + id.String()[0:8],
		BlitzRating:  400,
		RapidRating:  400,
		BulletRating: 400,
		GamesCount:   0,
		Likes:        0,
		RegisteredAt: time.Now(),
		IsDeleted:    false,
		Role:         jwt_auth.Guest,
		password:     string(b),
	}
}

type CreateUserDTO struct {
	Id uuid.UUID `json:"id"`
}

func (u *U) GetPassword() string {
	return u.password
}
