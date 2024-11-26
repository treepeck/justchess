package repository

import (
	"log/slog"

	"justchess/pkg/db"
	"justchess/pkg/models/user"

	"github.com/google/uuid"
)

// FindUserById fetches user data by provided id.
func FindUserById(id uuid.UUID) *user.U {
	const queryText string = `
		SELECT
			id, name, is_deleted,
			games_count, blitz_rating, 
			bullet_rating, rapid_rating,
			registered_at, likes FROM users
		WHERE id = $1 
	`

	if db.DB == nil {
		slog.Error("db connection is closed.")
		return nil
	}

	rows, err := db.DB.Query(queryText, id.String())
	if err != nil {
		slog.Error("cannot execute query", "err", err)
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
	}
	return nil
}
