package entity

import "time"

type Apikey struct {
	ID        int       `json:"id"`
	AppID     string    `json:"appid" db:"appid"`
	Name      string    `json:"name"`
	Token     string    `json:"token"`
	Scope     string    `json:"scope"`
	Deleted   bool      `json:"deleted"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at"`
}
