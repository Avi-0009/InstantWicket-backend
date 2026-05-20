package dbHelper

import "github.com/Avi-0009/InstantWicket-backend/database"

func CreateTeam(name, createdBy string) (string, error) {
	var teamID string
	query := `INSERT INTO matches (name, created_by) VALUES ($1, $2) RETURNING id`
	err := database.DB.Get(&teamID, query, name, createdBy)
	if err != nil {
		return "", err
	}
	return teamID, nil
}
