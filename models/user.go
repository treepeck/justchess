package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
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

func (u *User) GetId() uuid.UUID {
	return u.Id
}

func (u *User) GetName() string {
	return u.Name
}

func (u *User) GetPassword() string {
	return u.password
}

type UserRepository interface {
	AddGuest(id uuid.UUID) *User
	Remove(u User)
	FindById(id uuid.UUID) User
	FindAll() []User
}
