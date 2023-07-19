package entity

import (
	"time"
)

// Message represents a message record.
type Message struct {
	ID           int       `json:"id"`
	ConnectionID int       `json:"-"`
	ISS          string    `json:"iss"`
	CID          string    `json:"cid"`
	JTI          string    `json:"jti"`
	RID          string    `json:"rid"`
	Body         string    `json:"body"`
	IAT          time.Time `json:"iat"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
