package dbHelper

import (
	"github.com/Avi-0009/InstantWicket-backend/database"
	"github.com/Avi-0009/InstantWicket-backend/models"
)

func CreateMatch(input models.CreateMatch, createdBy string) (string, error) {
	var matchID string
	query := `INSERT INTO matches (
                     team_a_id,
                     team_b_id,
                     toss_winner_team_id,
                     toss_decision,
                     allow_commom_player,
                     allow_solo_player,
                     over_limit,
                     umpire_id,
                     created_by
                     ) VALUES (
                               $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
                     )
                     RETURNING id
`
	err := database.DB.Get(
		&matchID,
		query,

		input.TeamAID,
		input.TeamBID,

		input.TossWinnerTeamID,
		input.TossDecision,

		input.AllowCommonPlayer,
		input.AllowSoloBatting,

		input.OversLimit,

		input.UmpireID,

		createdBy,
	)
	if err != nil {
		return "", err
	}
	return matchID, nil
}

func GetMatches() ([]models.Match, error) {
	var matches []models.Match

	query := `SELECT id, team_a_id, team_b_id, toss_winner_team_id, toss_decision, allow_commom_player, allow_solo_batting, over_limit, status, winner_team_id, man_of_match, worst_player, umpire_id 
			FROM matches
			ORDER BY created_at DESC`

	err := database.DB.Select(&matches, query)

	if err != nil {
		return nil, err
	}
	return matches, nil
}

func GetMatchByID(matchID string) (*models.Match, error) {
	var match models.Match
	query := `SELECT id, team_a_id, team_b_id, toss_winner_team_id, toss_decision, allow_commom_player, allow_solo_batting, over_limit, status, winner_team_id, man_of_match, worst_player, umpire_id, created_by, created_at, updated_at
			FROM matches
			WHERE id = $1`
	err := database.DB.Get(&match, query, matchID)
	if err != nil {
		return nil, err
	}
	return &match, nil
}
