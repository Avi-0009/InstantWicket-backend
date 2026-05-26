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
	TeamAName         string             `json:"team_a_name"`
	TeamBName         string             `json:"team_b_name"`
	TeamAPlayers      []MatchPlayerInput `json:"team_a_players"`
	TeamBPlayers      []MatchPlayerInput `json:"team_b_players"`
	TossWinner        string             `json:"toss_winner_team_id"`
	TossDecision      string             `json:"toss_decision"`
	AllowCommonPlayer bool               `json:"allow_common_player"`
	AllowSoloBatting  bool               `json:"allow_solo_batting"`
	OversLimit        int64              `json:"overs_limit"`
	UmpireID          string             `json:"umpire_id"`
}

type Match struct {
	ID                string        `db:"id" json:"id"`
	TeamAID           string        `db:"team_a_id" json:"team_a_id"`
	TeamAName         string        `db:"team_a_name" json:"team_a_name"`
	TeamAPlayers      []MatchPlayer `json:"team_a_players"`
	TeamBID           string        `db:"team_b_id" json:"team_b_id"`
	TeamBName         string        `db:"team_b_name" json:"team_b_name"`
	TeamBPlayers      []MatchPlayer `json:"team_b_players"`
	TossWinner        string        `db:"toss_winner_team_id" json:"toss_winner_team_id"`
	TossDecision      string        `db:"toss_decision" json:"toss_decision"`
	AllowCommonPlayer bool          `db:"allow_common_player" json:"allow_common_player"`
	AllowSoloBatting  bool          `db:"allow_solo_batting" json:"allow_solo_batting"`
	OversLimit        int64         `db:"overs_limit" json:"overs_limit"`
	Status            string        `db:"status" json:"status"`
	WinnerTeamID      *string       `db:"winner_team_id" json:"winner_team_id"`
	ManOfMatch        *string       `db:"man_of_match" json:"man_of_match"`
	WorstPlayer       *string       `db:"worst_player" json:"worst_player"`
	UmpireID          *string       `db:"umpire_id" json:"umpire_id"`
	CreatedBy         string        `db:"created_by" json:"created_by"`
	CreatedAt         string        `db:"created_at" json:"created_at"`
	UpdatedAt         string        `db:"updated_at" json:"updated_at"`
}

// for incoming match creation

type MatchPlayerInput struct {
	ID             string `db:"id" json:"id"` // This is the player_stats ID
	Name           string `db:"name" json:"name"`
	PhoneNo        string `db:"phone_no" json:"phone_no"`
	IsCommonPlayer bool   `db:"is_common_player" json:"is_common_player"`
	IsCaptain      bool   `db:"is_captain" json:"is_captain"`
	IsWicketKeeper bool   `db:"is_wicket_keeper" json:"is_wicket_keeper"`
}

// for outgoing player fetching

type MatchPlayer struct {
	TeamID         string `db:"team_id" json:"team_id"`
	ID             string `db:"id" json:"id"`
	Name           string `db:"name" json:"name"`
	IsCommonPlayer bool   `db:"is_common_player" json:"is_common_player"`
	IsCaptain      bool   `db:"is_captain" json:"is_captain"`
	IsWicketKeeper bool   `db:"is_wicket_keeper" json:"is_wicket_keeper"`
	IsRetired      bool   `db:"is_retired" json:"is_retired"`
	ReturnedToPlay bool   `db:"returned_to_play" json:"returned_to_play"`
}
