package dbHelper

import (
	"github.com/Avi-0009/InstantWicket-backend/database"
	"github.com/Avi-0009/InstantWicket-backend/models"
)

func IsPlayerStatsExist(userID string) (bool, error) {

	var exist bool
	query := `SELECT EXISTS (SELECT 1 FROM player_stats WHERE user_id = $1)`
	err := database.DB.Get(&exist, query, userID)
	return exist, err
}

func CreatePlayerStats(userID, battingStyle, bowlingStyle string) (string, error) {
	var playerID string
	query := `INSERT INTO player_stats (user_id, batting_style, bowling_style) VALUES ($1, $2, $3) RETURNING id`
	err := database.DB.Get(&playerID, query, userID, battingStyle, bowlingStyle)
	return playerID, err
}

func GetPlayerIDByUserID(userID string) (string, error) {
	var playerID string
	query := `SELECT id FROM player_stats WHERE user_id = $1`
	err := database.DB.Get(&playerID, query, userID)
	return playerID, err
}

func GetPlayerStatsByPlayerID(playerID string) (*models.PlayerStats, error) {
	var playerStats models.PlayerStats

	query := `
	SELECT 
		id,
		user_id,
		batting_style,
		bowling_style,

		career_matches,
		career_matches_won,
		career_matches_lost,

		career_runs,
		career_balls_faced,
		career_innings_batted,
		career_not_outs,
		career_highest_score,
		career_ducks,
		career_golden_ducks,
		career_fifties,
		career_hundreds,
		career_fours,
		career_sixes,
		strike_rate,

		career_wickets,
		career_balls_bowled,
		career_runs_conceded,
		career_maiden_overs,
		career_wides,
		career_no_balls,
		career_best_bowling_wickets,
		career_best_bowling_runs,
		career_innings_bowled,
		economy,

		career_catches,
		career_runouts,
		career_stumpings,

		career_total_points,
		career_mvps,
		
		created_at,
		updated_at,
		archived_at

	FROM player_stats
	WHERE id = $1
`
	err := database.DB.Get(&playerStats, query, playerID)

	if err != nil {
		return nil, err
	}
	return &playerStats, nil
}

func UpdatePlayerStats(userID, battingStyle, bowlingStyle string) error {
	query := `UPDATE player_stats SET batting_style = $1, bowling_style = $2, updated_at = CURRENT_TIMESTAMP WHERE user_id = $3`
	_, err := database.DB.Exec(query, battingStyle, bowlingStyle, userID)
	return err
}
