package account

import (
	"net/http"
	"unicode"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/joinself/restful-client/pkg/response"
)

// CreateAccountRequest represents an account creation request.
type CreateAccountRequest struct {
	Username               string   `json:"username"`
	Password               string   `json:"password"`
	Resources              []string `json:"resources"`
	RequiresPasswordChange *bool    `json:"requires_password_change,omitempty"`
}

// CreateAccountResponse the response returned on account creation.
type CreateAccountResponse struct {
	UserName               string `json:"user_name"`
	Resources              string `json:"resources"`
	RequiresPasswordChange bool   `json:"requires_password_change"`
}

// Validate validates the CreateAccountRequest fields.
func (m CreateAccountRequest) Validate() *response.Error {
	// Custom password validation rule
	passwordValidation := validation.By(passwordValidator())

	err := validation.ValidateStruct(&m,
		validation.Field(&m.Username, validation.Required, validation.Length(5, 128), is.PrintableASCII),
		validation.Field(&m.Password, validation.Required, passwordValidation, is.PrintableASCII),
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

// ChangePasswordRequest represents an account update request.
type ChangePasswordRequest struct {
	Password    string `json:"password"`
	NewPassword string `json:"new_password"`
}

// Validate validates the CreateAccountRequest fields.
func (m ChangePasswordRequest) Validate() *response.Error {
	passwordValidation := validation.By(passwordValidator())

	err := validation.ValidateStruct(&m,
		validation.Field(&m.Password, validation.Required, validation.Length(5, 128)),
		validation.Field(&m.NewPassword, passwordValidation),
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

func passwordValidator() func(value interface{}) error {
	return func(value interface{}) error {
		s, _ := value.(string)
		var (
			hasMinLen  = false
			hasUpper   = false
			hasLower   = false
			hasNumber  = false
			hasSpecial = false
		)

		if len(s) >= 8 {
			hasMinLen = true
		}

		for _, char := range s {
			switch {
			case unicode.IsUpper(char):
				hasUpper = true
			case unicode.IsLower(char):
				hasLower = true
			case unicode.IsDigit(char):
				hasNumber = true
			case unicode.IsPunct(char) || unicode.IsSymbol(char):
				hasSpecial = true
			}
		}

		if !(hasMinLen && hasUpper && hasLower && hasNumber && hasSpecial) {
			return validation.NewError("validation_password_strength", "password must be at least 8 characters long, include upper and lower case letters, numbers, and special characters")
		}

		return nil
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
