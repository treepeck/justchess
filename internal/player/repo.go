package player

import (
	"database/sql"
	"time"
)

/*
Player represents a registered player.  Senstive data, such as password hash and
email will not be encoded into a JSON.
*/
type Player struct {
	Id           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"-"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type InsertPlayerDTO struct {
	Id           string
	Name         string
	Email        string
	PasswordHash string
}

const (
	createQuery = `CREATE TABLE IF NOT EXISTS player (
		id CHAR(12) PRIMARY KEY,
		name VARCHAR(60) NOT NULL UNIQUE,
		email VARCHAR(100) NOT NULL UNIQUE,
		password_hash VARCHAR(60) NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`

	insertQuery = `INSERT INTO player (id, name, email, password_hash) VALUES
	(?, ?, ?, ?)`

	selectByIdQuery = `SELECT * FROM player WHERE id = ?`

	selectByEmailQuery = `SELECT * FROM player WHERE email = ?`
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

/*
SelectById selects a single record with the same id as provided from the player
table.
*/
func SelectById(pool *sql.DB, id string) (Player, error) {
	row := pool.QueryRow(selectByIdQuery, id)

	var p Player
	err := row.Scan(&p.Id, &p.Name, &p.Email, &p.PasswordHash, &p.CreatedAt,
		&p.UpdatedAt)
	return p, err
}

/*
SelectByEmail selects a single record with the same email as provided from the
player table.
*/
func SelectByEmail(pool *sql.DB, email string) (Player, error) {
	row := pool.QueryRow(selectByEmailQuery, email)

	var p Player
	err := row.Scan(&p.Id, &p.Name, &p.Email, &p.PasswordHash, &p.CreatedAt,
		&p.UpdatedAt)
	return p, err
}
