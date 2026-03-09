package db

import (
	"database/sql"
	"time"
)

// The following block of constants declares SQL queries used to access and
// modify the player-related tables.
const (
	insertPlayer = `
	INSERT INTO player (
		id,
		name,
		email,
		password_hash
	)
	VALUES (?, ?, ?, ?)`

	areNameAndEmailUnique = `SELECT COUNT(*) FROM player WHERE name = ? OR email = ?`

	selectPlayerById = `
	SELECT
		id, name, rating, rating_deviation, rating_volatility
	FROM player WHERE id = ?`

	selectCredentialsByEmail = `
	SELECT id, password_hash
	FROM player WHERE email = ?`

	selectIdAndNameByEmail = `
	SELECT id, name
	FROM player WHERE email = ?`

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

	updatePasswordHash = `UPDATE player SET password_hash = ? WHERE id = ?`

	insertSession = `
	INSERT INTO session (
		id,
		player_id
	)
	VALUES (?, ?)`

	selectSessionById = `
	SELECT * FROM session
	WHERE id = ?
	AND	expires_at > NOW()`

	selectSessionsByPlayerId = `
	SELECT * FROM session
	WHERE player_id = ?
	AND	expires_at > NOW()`

	deleteSessionById = `DELETE FROM session WHERE id = ?`

	insertSignupToken = `
	INSERT INTO signup_token (
		id, name, email, password_hash
	)
	VALUES (?, ?, ?, ?)`

	selectSignupToken = `
	SELECT name, email, password_hash
	FROM signup_token
	WHERE id = ? AND created_at >= NOW() - INTERVAL 15 MINUTE`

	deleteSignupToken = `DELETE FROM signup_token WHERE id = ?`

	insertPasswordResetToken = `
	INSERT INTO password_reset_token (
		id, player_id, new_password_hash
	)
	VALUES (?, ?, ?)`

	selectPasswordResetToken = `
	SELECT player_id, new_password_hash
	FROM password_reset_token
	WHERE id = ? AND created_at >= NOW() - INTERVAL 15 MINUTE`

	deletePasswordResetToken = `DELETE FROM password_reset_token WHERE id = ?`
)

// Player represents a registered player.
type Player struct {
	Id         string
	Name       string
	Rating     float64
	Deviation  float64
	Volatility float64
}

// Credentials is a password hash and id of a single player.
type Credentials struct {
	PasswordHash []byte
	Id           string
}

// ProfileData is a data object used to fill up the player.tmpl file while executing
// a template.
type ProfileData struct {
	CreatedAt  time.Time
	Name       string
	Rating     float64
	NumOfGames int
}

// Session is an authorization token for a player.  Each protected endpoint
// expects the Auth cookie to contain valid and not expired session.
type Session struct {
	CreatedAt time.Time
	ExpiresAt time.Time
	Id        string
	PlayerId  string
}

// SignupData represents the registration credentials.
type SignupData struct {
	PasswordHash []byte
	Name         string
	Email        string
}

// PlayerRepo wraps the database connection pool and provides methods to access
// and modify the player table.
type PlayerRepo struct {
	pool *sql.DB
}

func NewPlayerRepo(p *sql.DB) PlayerRepo { return PlayerRepo{pool: p} }

// Insert inserts a single record into the player table, using the provided
// credentials.
func (r PlayerRepo) Insert(id string, d SignupData) error {
	_, err := r.pool.Exec(insertPlayer, id, d.Name, d.Email, d.PasswordHash)
	return err
}

// SelectById selects all info of a single player by id.
func (r PlayerRepo) SelectById(id string) (Player, error) {
	row := r.pool.QueryRow(selectPlayerById, id)

	var p Player
	return p, row.Scan(&p.Id, &p.Name, &p.Rating, &p.Deviation, &p.Volatility)
}

// AreNameAndEmailUnique checks whether the provided name and email unique in
// the player table.
func (r PlayerRepo) AreNameAndEmailUnique(name, email string) (bool, error) {
	row := r.pool.QueryRow(areNameAndEmailUnique, name, email)
	var count int
	return count == 0, row.Scan(&count)
}

// SelectCredentialsByEmail selects [Credential] from the player table by email.
func (r PlayerRepo) SelectCredentialsByEmail(email string) (Credentials, error) {
	row := r.pool.QueryRow(selectCredentialsByEmail, email)

	var c Credentials
	return c, row.Scan(&c.Id, &c.PasswordHash)
}

// SelectIdAndNameByEmaild selects ONLY id and name from the player table.
func (r PlayerRepo) SelectIdAndNameByEmail(email string) (Player, error) {
	row := r.pool.QueryRow(selectIdAndNameByEmail, email)
	var p Player
	return p, row.Scan(&p.Id, &p.Name)
}

