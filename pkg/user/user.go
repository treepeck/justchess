// Package user provides the access to the 'users' and 'unverified' db tables.
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

// Used for sign in/up.
type Register struct {
	Mail     string `json:"mail"`
	Name     string `json:"username"`
	Password string `json:"password"`
}

///////////////////////////////////////////////////////////////
//                          SELECT                           //
///////////////////////////////////////////////////////////////

// SelectById may return an empty user. The caller must ensure that u.Id != uuid.Nil.
func SelectById(id string) (u User, err error) {
	query := "SELECT * FROM users WHERE id = $1;"
	rows, err := db.Pool.Query(query, id)
	if err != nil {
		return
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&u.Id, &u.Name, &u.PasswordHash, &u.RegisteredAt, &u.UpdatedAt, &u.Mail)
	}
	return
}

func IsTakenUsernameOrMail(name, mail string) bool {
	query := "SELECT id FROM users WHERE user_name = $1 OR mail = $2;"
	rows, err := db.Pool.Query(query, name, mail)
	if err != nil {
		return false
	}
	defer rows.Close()

	return rows.Next()
}

func IsUnverifiedId(id string) bool {
	query := "SELECT user_name FROM unverified WHERE id = $1;"
	rows, err := db.Pool.Query(query, id)
	if err != nil {
		return false
	}
	defer rows.Close()

	return rows.Next()
}

///////////////////////////////////////////////////////////////
//                          INSERT                           //
///////////////////////////////////////////////////////////////

// InsertUnverified returns the new unverified user ID.
func InsertUnverified(r Register, tx *sql.Tx) (id string, err error) {
	query := "INSERT INTO unverified (mail, password_hash, user_name)\n" +
		"VALUES ($1, $2, $3) RETURNING id;"
	rows, err := tx.Query(query, r.Mail, r.Password, r.Name)
	if err != nil {
		return
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&id)
	}
	return
}

// InsertUser may return an empty user. The called must ensure that u.Id != uuid.Nil.
func InsertUser(id string, r Register, tx *sql.Tx) (u User, err error) {
	query := "INSERT INTO users (id, mail, password_hash, user_name)\n" +
		"VALUES ($1, $2, $3, $4) RETURNING *;"
	rows, err := tx.Query(query, id, r.Mail, r.Password, r.Name)
	if err != nil {
		return
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&u.Id, &u.Name, &u.PasswordHash, &u.RegisteredAt,
			&u.UpdatedAt, &u.Mail,
		)
	}
	return
}

///////////////////////////////////////////////////////////////
//                          DELETE                           //
///////////////////////////////////////////////////////////////

// DeleteUnverified returns the deleted record data.
func DeleteUnverified(id string, tx *sql.Tx) (r Register, err error) {
	query := "DELETE FROM unverified WHERE id = $1 RETURNING mail, user_name, password_hash;"
	rows, err := tx.Query(query, id)
	if err != nil {
		return
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&r.Mail, &r.Name, &r.Password)
	}
	return
}
