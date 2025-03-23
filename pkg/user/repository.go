package user

import (
	"justchess/pkg/db"
)

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

func insertUser(u User) error {
	query := "INSERT INTO users (id, user_name, password_hash, mail)\n" +
		"VALUES ($1, $2, $3, $4);"
	_, err := db.Pool.Exec(query, u.Id, u.Name, u.PasswordHash, u.Mail)
	return err
}

func insertUnverified(r Register) (id string, err error) {
	query := "INSERT INTO unverified (mail, password_hash, user_name)\n" +
		"VALUES ($1, $2, $3) RETURNING id;"
	rows, err := db.Pool.Query(query, r.Mail, r.Password, r.Name)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&id)
	}
	return
}

func isUniqueNameAndMail(name, mail string) bool {
	query := "SELECT * FROM users WHERE user_name = $1 OR mail = $2;"
	rows, err := db.Pool.Query(query, name, mail)
	if err != nil {
		return false
	}
	defer rows.Close()

	return !rows.Next()
}

func isUnverifiedId(id string) bool {
	query := "SELECT * FROM unverified WHERE id = $1;"
	rows, err := db.Pool.Query(query, id)
	if err != nil {
		return false
	}
	defer rows.Close()

	return rows.Next()
}

// deleteUnverified deletes the record with the specified id and returns the record's data.
func deleteUnverified(id string) (u User, err error) {
	query := "DELETE FROM unverified WHERE id = $1 RETURNING *;"
	rows, err := db.Pool.Query(query, id)
	if err != nil {
		return
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&u.Id, &u.Mail, &u.PasswordHash, &u.Name)
	}
	return
}
