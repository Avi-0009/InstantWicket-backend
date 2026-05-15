package dbHelper

import "github.com/Avi-0009/InstantWicket-backend/database"

func IsPlayerStatsExist(userID string) (bool, error) {

	var exist bool

	query := `
		SELECT EXISTS (
			SELECT 1
			FROM player_stats
			WHERE user_id = $1
		)
	`

	err := database.DB.Get(
		&exist,
		query,
		userID,
	)

	return exist, err
}

func CreatePlayerStats(userID string, battingStyle string, bowlingStyle string) error {
	query := `INSERT INTO player_stats (user_id, batting_style, bowling_style) VALUES ($1, $2, $3)`
	_, err := database.DB.Exec(query, userID, battingStyle, bowlingStyle)
	return err
}
