package repository

import (
	"chess-api/models/user"
	"log/slog"

	"chess-api/db"

	"github.com/google/uuid"
)

func AddGuest(id uuid.UUID) *user.U {
	fn := slog.String("func", "AddGuest")

	const queryText = `
		INSERT INTO users (id, name, password)
		VALUES ($1, $2, $3)
		RETURNING id, name, is_deleted,
			games_count, blitz_rating, 
			bullet_rating, rapid_rating,
			registered_at, likes
	`

	// TODO: replace uuid.New() with the safer password generation later
	name := "Player-" + id.String()[0:8]
	rows, err := db.DB.Query(queryText, id, name, uuid.New())
	if err != nil {
		slog.Warn("user cannot be created", fn, "err", err)
		return nil
	}

	defer rows.Close()
	var user user.U
	if rows.Next() {
		rows.Scan(&user.Id, &user.Name, &user.IsDeleted,
			&user.GamesCount, &user.BlitzRating, &user.BulletRating,
			&user.RapidRating, &user.RegisteredAt, &user.Likes,
		)
	}
	return &user
}

func FindUserById(id uuid.UUID) *user.U {
	fn := slog.String("func", "FindUserById")

	const queryText string = `
		SELECT
			id, name, is_deleted,
			games_count, blitz_rating, 
			bullet_rating, rapid_rating,
			registered_at, likes FROM users
		WHERE id = $1 
	`

	rows, err := db.DB.Query(queryText, id.String())
	if err != nil {
		slog.Warn("cannot execute query", fn, "err", err)
		return nil
	}
	defer rows.Close()

	var user user.U
	if rows.Next() {
		rows.Scan(&user.Id, &user.Name, &user.IsDeleted,
			&user.GamesCount, &user.BlitzRating, &user.BulletRating,
			&user.RapidRating, &user.RegisteredAt, &user.Likes,
		)
		return &user
	} else {
		return nil
	}
}
