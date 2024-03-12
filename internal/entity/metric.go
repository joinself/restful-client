package entity

import "time"

type Metric struct {
	ID        int       `json:"id"`
	AppID     string    `json:"appid" db:"appid"`
	UUID      int       `json:"uuid"`
	Recipient string    `json:"recipient"`
	Actions   string    `json:"actions"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
