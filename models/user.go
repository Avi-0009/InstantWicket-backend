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
