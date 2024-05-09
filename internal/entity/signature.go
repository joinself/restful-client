package entity

import (
	"time"
)

const (
	SIGNATURE_REQUESTED_STATUS = "requested"
	SIGNATURE_ACCEPTED_STATUS  = "accepted"
	SIGNATURE_REJECTED_STATUS  = "rejected"
	SIGNATURE_ERRORED_STATUS   = "errored"
)

// Signature
type Signature struct {
	ID          string    `json:"id"`
	AppID       string    `json:"app_id"`
	SelfID      string    `json:"selfid"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Data        []byte    `json:"data"`
	Signature   string    `json:"signature"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
