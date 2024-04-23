package entity

import (
	"time"
)

const (
	VOICE_CALL_SETUP     = "setup"
	VOICE_CALL_STARTED   = "started"
	VOICE_CALL_ACCEPTED  = "accepted"
	VOICE_CALL_BUSY      = "busy"
	VOICE_CALL_ENDED     = "ended"
	VOICE_CALL_CANCELLED = "cancelled"
)

// Call
type Call struct {
	ID        int       `json:"id"`
	SelfID    string    `json:"selfid" db:"selfid"`
	AppID     string    `json:"appid" db:"appid"`
	PeerInfo  string    `json:"peer_info"`
	CallID    string    `json:"call_id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
