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
