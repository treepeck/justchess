package repository

import (
	"chess-api/models"
	"log"

	"chess-api/db"

	"github.com/google/uuid"
)

func AddGuest(id uuid.UUID) *models.User {
	const queryText string = `
		INSERT INTO users (id, name, password)
		VALUES ($1, $2, $3)
		RETURNING id, name, is_deleted,
			games_count, blitz_rating, 
			bullet_rating, rapid_rating,
			registered_at, likes
	`

	// TODO: replace uuid.New() with the safer newUserpassword generation later
	name := "Guest-" + id.String()[0:8]
	rows, err := db.DB.Query(queryText, id, name, uuid.New())
	if err != nil {
		log.Println("Adduser: user cannot be created", err)
		return nil
	}

	defer rows.Close()
	var user models.User
	if rows.Next() {
		err = rows.Scan(&user.Id, &user.Name, &user.IsDeleted,
			&user.GamesCount, &user.BlitzRating, &user.BulletRating,
			&user.RapidRating, &user.RegisteredAt, &user.Likes,
		)
		if err != nil {
			log.Println(err)
		}
	}
	return &user
}

func FindById(id uuid.UUID) *models.User {
	const queryText string = `
		SELECT * FROM users
		WHERE id = $1 
	`

	rows, err := db.DB.Query(queryText, id.String())
	if err != nil {
		log.Println("FindById: ", err)
		return nil
	}

	defer rows.Close()
	var user models.User
	if rows.Next() {
		rows.Scan(&user.Id, &user.Name, &user.IsDeleted,
			&user.GamesCount, &user.BlitzRating, &user.BulletRating,
			&user.RapidRating, &user.RegisteredAt, &user.Likes,
		)
	}
	return &user
}
