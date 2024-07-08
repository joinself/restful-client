package signature

import (
	"encoding/json"
	"net/http"
	"regexp"
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

func (o Object) Validate() error {
	dataURLPattern := `^data:([a-zA-Z0-9!#$&^_-]+/[a-zA-Z0-9!#$&^_-]+)?(;charset=[a-zA-Z0-9!#$&^_-]+)?(;base64)?,.*$`
	return validation.ValidateStruct(&o,
		validation.Field(&o.Title, validation.Required, validation.Length(3, 128)),
		validation.Field(&o.DataURI, validation.Required, validation.Match(regexp.MustCompile(dataURLPattern))),
	)
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
	if err != nil {
		return &response.Error{
			Status:  http.StatusBadRequest,
			Error:   "Invalid input",
			Details: err.Error(),
		}
	}
	if len(m.Objects) > 0 {
		for _, o := range m.Objects {
			if err := o.Validate(); err != nil {
				return &response.Error{
					Status:  http.StatusBadRequest,
					Error:   "Invalid input",
					Details: err.Error(),
				}
			}
		}
	}

	return nil
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
