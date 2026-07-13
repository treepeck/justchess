package db

import (
	"database/sql"
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

// AuthRepo provides access to authorization and authentication data.
type AuthRepo interface {
	InsertPlayer(id string, d SignupData) error
	IsEmailUnique(email string) (bool, error)
	SelectCredentialsByEmail(email string) (Credentials, error)
	SelectIdentityByEmail(email string) (Identity, error)
	UpdatePasswordHash(id string, pwdHash []byte) error

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

func (r SQLAuthRepo) UpdatePasswordHash(id string, pwdHash []byte) error {
	_, err := r.pool.Exec(updatePasswordHash, pwdHash, id)
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
	insertPlayer = `INSERT INTO player (id, name, email, password_hash)	VALUES ($1, $2, $3, $4)`

	isEmailUnique = `SELECT COUNT(*) FROM player WHERE email = $1`

	selectCredentialsByEmail = `SELECT id, password_hash FROM player WHERE email = $1`

	selectIdentityByEmail = `SELECT id, name FROM player WHERE email = $1`

	updatePasswordHash = `UPDATE player SET password_hash = $1 WHERE id = $2`

	insertSignupToken = `INSERT INTO signup_token (id, name, email, password_hash) VALUES ($1, $2, $3, $4)`

	selectSignupDataByToken = `SELECT name, email, password_hash
	FROM signup_token WHERE id = $1 AND created_at >= NOW() - INTERVAL '15 MINUTES'`

	deleteSignupToken = `DELETE FROM signup_token WHERE id = $1`

	insertPasswordResetToken = `INSERT INTO password_reset_token (
		id, player_id, new_password_hash
	)
	VALUES ($1, $2, $3)`

	selectCredentialsByResetToken = `SELECT player_id, new_password_hash
	FROM password_reset_token WHERE id = $1 AND created_at >= NOW() - INTERVAL '15 MINUTES'`

	deletePasswordResetToken = `DELETE FROM password_reset_token WHERE id = $1`
)
