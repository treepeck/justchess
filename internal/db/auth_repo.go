package db

import (
	"database/sql"
	"time"
)

// SignupData represents the registration credentials.
type SignupData struct {
	PasswordHash []byte
	Name         string
	Email        string
}

// Credentials is a password hash and id of a single player.
type Credentials struct {
	PasswordHash []byte
	Id           string
}

// Identity is id and name of a single player.
type Identity struct {
	Id   string
	Name string
}

// Session is an authorization token for a player.  Each protected endpoint
// expects the Auth cookie to contain valid and not expired session.
type Session struct {
	CreatedAt time.Time
	ExpiresAt time.Time
	Id        string
	PlayerId  string
}

// AuthRepo provides access to authorization and authentication data.
type AuthRepo interface {
	InsertGuest(id string) error
	InsertPlayer(id string, d SignupData) error
	IsEmailUnique(email string) (bool, error)
	SelectCredentialsByEmail(email string) (Credentials, error)
	SelectIdentityByEmail(email string) (Identity, error)
	// SelectPlayerBySessionId skips expired sessions.
	SelectPlayerBySessionId(id string) (Player, error)
	UpdatePasswordHash(id string, pwdHash []byte) error

	InsertSession(id, playerId string) error
	SelectSessionById(id string) (Session, error)
	SelectSessionsByPlayerId(id string) ([]Session, error)
	DeleteSession(id string) error

	InsertSignupToken(id string, d SignupData) error
	SelectSignupDataByToken(id string) (SignupData, error)
	DeleteSignupToken(id string) error

	InsertPasswordResetToken(id, playerId string, pwdHash []byte) error
	SelectCredentialsByResetToken(id string) (Credentials, error)
	DeletePasswordResetToken(id string) error
}

// SQLAuthRepo wraps the SQL database handle and implements [AuthRepo].
type SQLAuthRepo struct {
	pool *sql.DB
}

func NewSQLAuthRepo(p *sql.DB) SQLAuthRepo { return SQLAuthRepo{pool: p} }

func (r SQLAuthRepo) InsertGuest(id string) error {
	_, err := r.pool.Exec(insertGuest, id)
	return err
}

func (r SQLAuthRepo) InsertPlayer(id string, d SignupData) error {
	_, err := r.pool.Exec(insertPlayer, id, d.Name, d.Email, d.PasswordHash)
	return err
}

func (r SQLAuthRepo) IsEmailUnique(email string) (bool, error) {
	row := r.pool.QueryRow(isEmailUnique, email)
	var count int
	return count == 0, row.Scan(&count)
}

func (r SQLAuthRepo) SelectCredentialsByEmail(email string) (Credentials, error) {
	row := r.pool.QueryRow(selectCredentialsByEmail, email)
	var c Credentials
	return c, row.Scan(&c.Id, &c.PasswordHash)
}

func (r SQLAuthRepo) SelectIdentityByEmail(email string) (Identity, error) {
	row := r.pool.QueryRow(selectIdentityByEmail, email)
	var i Identity
	return i, row.Scan(&i.Id, &i.Name)
}

func (r SQLAuthRepo) SelectPlayerBySessionId(id string) (Player, error) {
	row := r.pool.QueryRow(selectPlayerBySessionId, id)
	var p Player
	return p, row.Scan(
		&p.Id, &p.Name, &p.Rating, &p.Deviation, &p.Volatility, &p.IsGuest,
	)
}

func (r SQLAuthRepo) UpdatePasswordHash(id string, pwdHash []byte) error {
	_, err := r.pool.Exec(updatePasswordHash, pwdHash, id)
	return err
}

func (r SQLAuthRepo) InsertSession(id, playerId string) error {
	_, err := r.pool.Exec(insertSession, id, playerId)
	return err
}

func (r SQLAuthRepo) SelectSessionById(id string) (Session, error) {
	row := r.pool.QueryRow(selectSessionById, id)
	var s Session
	return s, row.Scan(&s.Id, &s.PlayerId, &s.CreatedAt, &s.ExpiresAt)
}

func (r SQLAuthRepo) SelectSessionsByPlayerId(id string) ([]Session, error) {
	rows, err := r.pool.Query(selectSessionsByPlayerId, id)
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

func (r SQLAuthRepo) DeleteSession(id string) error {
	_, err := r.pool.Exec(deleteSession, id)
	return err
}

func (r SQLAuthRepo) InsertSignupToken(id string, d SignupData) error {
	_, err := r.pool.Exec(insertSignupToken, id, d.Name, d.Email, d.PasswordHash)
	return err
}

func (r SQLAuthRepo) SelectSignupDataByToken(id string) (SignupData, error) {
	row := r.pool.QueryRow(selectSignupDataByToken, id)
	var s SignupData
	return s, row.Scan(&s.Name, &s.Email, &s.PasswordHash)
}

func (r SQLAuthRepo) DeleteSignupToken(id string) error {
	_, err := r.pool.Exec(deleteSignupToken, id)
	return err
}

func (r SQLAuthRepo) InsertPasswordResetToken(id, playerId string, pwdHash []byte) error {
	_, err := r.pool.Exec(insertPasswordResetToken, id, playerId, pwdHash)
	return err
}

func (r SQLAuthRepo) SelectCredentialsByResetToken(id string) (Credentials, error) {
	row := r.pool.QueryRow(selectCredentialsByResetToken, id)
	var c Credentials
	return c, row.Scan(&c.Id, &c.PasswordHash)
}

func (r SQLAuthRepo) DeletePasswordResetToken(id string) error {
	_, err := r.pool.Exec(deletePasswordResetToken, id)
	return err
}

const (
	insertGuest = `
	INSERT INTO player (id, name, is_guest)
	VALUES (?, 'Guest', TRUE)`

	insertPlayer = `
	INSERT INTO player (id, name, email, password_hash, is_guest)
	VALUES (?, ?, ?, ?, FALSE)`

	isEmailUnique = `SELECT COUNT(*) FROM player WHERE email = ?`

	selectCredentialsByEmail = `
	SELECT id, password_hash FROM player WHERE email = ?`

	selectIdentityByEmail = `
	SELECT id, name	FROM player WHERE email = ?`

	selectPlayerBySessionId = `
	SELECT
		p.id, p.name, p.rating, p.rating_deviation,
		p.rating_volatility, p.is_guest
	FROM player p
	INNER JOIN session s
	ON p.id = s.player_id
	WHERE s.id = ? AND s.expires_at > NOW()`

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

	deleteSession = `DELETE FROM session WHERE id = ?`

	insertSignupToken = `
	INSERT INTO signup_token (
		id, name, email, password_hash
	)
	VALUES (?, ?, ?, ?)`

	selectSignupDataByToken = `
	SELECT name, email, password_hash
	FROM signup_token
	WHERE id = ? AND created_at >= NOW() - INTERVAL 15 MINUTE`

	deleteSignupToken = `DELETE FROM signup_token WHERE id = ?`

	insertPasswordResetToken = `
	INSERT INTO password_reset_token (
		id, player_id, new_password_hash
	)
	VALUES (?, ?, ?)`

	selectCredentialsByResetToken = `
	SELECT player_id, new_password_hash
	FROM password_reset_token
	WHERE id = ? AND created_at >= NOW() - INTERVAL 15 MINUTE`

	deletePasswordResetToken = `DELETE FROM password_reset_token WHERE id = ?`
)
