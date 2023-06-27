package entity

import (
	"fmt"
	"time"
)

// Fact represents a fact record.
type Fact struct {
	ID           string    `json:"id"`
	ConnectionID int       `json:"-"`
	RequestID    string    `json:"request_id"`
	ISS          string    `json:"iss"`
	CID          string    `json:"cid"`
	JTI          string    `json:"jti"`
	Status       string    `json:"status"`
	Source       string    `json:"source"`
	Fact         string    `json:"fact"`
	Body         string    `json:"body"`
	IAT          time.Time `json:"iat"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (f *Fact) URI(app string) string {
	return fmt.Sprintf("/v1/apps/%s/connections/%d/facts/%s", app, f.ConnectionID, f.ID)
}
