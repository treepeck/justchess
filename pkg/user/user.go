// Package user provides the access to the user-related tables.
// See "schema.sql" for table details.
//
// All insert and delete operations are made using Transactions.
// It is a caller responsibility to end a transaction.
package user

import (
	"database/sql"
	"justchess/pkg/db"
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id           uuid.UUID `json:"id"`
	Name         string    `json:"username"`
	RegisteredAt time.Time `json:"registeredAt"`
	Mail         string    `json:"-"`
	PasswordHash string    `json:"-"`
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

// SelectUserByXXX functions may return an empty user.
// The caller must ensure that u.Id != uuid.Nil.

func SelectUserById(id string) (u User, err error) {
	query := "SELECT * FROM users WHERE id = $1;"
	rows, err := db.Pool.Query(query, id)
	if err != nil {
		return
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&u.Id, &u.Name, &u.PasswordHash, &u.RegisteredAt,
			&u.UpdatedAt, &u.Mail)
	}
	return
}

// SelectUserByLogin excepts login to be either the user_name or mail.
func SelectUserByLogin(login string) (u User, err error) {
	query := "SELECT * FROM users WHERE user_name = $1 OR mail = $1;"
	rows, err := db.Pool.Query(query, login)
	if err != nil {
		return
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&u.Id, &u.Name, &u.PasswordHash, &u.RegisteredAt,
			&u.UpdatedAt, &u.Mail)
	}
	return
}

func SelectUserByMail(mail string) (u User, err error) {
	query := "SELECT * FROM users WHERE mail = $1;"
	rows, err := db.Pool.Query(query, mail)
	if err != nil {
		return
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&u.Id, &u.Name, &u.PasswordHash, &u.RegisteredAt,
			&u.UpdatedAt, &u.Mail)
	}
	return
}

// IsTakenUsernameOrMail is a helper function to quickly check if the user_name and mail are unique.
func IsTakenUsernameOrMail(name, mail string) bool {
	query := "SELECT id FROM users WHERE user_name = $1 OR mail = $2;"
	rows, err := db.Pool.Query(query, name, mail)
	if err != nil {
		return false
	}
	defer rows.Close()

	return rows.Next()
}

// SelectTokenRegistration returns user registration info by token.
func SelectTokenRegistration(token string) (r Register, err error) {
	query := "SELECT mail, password_hash, user_name FROM tokenregistration WHERE token = $1\n" +
		"AND created_at >= NOW() - INTERVAL '20 minutes';"
	rows, err := db.Pool.Query(query, token)
	if err != nil {
		return
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&r.Mail, &r.Password, &r.Name)
	}
	return
}

// SelectTokenReset returns user id and password by token.
func SelectTokenReset(token string) (id, passwordHash string, err error) {
	query := "SELECT user_id, password_hash FROM tokenreset WHERE token = $1\n" +
		"AND created_at >= NOW() - INTERVAL '20 minutes';"
	rows, err := db.Pool.Query(query, token)
	if err != nil {
		return
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&id, &passwordHash)
	}
	return
}

///////////////////////////////////////////////////////////////
//                          INSERT                           //
///////////////////////////////////////////////////////////////

func InsertTokenRegistration(token string, r Register, tx *sql.Tx) error {
	query := "INSERT INTO tokenregistration (token, mail, password_hash, user_name)\n" +
		"VALUES ($1, $2, $3, $4);"
	_, err := tx.Exec(query, token, r.Mail, r.Password, r.Name)
	return err
}

func InsertUser(id string, r Register) error {
	query := "INSERT INTO users (id, mail, password_hash, user_name) \n" +
		"VALUES ($1, $2, $3, $4);"
	_, err := db.Pool.Exec(query, id, r.Mail, r.Password, r.Name)
	return err
}

func InsertTokenReset(token, userId, hash string, tx *sql.Tx) error {
	query := "INSERT INTO tokenreset (token, user_id, password_hash)\n" +
		"VALUES ($1, $2, $3);"
	_, err := tx.Exec(query, token, userId, hash)
	return err
}

///////////////////////////////////////////////////////////////
//                          UPDATE                           //
///////////////////////////////////////////////////////////////

func UpdatePasswordHash(hash, id string) (u User, err error) {
	query := "UPDATE users SET password_hash = $1 WHERE id = $2 RETURNING *;"
	rows, err := db.Pool.Query(query, hash, id)
	if err != nil {
		return
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&u.Id, &u.Name, &u.PasswordHash, &u.RegisteredAt,
			&u.UpdatedAt, &u.Mail)
	}
	return
}
