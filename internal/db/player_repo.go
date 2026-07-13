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
	selectPlayerById = `SELECT id, name, rating FROM player WHERE id = $1`

	selectProfile = `SELECT
		p.name,
		p.rating,
		p.created_at,
		count(r.game_id) as rated_games,
        count(e.game_id) as engine_games
	FROM player p
	INNER JOIN rated_game r ON (r.white_id = p.id OR r.black_id = p.id)
    INNER JOIN engine_game e ON e.player_id = p.id
	WHERE p.id = $1
	GROUP BY p.name, p.rating, p.created_at`

	selectLeaderboard = `SELECT
		p.id,
		p.name,
	    p.rating,
	    p.created_at,
	    count(r.game_id) as num_of_games
	FROM player p
	LEFT JOIN rated_game r
	ON r.white_id = p.id OR r.black_id = p.id
	GROUP BY p.id, p.name, p.rating, p.created_at
	ORDER BY p.rating DESC, num_of_games DESC
	LIMIT 100`
)
