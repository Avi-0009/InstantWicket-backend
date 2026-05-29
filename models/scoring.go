package models

type StartInningsRequest struct {
	MatchID       string  `json:"match_id" binding:"required"`
	BattingTeamID string  `json:"batting_team_id" binding:"required"`
	BowlingTeamID string  `json:"bowling_team_id" binding:"required"`
	StrikerID     string  `json:"striker_id" binding:"required"`
	NonStrikerID  *string `json:"non_striker_id"`
	BowlerID      string  `json:"bowler_id" binding:"required"`
	InningsNo     int     `json:"innings_no" binding:"required"`
	TargetRuns    *int    `json:"target_runs"`
}

type RecordBallRequest struct {
	InningsID    string  `json:"innings_id" binding:"required"`
	OverNumber   int     `json:"over_number"`
	BallNumber   int     `json:"ball_number"`
	StrikerID    string  `json:"striker_id" binding:"required"`
	NonStrikerID *string `json:"non_striker_id"`
	BowlerID     string  `json:"bowler_id" binding:"required"`
	IsLegalBall  bool    `json:"is_legal_ball"`
	RunsFromBat  int     `json:"runs_from_bat"`
	Extras       int     `json:"extras"`
	ExtraType    *string `json:"extra_type"`
	IsWicket     bool    `json:"is_wicket"`
	WicketType   *string `json:"wicket_type"`
	OutPlayerID  *string `json:"out_player_id"`
	FielderID    *string `json:"fielder_id"`
}

type LiveScoreboardResponse struct {
	MatchID          string  `json:"match_id" db:"match_id"`
	InningsID        string  `json:"innings_id" db:"innings_id"`
	BattingTeamID    string  `json:"batting_team_id" db:"batting_team_id"`
	BowlingTeamID    string  `json:"bowling_team_id" db:"bowling_team_id"`
	CurrentScore     int     `json:"current_score" db:"current_score"`
	Wickets          int     `json:"wickets" db:"wickets"`
	LegalBalls       int     `json:"legal_balls" db:"legal_balls"`
	RequiredRuns     *int    `json:"target_runs" db:"required_runs"`
	StrikerID        *string `json:"striker_id" db:"striker_id"`
	StrikerName      string  `json:"striker_name" db:"striker_name"`
	StrikerRuns      int     `json:"striker_runs" db:"striker_runs"`
	StrikerBalls     int     `json:"striker_balls" db:"striker_balls"`
	NonStrikerID     *string `json:"non_striker_id" db:"non_striker_id"`
	NonStrikerName   string  `json:"non_striker_name" db:"non_striker_name"`
	NonStrikerRuns   int     `json:"non_striker_runs" db:"non_striker_runs"`
	NonStrikerBalls  int     `json:"non_striker_balls" db:"non_striker_balls"`
	BowlerID         *string `json:"bowler_id" db:"bowler_id"`
	BowlerName       string  `json:"bowler_name" db:"bowler_name"`
	BowlerRuns       int     `json:"bowler_runs" db:"bowler_runs"`
	BowlerWickets    int     `json:"bowler_wickets" db:"bowler_wickets"`
	PartnershipRuns  int     `json:"partnership_runs" db:"partnership_runs"`
	PartnershipBalls int     `json:"partnership_balls" db:"partnership_balls"`
}

type PlayerScorecard struct {
	TeamID       string `json:"team_id" db:"team_id"`
	PlayerID     string `json:"player_id" db:"player_id"`
	PlayerName   string `json:"player_name" db:"player_name"`
	RunsScored   int    `json:"runs_scored" db:"runs_scored"`
	BallsPlayed  int    `json:"balls_played" db:"balls_played"`
	Fours        int    `json:"fours" db:"fours"`
	Sixes        int    `json:"sixes" db:"sixes"`
	IsNotOut     bool   `json:"is_not_out" db:"is_not_out"`
	RunsConceded int    `json:"runs_conceded" db:"runs_conceded"`
	WicketsTaken int    `json:"wickets_taken" db:"wickets_taken"`
	BallsBowled  int    `json:"balls_bowled" db:"balls_bowled"`
	Maidens      int    `json:"maidens" db:"maidens"`
	Wides        int    `json:"wides" db:"wides"`
	NoBalls      int    `json:"no_balls" db:"no_balls"`
	Catches      int    `json:"catches" db:"catches"`
	Runouts      int    `json:"runouts" db:"runouts"`
	Stumpings    int    `json:"stumpings" db:"stumpings"`
}
