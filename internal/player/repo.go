package player

import (
	"database/sql"
	"time"
)

type Player struct {
	Id           string
	Name         string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type InsertPlayerDTO struct {
	Id           string
	Name         string
	Email        string
	PasswordHash string
}

const (
	createQuery = `CREATE TABLE IF NOT EXISTS player (
		id CHAR(8) PRIMARY KEY,
		name VARCHAR(60) NOT NULL UNIQUE,
		email VARCHAR(100) NOT NULL UNIQUE,
		password_hash VARCHAR(60) NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`

	insertQuery = `INSERT INTO player (id, name, email, password_hash) VALUES
	(?, ?, ?, ?)`

	selectByIdQuery = `SELECT * FROM player WHERE id = ?`
)

/*
Create creates the player table.
*/
func Create(pool *sql.DB) error {
	_, err := pool.Exec(createQuery)
	return err
}

/*
Insert inserts a single record into the player table, using the provided DTO.
*/
func Insert(pool *sql.DB, dto InsertPlayerDTO) error {
	_, err := pool.Exec(insertQuery, dto.Id, dto.Name, dto.Email, dto.PasswordHash)
	return err
}

func SelectById(pool *sql.DB, id string) {

}
