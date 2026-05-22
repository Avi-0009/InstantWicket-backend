package dbHelper

import (
	"github.com/Avi-0009/InstantWicket-backend/database"
	"github.com/Avi-0009/InstantWicket-backend/models"
	"github.com/jmoiron/sqlx"
)

func IsPlayerStatsExist(userID string) (bool, error) {

	var exist bool
	query := `SELECT EXISTS (SELECT 1 FROM player_stats WHERE user_id = $1)`
	err := database.DB.Get(&exist, query, userID)
	return exist, err
}

func CreatePlayerStats(tx *sqlx.Tx, userID string) error {
	query := `INSERT INTO player_stats (user_id, batting_style, bowling_style) VALUES ($1, 'Right Handed', 'Right Arm Fast')`
	_, err := tx.Exec(query, userID)
	return err
}

func AddGuest(name, phoneNo string) (string, error) {
	var userID string

	userQuery := `INSERT INTO users (name, phone_no, password) VALUES ($1, $2, 'guest_account') ON CONFLICT (phone_no) DO UPDATE SET updated_at = CURRENT_TIMESTAMP RETURNING id`
	err := database.DB.Get(&userID, userQuery, name, phoneNo)
	if err != nil {
		return "", err
	}

	var playerID string
	statsQuery := `INSERT INTO player_stats (user_id, batting_style, bowling_style) VALUES ($1, 'Right-hand bat', 'Right-arm medium') ON CONFLICT (user_id) DO UPDATE SET updated_at = CURRENT_TIMESTAMP RETURNING id`

	err = database.DB.Get(&playerID, statsQuery, userID)
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
		ps.id,
		ps.user_id,
		u.name, 
		ps.batting_style,
		ps.bowling_style,

		ps.career_matches,
		ps.career_matches_won,
		ps.career_matches_lost,

		ps.career_runs,
		ps.career_balls_faced,
		ps.career_innings_batted,
		ps.career_not_outs,
		ps.career_highest_score,
		ps.career_ducks,
		ps.career_golden_ducks,
		ps.career_fifties,
		ps.career_hundreds,
		ps.career_fours,
		ps.career_sixes,
		ps.strike_rate,

		ps.career_wickets,
		ps.career_balls_bowled,
		ps.career_runs_conceded,
		ps.career_maiden_overs,
		ps.career_wides,
		ps.career_no_balls,
		ps.career_best_bowling_wickets,
		ps.career_best_bowling_runs,
		ps.career_innings_bowled,
		ps.economy,

		ps.career_catches,
		ps.career_runouts,
		ps.career_stumpings,

		ps.career_total_points,
		ps.career_mvps,
		
		ps.created_at,
		ps.updated_at,
		ps.archived_at

	FROM player_stats ps
	JOIN users u ON ps.user_id = u.id
	WHERE ps.id = $1
`
	err := database.DB.Get(&playerStats, query, playerID)

	if err != nil {
		return nil, err
	}
	return &playerStats, nil
}

func GetAllPlayersStats() ([]models.PlayerStats, error) {
	var playerStats []models.PlayerStats
	query := `
		SELECT 
			player_stats.id,
			player_stats.user_id,
			player_stats.batting_style,
			player_stats.bowling_style,

			player_stats.career_matches,
			player_stats.career_matches_won,
			player_stats.career_matches_lost,

			player_stats.career_runs,
			player_stats.career_balls_faced,
			player_stats.career_innings_batted,
			player_stats.career_not_outs,
			player_stats.career_highest_score,
			player_stats.career_ducks,
			player_stats.career_golden_ducks,
			player_stats.career_fifties,
			player_stats.career_hundreds,
			player_stats.career_fours,
			player_stats.career_sixes,
			player_stats.strike_rate,

			player_stats.career_wickets,
			player_stats.career_balls_bowled,
			player_stats.career_runs_conceded,
			player_stats.career_maiden_overs,
			player_stats.career_wides,
			player_stats.career_no_balls,
			player_stats.career_best_bowling_wickets,
			player_stats.career_best_bowling_runs,
			player_stats.career_innings_bowled,
			player_stats.economy,

			player_stats.career_catches,
			player_stats.career_runouts,
			player_stats.career_stumpings,

			player_stats.career_total_points,
			player_stats.career_mvps,
			
			player_stats.created_at,
			player_stats.updated_at,
			player_stats.archived_at, 
            u.name

		FROM player_stats
		JOIN users u ON player_stats.user_id = u.id 
		WHERE player_stats.archived_at IS NULL
		ORDER BY player_stats.career_runs DESC
	`

	err := database.DB.Select(&playerStats, query)

	if err != nil {
		return nil, err
	}
	return playerStats, nil
}

func SearchPlayerStats(query string) ([]models.PlayerSearchResponse, error) {

	var playerStats []models.PlayerSearchResponse

	search := "%" + query + "%"

	sqlQuery := `	
				SELECT 
				    ps.id AS player_id,
				    ps.user_id, u.name, 
				    u.phone_no, 
				     ps.career_runs, 
				     ps.career_wickets 
				FROM player_stats AS ps
				JOIN users u ON ps.user_id = u.id
				WHERE LOWER(u.name) LIKE LOWER($1) OR u.phone_no LIKE $1
				ORDER BY ps.career_runs DESC
				LIMIT 10
				`
	err := database.DB.Select(&playerStats, sqlQuery, search)
	if err != nil {
		return nil, err
	}
	return playerStats, nil
}

func UpdatePlayerStats(userID, battingStyle, bowlingStyle string) error {
	query := `UPDATE player_stats SET batting_style = $1, bowling_style = $2, updated_at = CURRENT_TIMESTAMP WHERE user_id = $3`
	_, err := database.DB.Exec(query, battingStyle, bowlingStyle, userID)
	return err
}
