package models

type CreateMatch struct {
	TeamAID           string `json:"team_a_id"`
	TeamBID           string `json:"team_b_id"`
	TossWinnerTeamID  string `json:"toss_winner_team_id"`
	TossDecision      string `json:"toss_decision"`
	AllowCommonPlayer bool   `json:"allow_common_player"`
	AllowSoloBatting  bool   `json:"allow_solo_batting"`
	OversLimit        int64  `json:"overs_limit"`
	UmpireID          string `json:"umpire_id"`
}

type StartLiveMatchRequest struct {
	TeamAName         string `json:"team_a_name"`
	TeamBName         string `json:"team_b_name"`
	TossWinner        string `json:"toss_winner_team_id"`
	TossDecision      string `json:"toss_decision"`
	AllowCommonPlayer bool   `json:"allow_common_player"`
	AllowSoloBatting  bool   `json:"allow_solo_batting"`
	OversLimit        int64  `json:"overs_limit"`
	UmpireID          string `json:"umpire_id"`
}

type Match struct {
	ID                string `db:"id" json:"id"`
	TeamAID           string `db:"team_a_id" json:"team_a_id"`
	TeamAName         string `db:"team_a_name" json:"team_a_name"`
	TeamBID           string `db:"team_b_id" json:"team_b_id"`
	TeamBName         string `db:"team_b_name" json:"team_b_name"`
	TossWinner        string `db:"toss_winner_team_id" json:"toss_winner_team_id"`
	TossDecision      string `db:"toss_decision" json:"toss_decision"`
	AllowCommonPlayer bool   `db:"allow_commom_player" json:"allow_common_player"`
	AllowSoloBatting  bool   `db:"allow_solo_batting" json:"allow_solo_batting"`
	OversLimit        int64  `db:"over_limit" json:"overs_limit"`
	Status            string `db:"status" json:"status"`
	WinnerTeamID      string `db:"winner_team_id" json:"winner_team_id"`
	ManOfMatch        string `db:"man_of_match" json:"man_of_match"`
	WorstPlayer       string `db:"worst_player" json:"worst_player"`
	UmpireID          string `db:"umpire_id" json:"umpire_id"`
	CreatedBy         string `db:"created_by" json:"created_by"`
	CreatedAt         string `db:"created_at" json:"created_at"`
	UpdatedAt         string `db:"updated_at" json:"updated_at"`
}
