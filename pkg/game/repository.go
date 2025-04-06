package game

import (
	"errors"
	"justchess/pkg/chess"
	"justchess/pkg/db"
)

func SelectById(id string) (g chess.Game, err error) {
	query := "SELECT * FROM game WHERE id = $1;"
	rows, err := db.Pool.Query(query, id)
	if err != nil {
		return
	}

	if !rows.Next() {
		return g, errors.New("game not found")
	}
	err = rows.Scan(&g.Id, &g.WhiteId, &g.BlackId, &g.TimeControl, &g.TimeBonus,
		&g.Result, &g.Winner, &g.Moves, &g.Mode)
	return
}

func Insert(g chess.Game) error {
	query := "INSERT INTO game (id, white_id, black_id,\n" +
		"time_control, time_bonus, result, winner, moves, mode)\n" +
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);"
	_, err := db.Pool.Exec(query, g.Id, g.WhiteId, g.BlackId, g.TimeControl,
		g.TimeBonus, g.Result, g.Winner, g.Moves, g.Mode)
	return err
}
