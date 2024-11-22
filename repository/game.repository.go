package repository

import (
	"chess-api/db"
	"chess-api/models/game"
	"database/sql"
	"log/slog"

	"github.com/google/uuid"
)

func SaveGame(g *game.G) {
	const queryText = `
		INSERT INTO games (
			id, black_id, white_id,
			control, bonus, result,
			moves, played_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	var rows *sql.Rows
	rows, err := db.Pool.Query(queryText, g.Id,
		g.Black.Id, g.White.Id, g.Control, g.Bonus, g.Result,
		g.Moves, g.PlayedAt,
	)
	if err != nil || !rows.Next() {
		slog.Warn("error while writing a game", "err", err)
		return
	}
	rows.Close()
}

func FindGameById(id uuid.UUID) *game.G {
	const queryText = `
		SELECT *
		WHERE id = $1 
	`

	rows, err := db.Pool.Query(queryText, id.String())
	if err != nil {
		slog.Warn("cannot execute query", "err", err)
		return nil
	}
	defer rows.Close()

	var game game.G
	if rows.Next() {
		rows.Scan(&game.Id, &game.Black.Id, &game.White.Id, &game.Control,
			&game.Bonus, &game.Status, &game.Moves, &game.PlayedAt,
		)
		return &game
	}
	return nil
}
