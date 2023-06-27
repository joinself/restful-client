package entity

import (
	"time"
)

type Resource struct {
	URI string `json:"uri"`
}

type RequestFacts struct {
	Sources []string `json:"sources"`
	Name    string   `json:"name"`
}

// Request represents a request record.
type Request struct {
	ID           string    `json:"id"`
	Type         string    `json:"typ"`
	ConnectionID int       `json:"-"`
	Facts        []byte    `json:"facts"`
	Auth         bool      `json:"auth,omitempty"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (r *Request) IsResponded() bool {
	return (r.Status == "responded")
}