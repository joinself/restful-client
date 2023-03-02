package entity

import (
	"time"
)

// Connection represents an connection record.
type Connection struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Selfid    string    `json:"selfid"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
