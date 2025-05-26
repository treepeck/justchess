package player

import (
	"database/sql"
	"errors"
	"justchess/pkg/db"
	"time"

	"github.com/google/uuid"
)

type Player struct {
	Id           uuid.UUID `json:"id"`
	Mail         string    `json:"-"`
	Name         string    `json:"username"`
	PasswordHash string    `json:"-"`
	IsEngine     bool      `json:"isEngine"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"-"`
}

type Register struct {
	Mail     string `json:"mail"`
	Name     string `json:"username"`
	Password string `json:"password"`
}

///////////////////////////////////////////////////////////////
//                          SELECT                           //
///////////////////////////////////////////////////////////////

func SelectPlayerById(id string) (p Player, err error) {
	query := "SELECT * FROM player WHERE id = $1;"
	return selectPlayer(query, id)
}

// SelectPlayerByLogin excepts login to be either the name or mail.
func SelectPlayerByLogin(login string) (p Player, err error) {
	query := "SELECT * FROM player WHERE name = $1 OR mail = $1 AND is_engine = false;"
	return selectPlayer(query, login)
}

func SelectPlayerByMail(mail string) (p Player, err error) {
	query := "SELECT * FROM player WHERE mail = $1 AND is_engine = false;"
	return selectPlayer(query, mail)
}

// IsTakenNameOrMail is a helper to quickly check are the name and mail unique.
func IsTakenNameOrMail(name, mail string) bool {
	query := "SELECT id FROM player WHERE name = $1 OR mail = $2 AND is_engine = false;"
	rows, err := db.Pool.Query(query, name, mail)
	if err != nil {
		return false
	}
	defer rows.Close()

	return rows.Next()
}

// SelectTokenRegistration returns user registration info by token.
func SelectTokenRegistration(token string) (r Register, err error) {
	query := "SELECT mail, password_hash, name FROM tokenregistration WHERE token = $1\n" +
		"AND created_at >= NOW() - INTERVAL '20 minutes';"
	rows, err := db.Pool.Query(query, token)
	if err != nil {
		return
	}
	defer rows.Close()

	if !rows.Next() {
		return r, errors.New("token not found")
	}
	err = rows.Scan(&r.Mail, &r.Password, &r.Name)
	return
}

// SelectTokenReset returns user id and password by token.
func SelectTokenReset(token string) (id, passwordHash string, err error) {
	query := "SELECT player_id, password_hash FROM tokenreset WHERE token = $1\n" +
		"AND created_at >= NOW() - INTERVAL '20 minutes';"
	rows, err := db.Pool.Query(query, token)
	if err != nil {
		return
	}
	defer rows.Close()

	if !rows.Next() {
		return id, passwordHash, errors.New("token not found")
	}
	err = rows.Scan(&id, &passwordHash)
	return
}

// selectPlayer is a helper and used to avoid code repetitions.
func selectPlayer(query, arg string) (p Player, err error) {
	rows, err := db.Pool.Query(query, arg)
	if err != nil {
		return
	}
	defer rows.Close()

	if !rows.Next() {
		return p, errors.New("player not found")
	}
	err = rows.Scan(&p.Id, &p.Mail, &p.Name, &p.PasswordHash, &p.IsEngine,
		&p.CreatedAt, &p.UpdatedAt)
	return
}

///////////////////////////////////////////////////////////////
//                          INSERT                           //
///////////////////////////////////////////////////////////////

func InsertTokenRegistration(token string, r Register, tx *sql.Tx) error {
	query := "INSERT INTO tokenregistration (token, mail, password_hash, name)\n" +
		"VALUES ($1, $2, $3, $4);"
	_, err := tx.Exec(query, token, r.Mail, r.Password, r.Name)
	return err
}

func InsertPlayer(id string, r Register) error {
	query := "INSERT INTO player (id, mail, password_hash, name)\n" +
		"VALUES ($1, $2, $3, $4);"
	_, err := db.Pool.Exec(query, id, r.Mail, r.Password, r.Name)
	return err
}

func InsertTokenReset(token, userId, hash string, tx *sql.Tx) error {
	query := "INSERT INTO tokenreset (token, player_id, password_hash)\n" +
		"VALUES ($1, $2, $3);"
	_, err := tx.Exec(query, token, userId, hash)
	return err
}

///////////////////////////////////////////////////////////////
//                          UPDATE                           //
///////////////////////////////////////////////////////////////

func UpdatePasswordHash(hash, id string) (p Player, err error) {
	query := "UPDATE player SET password_hash = $1 WHERE id = $2 RETURNING *;"
	rows, err := db.Pool.Query(query, hash, id)
	if err != nil {
		return
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&p.Id, &p.Mail, &p.Name, &p.PasswordHash, &p.IsEngine,
			&p.CreatedAt, &p.UpdatedAt)
	}
	return
}
