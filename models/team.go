package models

type CreateTeam struct {
	Name string `json:"name"`
}

type UpdateTeam struct {
	Name string `json:"name"`
}

type Team struct {
	ID        string `db:"id" json:"id"`
	Name      string `db:"name" json:"name"`
	CreatedBy string `db:"created_by" json:"created_by"`
	CreatedAt string `db:"created_at" json:"created_at"`
	UpdatedAt string `db:"updated_at" json:"updated_at"`
}

type TeamPlayer struct {
	ID             string `json:"id" db:"id"`
	Name           string `json:"name" db:"name"`
	BattingStyle   string `json:"batting_style" db:"batting_style"`
	BowlingStyle   string `json:"bowling_style" db:"bowling_style"`
	IsCommonPlayer bool   `json:"is_common_player" db:"is_common_player"`
}
