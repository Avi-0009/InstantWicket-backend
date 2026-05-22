package dbHelper

import (
	"github.com/Avi-0009/InstantWicket-backend/database"
	"github.com/Avi-0009/InstantWicket-backend/models"
	"github.com/jmoiron/sqlx"
)

func CreateTeam(tx *sqlx.Tx, name, createdBy string) (string, error) {
	var teamID string
	query := `INSERT INTO teams (name, created_by) VALUES ($1, $2) RETURNING id`
	err := tx.Get(&teamID, query, name, createdBy)
	if err != nil {
		return "", err
	}
	return teamID, nil
}

func GetTeams() ([]models.Team, error) {
	var teams []models.Team
	query := `SELECT id, name, created_by, created_at, updated_at FROM teams ORDER BY created_at DESC`
	err := database.DB.Select(&teams, query)
	if err != nil {
		return nil, err
	}
	return teams, nil
}

func GetTeam(teamID string) (*models.Team, error) {
	var team models.Team

	query := `SELECT id, name, created_by, created_at, updated_at FROM teams ORDER BY created_at DESC`

	err := database.DB.Get(&team, query, teamID)

	if err != nil {
		return nil, err
	}
	return &team, nil
}

func UpdateTeam(teamID, name string) error {
	query := `UPDATE teams SET name = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := database.DB.Exec(query, name, teamID)
	return err
}
