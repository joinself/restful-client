package signature

import (
	"encoding/json"
	"net/http"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/joinself/restful-client/pkg/response"
)

type Object struct {
	// https://en.wikipedia.org/wiki/Data_URI_scheme on a format like
	// data:[<MIME-type>][;charset=<encoding>][;base64],<data>
	DataURI string `json:"data_uri"`
	Title   string `json:"title"`
}

type CreateSignatureRequest struct {
	Description string
	Objects     []Object
}

// Validate validates the CreateSignatureRequest fields.
func (m CreateSignatureRequest) Validate() *response.Error {
	err := validation.ValidateStruct(&m,
		validation.Field(&m.Description, validation.Required, validation.Length(0, 128)),
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

type ExtSignature struct {
	ID          string          `json:"id"`
	Description string          `json:"description"`
	Status      string          `json:"status"`
	Data        json.RawMessage `json:"data"`
	Signature   string          `json:"signature"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}
