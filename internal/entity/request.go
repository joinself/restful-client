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
	ID           string        `json:"id"`
	AppID        string        `json:"app_id"`
	Type         string        `json:"typ"`
	ConnectionID *int          `json:"-"`
	Facts        []byte        `json:"facts"`
	Description  string        `json:"description"`
	Auth         bool          `json:"auth,omitempty"`
	OutOfBand    bool          `json:"out_of_band,omitempty"`
	Status       string        `json:"status"`
	Callback     string        `json:"callback"`
	AllowedFor   time.Duration `json:"allowed_for" db:"-"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

func (r *Request) IsResponded() bool {
	return (r.Status == "responded")
}

func (r *Request) IsOutOfBand() bool {
	return r.OutOfBand
}
