package dbHelper

import (
	"github.com/Avi-0009/InstantWicket-backend/database"
	"github.com/Avi-0009/InstantWicket-backend/models"
)

func IsUserExist(phoneNo string) (bool, error) {

	var exist bool

	query := `SELECT EXISTS (SELECT 1 FROM users WHERE phone_no = $1 AND archived_at IS NULL)`

	err := database.DB.Get(&exist, query, phoneNo)

	return exist, err
}

func CreateUser(name, phoneNo, password string) error {

	query := `INSERT INTO users (name, phone_no, password)VALUES ($1, $2, $3)`
	_, err := database.DB.Exec(query, name, phoneNo, password)
	return err
}

func GetUserByPhoneNo(phoneNo string) (*models.User, error) {

	var user models.User
	query := `SELECT id, name, phone_no, password FROM users WHERE phone_no = $1 AND archived_at IS NULL`
	err := database.DB.Get(&user, query, phoneNo) // return exactly one row
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func CreateUserSession(userID string) (string, error) {

	var sessionID string

	query := `INSERT INTO user_sessions (user_id)VALUES ($1)RETURNING id`

	err := database.DB.Get(
		&sessionID,
		query,
		userID,
	)

	return sessionID, err
}

func DeleteUserSession(sessionID string) error {

	query := `UPDATE user_sessions SET archived_at = NOW() WHERE id = $1 AND archived_at IS NULL`
	_, err := database.DB.Exec(query, sessionID) // modifies the database but does not return any rows
	return err
}
