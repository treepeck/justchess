package repository

import (
	"chess-api/db"
	"chess-api/models/game"
	"log/slog"

	"github.com/google/uuid"
)

// func AddGame() *game.G {
// 	fn := slog.String("func", "AddGame")

// 	const queryText = `
// 		INSERT INTO games (
// 			id, black_id, white_id,
// 			control, bonus, status,
// 			moves, played_at,
// 		)
// 	`

// 	// defer rows.Close()
// 	// var game models.Game
// 	// if rows.Next() {
// 	// 	rows.Scan(&game.Id, &game.BlackId, &game.WhiteId,
// 	// 		&game.Control, &game.Bonus, &game.Status, &game.Moves,
// 	// 		&game.PlayedAt,
// 	// 	)
// 	// }
// 	// return &game
// 	return nil
// }

func FindGameById(id uuid.UUID) *game.G {
	fn := slog.String("func", "FindGameById")

	const queryText = `
		SELECT *
		WHERE id = $1 
	`

	rows, err := db.DB.Query(queryText, id.String())
	if err != nil {
		slog.Warn("cannot execute query", fn, "err", err)
		return nil
	}
	defer rows.Close()

	var game game.G
	if rows.Next() {
		rows.Scan(&game.Id, &game.BlackId, &game.WhiteId, &game.Control,
			&game.Bonus, &game.Status, &game.Moves, &game.PlayedAt,
		)
		return &game
	}
	return nil
}
