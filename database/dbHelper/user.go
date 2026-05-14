package dbHelper

import (
	"github.com/Avi-0009/InstantWicket-backend/database"
)

func IsUserExist(phoneNo string) (bool, error) {

	var exist bool

	query := `
		SELECT EXISTS (
			SELECT 1
			FROM users
			WHERE phone_no = $1
			AND archived_at IS NULL
		)
	`

	err := database.Todo.Get(
		&exist,
		query,
		phoneNo,
	)

	return exist, err
}

func CreateUser(
	name string,
	phoneNo string,
	password string,
) error {

	query := `
		INSERT INTO users (
			name,
			phone_no,
			password
		)
		VALUES ($1, $2, $3)
	`

	_, err := database.Todo.Exec(
		query,
		name,
		phoneNo,
		password,
	)

	return err
}
