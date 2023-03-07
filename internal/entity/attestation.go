package entity

import (
	"time"
)

// Attestation represents an attestation record.
type Attestation struct {
	ID        string    `json:"id"`
	FactID    string    `json:"-"`
	Body      string    `json:"body"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
