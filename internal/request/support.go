package request

import (
	"net/http"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/joinself/restful-client/pkg/response"
)

type ExtResource struct {
	ID           string `json:"id"`
	ConnectionID string `json:"connection_id"`
}

type ExtRequest struct {
	ID        string        `json:"id"`
	AppID     string        `json:"app_id"`
	Status    string        `json:"status,omitempty"`
	QRCode    string        `json:"qr_code,omitempty"`
	DeepLink  string        `json:"deep_link,omitempty"`
	Resources []ExtResource `json:"resources,omitempty"`
}

type FactRequest struct {
	Sources []string `json:"sources,omitempty"`
	Name    string   `json:"name"`
}

// CreateRequest represents an request creation request.
type CreateRequest struct {
	Type        string        `json:"type"`
	Facts       []FactRequest `json:"facts"`
	Description string        `json:"description"`
	Callback    string        `json:"callback"`
	SelfID      string        `json:"connection_self_id"`
	OutOfBand   bool          `json:"out_of_band,omitempty"`
	AllowedFor  time.Duration `json:"allowed_for,omitempty"`
}

// Validate validates the CreateRequest fields.
func (m CreateRequest) Validate() *response.Error {
	err := validation.ValidateStruct(&m,
		validation.Field(&m.Type, validation.Required, validation.Length(0, 128)),
		validation.Field(&m.Type, validation.In("auth", "fact")),
	)
	if err == nil {
		return nil
	}

	return &response.Error{
		Status:  http.StatusBadRequest,
		Error:   "Invalid input",
		Details: err.Error(),
	}
}
