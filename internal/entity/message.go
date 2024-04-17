package entity

import (
	"time"
)

// Message represents a message record.
type Message struct {
	ID           int       `json:"-"`
	ConnectionID int       `json:"-"`
	ISS          string    `json:"iss"`
	CID          string    `json:"cid"`
	JTI          string    `json:"jti"`
	RID          string    `json:"rid"`
	Body         string    `json:"body"`
	IAT          time.Time `json:"iat"`
	Read         bool      `json:"read"`
	Received     bool      `json:"received"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
