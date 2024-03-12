package metric

import "time"

type ExtMetric struct {
	ID        int       `json:"id"`
	Recipient string    `json:"recipient"`
	Actions   string    `json:"actions"`
	CreatedAt time.Time `json:"created_at"`
}
