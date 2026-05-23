package dbHelper

import (
	"github.com/Avi-0009/InstantWicket-backend/database"
	"github.com/Avi-0009/InstantWicket-backend/models"
	"github.com/jmoiron/sqlx"
)

func IsUserExist(phoneNo string) (bool, error) {

	var exist bool

	query := `SELECT EXISTS (SELECT 1 FROM users WHERE phone_no = $1 AND archived_at IS NULL)`

	err := database.DB.Get(&exist, query, phoneNo)

	return exist, err
}

func CreateUser(tx *sqlx.Tx, name, phoneNo, password string) (string, error) {
	var userID string
	query := `INSERT INTO users (name, phone_no, password)VALUES ($1, $2, $3) RETURNING id`
	err := tx.Get(&userID, query, name, phoneNo, password)
	return userID, err
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

func UpdatePassword(phone string, password string) (string, error) {

	var userID string

	query := `UPDATE users SET password = $1, updated_at = NOW()
		WHERE
			phone_no = $2
			AND archived_at IS NULL
		RETURNING id
	`
	err := database.DB.Get(&userID, query, password, phone)
	return userID, err
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
	_, err := database.DB.Exec(query, sessionID)
	return err
}

func DeleteAllUserSessions(userID string) error {
	query := `DELETE FROM user_sessions SET archived_at = NOW() WHERE user_id = $1 AND archived_at IS NULL`
	_, err := database.DB.Exec(query, userID)
	return err
}
