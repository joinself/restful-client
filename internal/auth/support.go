package auth

import (
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/joinself/restful-client/pkg/response"
)

// LoginRequest authentication login input request.
type LoginRequest struct {
	// Username the username to authenticate with. Length has to be between 5 and
	// 128 characters.
	Username string `json:"username"`
	// Password password to authenticate with. Length has to be between 5 and
	// 128 characters.
	Password string `json:"password"`
}

// Validate validates the LoginRequest struct.
func (r *LoginRequest) Validate() *response.Error {
	err := validation.ValidateStruct(r,
		validation.Field(&r.Username, validation.Required, validation.Length(5, 128)),
		validation.Field(&r.Password, validation.Required, validation.Length(5, 128)),
	)

	if err != nil {
		return &response.Error{
			Status:  http.StatusBadRequest,
			Error:   "Invalid input",
			Details: err.Error(),
		}
	}

	return nil
}

// LoginResponse authentication login output response.
type LoginResponse struct {
	// AccessToken the token to be used on authenticated requests.
	AccessToken string `json:"token"`
	// RefreshToken the token to be used to refresh the access token.
	RefreshToken string `json:"refresh_token,omitempty"`
}

// RefreshRequest refresh your JWT token.
type RefreshRequest struct {
	// RefreshToken the token to be used to refresh the access token.
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
