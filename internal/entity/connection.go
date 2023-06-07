package entity

import (
	"time"
)

// Connection represents an connection record.
type Connection struct {
	ID        int       `json:"id"`
	SelfID    string    `json:"selfid" db:"selfid"`
	AppID     string    `json:"appid" db:"appid"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
