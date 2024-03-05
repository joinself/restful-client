package auth

import (
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/joinself/restful-client/pkg/response"
)

// LoginRequest authentication login input request.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Validate validates the LoginRequest struct
func (r *LoginRequest) Validate() *response.Error {
	errBasic := validation.ValidateStruct(r,
		validation.Field(&r.Username, validation.Required, validation.Length(5, 128)),
		validation.Field(&r.Password, validation.Required, validation.Length(5, 128)),
	)

	if errBasic != nil {
		return &response.Error{
			Status:  http.StatusBadRequest,
			Error:   "Invalid input",
			Details: "You must provide user and password",
		}
	}

	return nil
}

// LoginResponse authentication login output response.
type LoginResponse struct {
	AccessToken  string `json:"token"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// RefreshRequest refresh your JWT token.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Validate validates the LoginRequest struct
func (r *RefreshRequest) Validate() *response.Error {
	errRefresh := validation.ValidateStruct(r,
		validation.Field(&r.RefreshToken, validation.Required, validation.Length(32, 255)),
	)

	if errRefresh != nil {
		return &response.Error{
			Status:  http.StatusBadRequest,
			Error:   "Invalid input",
			Details: "You must provide a refresh_token",
		}
	}

	return nil
}
