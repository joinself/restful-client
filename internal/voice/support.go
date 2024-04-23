package voice

import (
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/joinself/restful-client/pkg/response"
)

type ProceedData struct {
	PeerInfo string `json:"peer_info"`
	Name     string `json:"name"`
}

func (p ProceedData) Validate() *response.Error {
	err := validation.ValidateStruct(&p,
		validation.Field(&p.PeerInfo, validation.Required, validation.Length(0, 128)),
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

type SetupData struct {
	Name string `json:"name"`
}

func (p SetupData) Validate() *response.Error {
	err := validation.ValidateStruct(&p,
		validation.Field(&p.Name, validation.Required, validation.Length(0, 128)),
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
