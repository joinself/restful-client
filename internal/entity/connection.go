package entity

import (
	"time"
)

// Connection represents an connection record.
type Connection struct {
	ID        string    `json:"id"`
	SelfID    string    `json:"selfid"`
	AppID     string    `json:"appid"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
