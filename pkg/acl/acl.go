package acl

import (
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/labstack/echo/v4"
)

// Identity represents an authenticated user identity.
type Identity interface {
	// GetID returns the user ID.
	GetID() string
	// GetName returns the user name.
	GetName() string
	IsAdmin() bool
	GetResources() []string
	IsPasswordChangeRequired() bool
}

type JWTCustomClaims struct {
	ID                       string   `json:"id"`
	Name                     string   `json:"name"`
	Admin                    bool     `json:"admin"`
	Resources                []string `json:"resources"`
	IsPasswordChangeRequired bool     `json:"change_password"`
	jwt.RegisteredClaims
}

// CurrentUser returns the user identity from the given context.
// Nil is returned if no user identity is found in the context.
func CurrentUser(c echo.Context) Identity {
	token, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return nil
	}
	claims, ok := token.Claims.(*JWTCustomClaims) // by default claims is of type `jwt.MapClaims`
	if !ok {
		return nil
	}
	return entity.User{
		ID:                     claims.ID,
		Name:                   claims.Name,
		Admin:                  claims.Admin,
		Resources:              claims.Resources,
		RequiresPasswordChange: claims.IsPasswordChangeRequired,
	}
}

// HasAccessToResource checks if the current user has access to a specific resource.
func HasAccessToResource(c echo.Context, resource string) bool {
	u := CurrentUser(c)
	if u == nil {
		c.JSON(http.StatusNotFound, map[string]string{
			"message": "resource not found (A",
		})
		return false
	}

	if u.IsAdmin() {
		return true
	}

	if u.IsPasswordChangeRequired() {
		c.JSON(http.StatusLocked, "you're required to change your password")
		return false
	}

	for _, v := range u.GetResources() {
		if v == resource {
			return true
		}
	}

	c.JSON(http.StatusNotFound, "resource not found")
	return false
}

func IsAdmin(c echo.Context) bool {
	u := CurrentUser(c)
	if u == nil {
		return false
	}
	return u.IsAdmin()
}
