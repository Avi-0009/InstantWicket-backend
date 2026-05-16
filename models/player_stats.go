package models

type CreatePlayerStats struct {
	BattingStyle string `json:"batting_style"`
	BowlingStyle string `json:"bowling_style"`
}

type UpdatePlayerStats struct {
	BattingStyle string `json:"batting_style"`
	BowlingStyle string `json:"bowling_style"`
}

type PlayerStats struct {
	ID              string  `db:"id" json:"id"`
	UserID          string  `db:"user_id" json:"user_id"`
	BattingStyle    string  `db:"batting_style" json:"batting_style"`
	BowlingStyle    string  `db:"bowling_style" json:"bowling_style"`
	CareerMatches   int     `db:"career_matches" json:"career_matches"`
	CareerRuns      int     `db:"career_runs" json:"career_runs"`
	CareerWickets   int     `db:"career_wickets" json:"career_wickets"`
	CareerCatches   int     `db:"career_catches" json:"career_catches"`
	CareerRunouts   int     `db:"career_runouts" json:"career_runouts"`
	CareerStumpings int     `db:"career_stumpings" json:"career_stumpings"`
	CareerFours     int     `db:"career_fours" json:"career_fours"`
	CareerSixes     int     `db:"career_sixes" json:"career_sixes"`
	StrikeRate      float64 `db:"strike_rate" json:"strike_rate"`
	Economy         float64 `db:"economy" json:"economy"`
}
