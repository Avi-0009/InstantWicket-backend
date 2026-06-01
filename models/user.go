package models

type RegisterUser struct {
	Name     string `json:"name" binding:"required"`
	PhoneNo  string `json:"phone_no" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginUser struct {
	PhoneNo  string `json:"phone_no" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type User struct {
	ID       string `db:"id" json:"id"`
	Name     string `db:"name" json:"name"`
	PhoneNo  string `db:"phone_no" json:"phone_no"`
	Password string `db:"password" json:"-"`
}

type UpdateProfile struct {
	Name         string `db:"name" json:"name"`
	PhoneNo      string `db:"phone_no" json:"phone_no"`
	BattingStyle string `db:"batting_style" json:"batting_style"`
	BowlingStyle string `db:"bowling_style" json:"bowling_style"`
}

type ResetPassword struct {
	PhoneNo  string `json:"phone_no" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

type Session struct {
	ID     string `db:"id" json:"id"`
	UserID string `db:"user_id" json:"user_id"`
}
