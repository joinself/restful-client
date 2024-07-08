package request

import (
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
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

func (f FactRequest) Validate() error {
	return validation.ValidateStruct(&f,
		validation.Field(&f.Name, validation.Required, validation.Length(3, 128)),
	)
}

// CreateRequest represents an request creation request.
type CreateRequest struct {
	Type        string        `json:"type"`
	Facts       []FactRequest `json:"facts"`
	Description string        `json:"description"`
	Callback    string        `json:"callback"`
	SelfID      string        `json:"connection_self_id"`
	OutOfBand   bool          `json:"out_of_band,omitempty"`
	AllowedFor  int64         `json:"allowed_for,omitempty"`
}

// Validate validates the CreateRequest fields.
func (m CreateRequest) Validate() *response.Error {
	err := validation.ValidateStruct(&m,
		validation.Field(&m.Type, validation.Required, validation.Length(0, 128)),
		validation.Field(&m.Type, validation.In("auth", "fact")),
		validation.Field(&m.Description, validation.Length(0, 128)),
		validation.Field(&m.Callback, is.URL),
	)
	if err != nil {
		return &response.Error{
			Status:  http.StatusBadRequest,
			Error:   "Invalid input",
			Details: err.Error(),
		}
	}

	for _, f := range m.Facts {
		if err = f.Validate(); err != nil {
			return &response.Error{
				Status:  http.StatusBadRequest,
				Error:   "Invalid input",
				Details: err.Error(),
			}
		}
	}

	return nil
}
