package dbHelper

import (
	"github.com/Avi-0009/InstantWicket-backend/database"
	"github.com/Avi-0009/InstantWicket-backend/models"
	"github.com/jmoiron/sqlx"
)

func CreateMatch(tx *sqlx.Tx, input models.CreateMatch, createdBy string) (string, error) {
	var matchID string
	var dbUmpireID interface{} = input.UmpireID
	if input.UmpireID == "" {
		dbUmpireID = nil
	}
	query := `INSERT INTO matches (
                     team_a_id,
                     team_b_id,
                     toss_winner_team_id,
                     toss_decision,
                     allow_common_player,
                     allow_solo_batting,
                     overs_limit,
                     umpire_id,
                     created_by
                     ) VALUES (
                               $1, $2, $3, $4, $5, $6, $7, $8, $9
                     )
                     RETURNING id
`
	err := tx.Get(&matchID, query, input.TeamAID, input.TeamBID, input.TossWinnerTeamID, input.TossDecision, input.AllowCommonPlayer, input.AllowSoloBatting, input.OversLimit, dbUmpireID, createdBy)
	if err != nil {
		return "", err
	}
	return matchID, nil
}

func GetMatches() ([]models.Match, error) {
	var matches []models.Match

	query := `SELECT 
			m.id, 
			m.team_a_id, 
			tA.name AS team_a_name, 
			m.team_b_id, 
			tB.name AS team_b_name, 
			m.toss_winner_team_id, 
			m.toss_decision, 
			m.allow_common_player, 
			m.allow_solo_batting, 
			m.overs_limit, 
			m.status, 
			m.winner_team_id, 
			m.man_of_match, 
			m.worst_player, 
			m.umpire_id,
			m.created_by,
			m.created_at,
			m.updated_at
		FROM matches m
		JOIN teams tA ON m.team_a_id = tA.id
		JOIN teams tB ON m.team_b_id = tB.id
		ORDER BY m.created_at DESC`

	err := database.DB.Select(&matches, query)
	if err != nil {
		return nil, err
	}
	return matches, nil
}

func GetMatchByID(matchID string) (*models.Match, error) {
	var match models.Match
	query := `SELECT m.id, m.team_a_id, tA.name AS team_a_name, m.team_b_id, tB.name AS team_b_name, m.toss_winner_team_id, m.toss_decision, m.allow_common_player, m.allow_solo_batting, m.overs_limit, m.status, m.winner_team_id, m.man_of_match, m.worst_player, m.umpire_id, m.created_by, m.created_at, m.updated_at
			FROM matches m
			JOIN teams tA ON m.team_a_id = tA.id
			JOIN teams tB ON m.team_b_id = tB.id
			WHERE m.id = $1`
	err := database.DB.Get(&match, query, matchID)
	if err != nil {
		return nil, err
	}
	return &match, nil
}
