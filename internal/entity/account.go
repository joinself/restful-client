package entity

import "time"

type Account struct {
	ID             int       `json:"id"`
	UserName       string    `json:"user_name"`
	HashedPassword string    `json:"hashed_password"`
	Password       string    `json:"password" db:"-"`
	Salt           string    `json:"salt"`
	Resources      string    `json:"resources"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
