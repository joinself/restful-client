package entity

import (
	"fmt"
	"time"
)

const (
	STATUS_ACCEPTED string = "accepted"
	STATUS_REJECTED string = "rejected"
	STATUS_ERRORED  string = "errored"
)

// Fact represents a fact record.
type Fact struct {
	ID           string    `json:"id"`
	ConnectionID int       `json:"-"`
	RequestID    string    `json:"request_id"`
	ISS          string    `json:"iss"`
	CID          string    `json:"cid,omitempty"`
	JTI          string    `json:"jti,omitempty"`
	Status       string    `json:"status"`
	Source       string    `json:"source"`
	Fact         string    `json:"fact"`
	Body         string    `json:"body"`
	IAT          time.Time `json:"iat"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (f *Fact) URI(app, connection string) string {
	return fmt.Sprintf("/v1/apps/%s/connections/%s/facts/%s", app, connection, f.ID)
}
