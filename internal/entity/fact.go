package entity

import (
	"time"
)

// Fact represents a fact record.
type Fact struct {
	ID           string    `json:"id"`
	ConnectionID string    `json:"-"`
	ISS          string    `json:"iss"`
	Source       string    `json:"source"`
	Body         string    `json:"body"`
	IAT          time.Time `json:"created_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}