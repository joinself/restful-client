package entity

import (
	"time"
)

// Call
type Call struct {
	ID        int       `json:"id"`
	SelfID    string    `json:"selfid" db:"selfid"`
	AppID     string    `json:"appid" db:"appid"`
	PeerInfo  string    `json:"peer_info"`
	CallID    string    `json:"call_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
