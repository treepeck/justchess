package db

import (
	"database/sql"
	"time"
)

// Player represents a registered player.
type Player struct {
	Id         string
	Name       string
	Rating     float64
	Deviation  float64
	Volatility float64
}

// ProfileData is a data object used to fill up the player.tmpl file while executing
// a template.
type ProfileData struct {
	CreatedAt  time.Time
	Name       string
	Rating     float64
	NumOfGames int
}

// RatingUpdate is used to update the player's rating after completed game.
type RatingUpdate struct {
	Id         string
	Rating     float64
	Deviation  float64
	Volatility float64
}

type PlayerRepo interface {
	SelectById(id string) (Player, error)
	SelectProfileData(name string) (ProfileData, error)
	// SelectLeaderboard selects [ProfileData] of 100 players with the biggest
	// rating sorted in descending order.
	SelectLeaderboard() ([]ProfileData, error)
	// SelectBySessionId skips expired sessions.
	SelectBySessionId(id string) (Player, error)
	UpdateRatings(white, black RatingUpdate) error
}

// SQLPlayerRepo wraps the database connection pool and implements [PlayerRepo].
type SQLPlayerRepo struct {
	pool *sql.DB
}

func NewSQLPlayerRepo(p *sql.DB) SQLPlayerRepo { return SQLPlayerRepo{pool: p} }

func (r SQLPlayerRepo) SelectById(id string) (Player, error) {
	row := r.pool.QueryRow(selectPlayerById, id)

	var p Player
	return p, row.Scan(&p.Id, &p.Name, &p.Rating, &p.Deviation, &p.Volatility)
}

func (r SQLPlayerRepo) SelectProfileData(name string) (ProfileData, error) {
	row := r.pool.QueryRow(selectProfileData, name)
	var p ProfileData
	return p, row.Scan(&p.Name, &p.Rating, &p.CreatedAt, &p.NumOfGames)
}

func (r SQLPlayerRepo) SelectLeaderboard() ([]ProfileData, error) {
	rows, err := r.pool.Query(selectLeaderboard)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	leaders := make([]ProfileData, 0, 20)
	for rows.Next() {
		var pd ProfileData
		if err = rows.Scan(
			&pd.Name,
			&pd.Rating,
			&pd.CreatedAt,
			&pd.NumOfGames,
		); err != nil {
			return nil, err
		}
		leaders = append(leaders, pd)
	}
	return leaders, err
}

func (r SQLPlayerRepo) SelectBySessionId(id string) (Player, error) {
	row := r.pool.QueryRow(selectPlayerBySessionId, id)

	var p Player
	return p, row.Scan(&p.Id, &p.Name, &p.Rating, &p.Deviation, &p.Volatility)
}

func (r SQLPlayerRepo) UpdateRatings(white, black RatingUpdate) error {
	_, err := r.pool.Exec(updateRatings,
		white.Id, white.Rating,
		black.Id, black.Rating,
		white.Id, white.Deviation,
		black.Id, black.Deviation,
		white.Id, white.Volatility,
		black.Id, black.Volatility,
		white.Id, black.Id,
	)
	return err
}

const (
	selectPlayerById = `
	SELECT
		id, name, rating, rating_deviation, rating_volatility
	FROM player WHERE id = ?`

	selectProfileData = `
	SELECT
		p.name,
		p.rating,
		p.created_at,
		count(g.id) as num_of_games
	FROM player p
	LEFT JOIN game g
	ON
		(g.white_id = p.id OR g.black_id = p.id)
		AND g.termination != 1
	WHERE p.name = ?
	GROUP BY p.name, p.rating, p.created_at`

	selectLeaderboard = `
	SELECT
		p.name,
	    p.rating,
	    p.created_at,
	    count(g.id) as num_of_games
	FROM player p
	LEFT JOIN game g
	ON
		(g.white_id = p.id OR g.black_id = p.id)
	    AND g.termination != 1
	GROUP BY p.name, p.rating, p.created_at
	ORDER BY p.rating DESC, num_of_games DESC
	LIMIT 100`

	selectPlayerBySessionId = `
	SELECT
		p.id, p.name, p.rating, p.rating_deviation, p.rating_volatility
	FROM player p
	INNER JOIN session s
	ON p.id = s.player_id
	WHERE s.id = ? AND s.expires_at > NOW()`

	updateRatings = `
	UPDATE player
	SET
		rating = CASE
			WHEN id = ? THEN ?
			WHEN id = ? THEN ?
			ELSE rating
		END,
		rating_deviation = CASE
			WHEN id = ? THEN ?
			WHEN id = ? THEN ?
			ELSE rating_deviation
		END,
		rating_volatility = CASE
			WHEN id = ? THEN ?
			WHEN id = ? THEN ?
			ELSE rating_volatility
		END
	WHERE player.id = ? OR player.id = ?`
)
