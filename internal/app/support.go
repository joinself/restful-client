package app

import (
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/joinself/restful-client/pkg/response"
)

type ExtApp struct {
	ID     string `json:"id"`
	Name   string `json:"name,omitempty"`
	Env    string `json:"env,omitempty"`
	Status string `json:"status,omitempty"`
}

type ExtListResponse struct {
	Page       int      `json:"page"`
	PerPage    int      `json:"per_page"`
	PageCount  int      `json:"page_count"`
	TotalCount int      `json:"total_count"`
	Items      []ExtApp `json:"items"`
}

type CreateAppRequest struct {
	ID       string `json:"id"`
	Secret   string `json:"secret"`
	Name     string `json:"name"`
	Env      string `json:"env"`
	Callback string `json:"callback"`
}

// Validate validates the CreateAppRequest fields.
func (m CreateAppRequest) Validate() *response.Error {
	err := validation.ValidateStruct(&m,
		// TODO: Improve validations
		validation.Field(&m.ID, validation.Required, validation.Length(5, 128)),
		validation.Field(&m.Secret, validation.Required, validation.Length(5, 128)),
		validation.Field(&m.Name, validation.Required, validation.Length(3, 50)),
		validation.Field(&m.Env, validation.Required, validation.Length(0, 20)),
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
