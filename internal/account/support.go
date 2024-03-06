package account

import (
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/joinself/restful-client/pkg/response"
)

// CreateAccountRequest represents an account creation request.
type CreateAccountRequest struct {
	Username  string   `json:"username"`
	Password  string   `json:"password"`
	Resources []string `json:"resources"`
}

// Validate validates the CreateAccountRequest fields.
func (m CreateAccountRequest) Validate() *response.Error {
	err := validation.ValidateStruct(&m,
		validation.Field(&m.Username, validation.Required, validation.Length(5, 128)),
		validation.Field(&m.Password, validation.Required, validation.Length(5, 128)),
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

// CreateAccountResponse the response returned on account creation.
type CreateAccountResponse struct {
	UserName               string `json:"user_name"`
	Resources              string `json:"resources"`
	RequiresPasswordChange int    `json:"requires_password_change"`
}

// ChangePasswordRequest represents an account update request.
type ChangePasswordRequest struct {
	Password    string `json:"password"`
	NewPassword string `json:"new_password"`
}

// Validate validates the CreateAccountRequest fields.
func (m ChangePasswordRequest) Validate() *response.Error {
	err := validation.ValidateStruct(&m,
		validation.Field(&m.Password, validation.Required, validation.Length(5, 128)),
		validation.Field(&m.NewPassword, validation.Required, validation.Length(5, 128)),
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

type ExtAccount struct {
	UserName               string `json:"id"`
	Resources              string `json:"resources"`
	RequiresPasswordChange bool   `json:"requires_password_change"`
}

type ExtListResponse struct {
	Page       int          `json:"page"`
	PerPage    int          `json:"per_page"`
	PageCount  int          `json:"page_count"`
	TotalCount int          `json:"total_count"`
	Items      []ExtAccount `json:"items"`
}
