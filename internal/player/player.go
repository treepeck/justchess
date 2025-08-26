package player

import (
	"database/sql"
	"net/http"
)

/*
PlayerService wraps a database connection pool and provides methods for handling
player-related HTTP requests.
*/
type PlayerService struct {
	pool *sql.DB
}

/*
InitPlayerService creates a new [PlayerService], initializes the player table and
adds routes to the specified mux.
*/
func InitPlayerService(pool *sql.DB, mux *http.ServeMux) error {
	// s := PlayerService{pool: pool}

	// Initializing session table.
	if _, err := pool.Exec(createQuery); err != nil {
		return err
	}

	return nil
}