// SelectProfileData selects [ProfileData] from player and game tables
// using player's name.
func (r PlayerRepo) SelectProfileData(name string) (ProfileData, error) {
	row := r.pool.QueryRow(selectProfileData, name)
	var p ProfileData
	return p, row.Scan(&p.Name, &p.Rating, &p.CreatedAt, &p.NumOfGames)
}

// SelectLeaderboard selects [ProfileData] of 100 players with the biggest
// rating sorted in descending order.
func (r PlayerRepo) SelectLeaderboard() ([]ProfileData, error) {
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

// SelectBySessionId selects a single player with the specified session_id.
// Expired sessions are omitted.
func (r PlayerRepo) SelectBySessionId(id string) (Player, error) {
	row := r.pool.QueryRow(selectPlayerBySessionId, id)

	var p Player
	return p, row.Scan(&p.Id, &p.Name, &p.Rating, &p.Deviation, &p.Volatility)
}

// UpdateRatings updates rating, deviation, and volatility columns of two players
// in a single query.
func (r PlayerRepo) UpdateRatings(whiteId, blackId string,
	whiteRating, whiteDeviation, whiteVolatility,
	blackRating, blackDeviation, blackVolatility float64,
) error {
	_, err := r.pool.Exec(updateRatings,
		whiteId, whiteRating,
		blackId, blackRating,
		whiteId, whiteDeviation,
		blackId, blackDeviation,
		whiteId, whiteVolatility,
		blackId, blackVolatility,
		whiteId, blackId,
	)
	return err
}

// UpdatePasswordHash updates the password_hash column for the record with the
// given id.
func (r PlayerRepo) UpdatePasswordHash(id string, pwdHash []byte) error {
	_, err := r.pool.Exec(updatePasswordHash, pwdHash, id)
	return err
}

// InsertSession inserts a single record into the session table.
func (r PlayerRepo) InsertSession(id, playerId string) error {
	_, err := r.pool.Exec(insertSession, id, playerId)
	return err
}

// SelectByPlayerId selects a single record with the specified player_id
// from the session table.  Expired sessions are omitted.
func (r PlayerRepo) SelectSessionById(id string) (Session, error) {
	row := r.pool.QueryRow(selectSessionById, id)

	var s Session
	return s, row.Scan(&s.Id, &s.PlayerId, &s.CreatedAt, &s.ExpiresAt)
}

// SelectByPlayerId selects multiple records with the specified player_id
// from the session table.  Expired sessions are omitted.
func (r PlayerRepo) SelectSessionByPlayerId(playerId string) ([]Session, error) {
	rows, err := r.pool.Query(selectSessionsByPlayerId, playerId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session

	for rows.Next() {
		var s Session
		if err = rows.Scan(&s.Id, &s.PlayerId, &s.CreatedAt, &s.ExpiresAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

// DeleteSessionById delets single record with the same id as provided from the
// session table.
func (r PlayerRepo) DeleteSessionById(id string) error {
	_, err := r.pool.Exec(deleteSessionById, id)
	return err
}

// InsertSignupToken inserts a single record into the signup_token table.
func (r PlayerRepo) InsertSignupToken(id string, s SignupData) error {
	_, err := r.pool.Exec(insertSignupToken, id, s.Name, s.Email, s.PasswordHash)
	return err
}

// SelectSignupToken returns [SignupData] by the token id.
func (r PlayerRepo) SelectSignupToken(id string) (SignupData, error) {
	row := r.pool.QueryRow(selectSignupToken, id)
	var s SignupData
	return s, row.Scan(&s.Name, &s.Email, &s.PasswordHash)
}

// DeleteSignupToken deletes record with the specified id from the signup_token table.
func (r PlayerRepo) DeleteSignupToken(id string) error {
	_, err := r.pool.Exec(deleteSignupToken, id)
	return err
}

func (r PlayerRepo) InsertPasswordResetToken(id, playerId string, pwdHash []byte) error {
	_, err := r.pool.Exec(insertPasswordResetToken, id, playerId, pwdHash)
	return err
}

func (r PlayerRepo) SelectPasswordResetToken(id string) (Credentials, error) {
	row := r.pool.QueryRow(selectPasswordResetToken, id)
	var c Credentials
	return c, row.Scan(&c.Id, &c.PasswordHash)
}

func (r PlayerRepo) DeletePasswordResetToken(id string) error {
	_, err := r.pool.Exec(deletePasswordResetToken, id)
	return err
}
