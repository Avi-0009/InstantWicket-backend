package models

type CreatePlayerStats struct {
	BattingStyle string `json:"batting_style"`
	BowlingStyle string `json:"bowling_style"`
}

type UpdatePlayerStats struct {
	BattingStyle string `json:"batting_style"`
	BowlingStyle string `json:"bowling_style"`
}

type PlayerSearchResponse struct {
	PlayerID      string `db:"player_id" json:"player_id"`
	UserID        string `db:"user_id" json:"user_id"`
	Name          string `db:"name" json:"name"`
	PhoneNo       string `db:"phone_no" json:"phone_no"`
	CareerRuns    int64  `db:"career_runs" json:"career_runs"`
	CareerWickets int64  `db:"career_wickets" json:"career_wickets"`
}
type PlayerStats struct {
	ID                       string  `db:"id" json:"id"`
	UserID                   string  `db:"user_id" json:"user_id"`
	Name                     string  `db:"name" json:"name"`
	BattingStyle             string  `db:"batting_style" json:"batting_style"`
	BowlingStyle             string  `db:"bowling_style" json:"bowling_style"`
	CareerMatches            int64   `db:"career_matches" json:"career_matches"`
	CareerMatchesWon         int64   `db:"career_matches_won" json:"career_matches_won"`
	CareerMatchesLost        int64   `db:"career_matches_lost" json:"career_matches_lost"`
	CareerRuns               int64   `db:"career_runs" json:"career_runs"`
	CareerBallsFaced         int64   `db:"career_balls_faced" json:"career_balls_faced"`
	CareerInningsBatted      int64   `db:"career_innings_batted" json:"career_innings_batted"`
	CareerNotOuts            int64   `db:"career_not_outs" json:"career_not_outs"`
	CareerHighestScore       int     `db:"career_highest_score" json:"career_highest_score"`
	CareerDucks              int64   `db:"career_ducks" json:"career_ducks"`
	CareerGoldenDucks        int64   `db:"career_golden_ducks" json:"career_golden_ducks"`
	CareerFifties            int64   `db:"career_fifties" json:"career_fifties"`
	CareerHundreds           int64   `db:"career_hundreds" json:"career_hundreds"`
	CareerFours              int64   `db:"career_fours" json:"career_fours"`
	CareerSixes              int64   `db:"career_sixes" json:"career_sixes"`
	StrikeRate               float64 `db:"strike_rate" json:"strike_rate"`
	CareerWickets            int64   `db:"career_wickets" json:"career_wickets"`
	CareerBallsBowled        int64   `db:"career_balls_bowled" json:"career_balls_bowled"`
	CareerRunsConceded       int64   `db:"career_runs_conceded" json:"career_runs_conceded"`
	CareerMaidenOvers        int64   `db:"career_maiden_overs" json:"career_maiden_overs"`
	CareerWides              int64   `db:"career_wides" json:"career_wides"`
	CareerNoBalls            int64   `db:"career_no_balls" json:"career_no_balls"`
	CareerBestBowlingWickets int     `db:"career_best_bowling_wickets" json:"career_best_bowling_wickets"`
	CareerBestBowlingRuns    int     `db:"career_best_bowling_runs" json:"career_best_bowling_runs"`
	CareerInningsBowled      int64   `db:"career_innings_bowled" json:"career_innings_bowled"`
	Economy                  float64 `db:"economy" json:"economy"`
	CareerCatches            int64   `db:"career_catches" json:"career_catches"`
	CareerRunouts            int64   `db:"career_runouts" json:"career_runouts"`
	CareerStumpings          int64   `db:"career_stumpings" json:"career_stumpings"`
	CareerTotalPoints        int64   `db:"career_total_points" json:"career_total_points"`
	CareerMVPS               int64   `db:"career_mvps" json:"career_mvps"`
	CreatedAt                string  `db:"created_at" json:"created_at"`
	UpdatedAt                string  `db:"updated_at" json:"updated_at"`
	ArchivedAt               *string `db:"archived_at" json:"archived_at"`
}
