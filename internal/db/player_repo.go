package db

import (
	"database/sql"
	"time"
)

// Player is a public player data commonly reused on multiple pages.
type Player struct {
	Id     string  `json:"i"`
	Name   string  `json:"n"`
	Rating float64 `json:"r"`
}

// Profile is used to fill up the player profile page.
type Profile struct {
	CreatedAt time.Time
	Id        string
	Name      string
	Rating    float64
	// TODO: Leaderboard place.
	Rank int
	// TODO: Rank confidence in %.
	RankConfidence int
	RatedGames     int
	EngineGames    int
}

type PlayerRepo interface {
	SelectById(id string) (Player, error)
	SelectProfile(id string) (Profile, error)
	// SelectLeaderboard selects [Profile] of 100 players with the biggest
	// rating sorted in descending order.
	SelectLeaderboard() ([]Profile, error)
}

type SQLPlayerRepo struct {
	pool *sql.DB
}

func NewSQLPlayerRepo(p *sql.DB) SQLPlayerRepo { return SQLPlayerRepo{pool: p} }

func (r SQLPlayerRepo) SelectById(id string) (Player, error) {
	row := r.pool.QueryRow(selectPlayerById, id)
	var p Player
	return p, row.Scan(&p.Id, &p.Name, &p.Rating)
}

func (r SQLPlayerRepo) SelectProfile(id string) (Profile, error) {
	row := r.pool.QueryRow(selectProfile, id)
	var p Profile
	return p, row.Scan(&p.Name, &p.Rating, &p.CreatedAt, &p.RatedGames, &p.EngineGames)
}

func (r SQLPlayerRepo) SelectLeaderboard() ([]Profile, error) {
	rows, err := r.pool.Query(selectLeaderboard)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	leaders := make([]Profile, 0, 20)
	for rows.Next() {
		var p Profile
		err = rows.Scan(&p.Id, &p.Name, &p.Rating, &p.CreatedAt, &p.RatedGames)
		if err != nil {
			return nil, err
		}
		leaders = append(leaders, p)
	}
	return leaders, err
}

const (
	selectPlayerById = `SELECT
	ID,
	NAME,
	RATING
FROM
	PLAYER
WHERE
	ID = $1`

	selectProfile = `SELECT
	P.NAME,
	P.RATING,
	P.CREATED_AT,
	COUNT(R.GAME_ID) AS RATED_GAMES,
	COUNT(E.GAME_ID) AS ENGINE_GAMES
FROM
	PLAYER P
	INNER JOIN RATED_GAME R ON (
		R.WHITE_ID = P.ID
		OR R.BLACK_ID = P.ID
	)
	INNER JOIN ENGINE_GAME E ON E.PLAYER_ID = P.ID
WHERE
	P.ID = $1
GROUP BY
	P.NAME,
	P.RATING,
	P.CREATED_AT`

	selectLeaderboard = `SELECT
	P.ID,
	P.NAME,
	P.RATING,
	P.CREATED_AT,
	COUNT(R.GAME_ID) AS NUM_OF_GAMES
FROM
	PLAYER P
	LEFT JOIN RATED_GAME R ON R.WHITE_ID = P.ID
	OR R.BLACK_ID = P.ID
GROUP BY
	P.ID,
	P.NAME,
	P.RATING,
	P.CREATED_AT
ORDER BY
	P.RATING DESC,
	NUM_OF_GAMES DESC
LIMIT
	100`
)
