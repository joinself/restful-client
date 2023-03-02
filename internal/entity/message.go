package entity

import (
	"time"
)

// Message represents a message record.
type Message struct {
	ID           string    `json:"id"`
	ConnectionID string    `json:"-"`
	ISS          string    `json:"iss"`
	CID          string    `json:"cid"`
	RID          string    `json:"rid"`
	Body         string    `json:"body"`
	IAT          time.Time `json:"iat"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}