package object

import (
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/joinself/restful-client/pkg/response"
)

type Object struct {
	// https://en.wikipedia.org/wiki/Data_URI_scheme on a format like
	// data:[<MIME-type>][;charset=<encoding>][;base64],<data>
	DataURI string `json:"data_uri"`
}

// Validate validates the CreateSignatureRequest fields.
func (m Object) Validate() *response.Error {
	err := validation.ValidateStruct(&m,
		validation.Field(&m.DataURI, validation.Required),
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

type CreateObjectRequest struct {
	Objects []Object
}

func (c CreateObjectRequest) Validate() *response.Error {
	if len(c.Objects) < 1 {
		return &response.Error{
			Status:  http.StatusBadRequest,
			Error:   "Invalid input",
			Details: "you must provide at least an attachment",
		}
	}

	for _, o := range c.Objects {
		err := o.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

type CreateObjectResponse struct {
	Objects []ExtObject `json:"objects"`
}

type ExtObject struct {
	Link    string `json:"link"`
	Name    string `json:"name"`
	Mime    string `json:"mime"`
	Expires int64  `json:"expires"`
	Key     string `json:"key"`
}
