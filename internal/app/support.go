package app

import (
	"errors"
	"net/http"
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
	"github.com/joinself/restful-client/pkg/response"
)

type ExtApp struct {
	ID       string `json:"id"`
	Name     string `json:"name,omitempty"`
	Env      string `json:"env,omitempty"`
	Status   string `json:"status,omitempty"`
	Callback string `json:"callback,,omitempty"`
}

type ExtListResponse struct {
	Page       int      `json:"page"`
	PerPage    int      `json:"per_page"`
	PageCount  int      `json:"page_count"`
	TotalCount int      `json:"total_count"`
	Items      []ExtApp `json:"items"`
}

type CreateAppRequest struct {
	ID             string `json:"id"`
	Secret         string `json:"secret"`
	Name           string `json:"name"`
	Env            string `json:"env"`
	Callback       string `json:"callback"`
	CallbackSecret string `json:"callback_secret"`
}

// Validate validates the CreateAppRequest fields.
func (m CreateAppRequest) Validate() *response.Error {
	err := validation.ValidateStruct(&m,
		// TODO: Improve validations
		validation.Field(&m.ID, validation.Required, validation.By(m.uuidValidator())),
		validation.Field(&m.Secret, validation.Required, validation.By(m.deviceKeyValidator())),
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

func (m CreateAppRequest) uuidValidator() func(value interface{}) error {
	return func(input interface{}) error {
		_, err := uuid.Parse(input.(string))
		if err == nil {
			return nil
		}

		return errors.New("not valid UUID")
	}
}

func (m CreateAppRequest) deviceKeyValidator() func(value interface{}) error {
	return func(input interface{}) error {
		pattern := `^sk_[A-Za-z0-9]+:[A-Za-z0-9/]+$`
		matched, err := regexp.MatchString(pattern, input.(string))
		if err != nil {
			return errors.New("not valid secret")
		}
		if !matched {
			return errors.New("not valid secret")
		}
		return nil
	}
}

type UpdateAppRequest struct {
	Callback       string `json:"callback"`
	CallbackSecret string `json:"callback_secret"`
}

// Validate validates the CreateAppRequest fields.
func (m UpdateAppRequest) Validate() *response.Error {
	return nil
}
