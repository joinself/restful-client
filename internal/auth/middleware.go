package auth

import (
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/joinself/restful-client/internal/entity"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

type jwtCustomClaims struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Admin     bool     `json:"admin"`
	Resources []string `json:"resources"`
	jwt.RegisteredClaims
}

// Handler returns a JWT-based authentication middleware.
func Handler(verificationKey string) echo.MiddlewareFunc {
	config := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(jwtCustomClaims)
		},
		SigningKey: []byte(verificationKey),
	}

	return echojwt.WithConfig(config)
}

type contextKey int

const (
	userKey contextKey = iota
)

// WithUser returns a context that contains the user identity from the given JWT.
func WithUser(ctx echo.Context, token *jwt.Token) {
	ctx.Set("user", token)
}

// CurrentUser returns the user identity from the given context.
// Nil is returned if no user identity is found in the context.
func CurrentUser(c echo.Context) Identity {
	token, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return nil
	}
	claims, ok := token.Claims.(*jwtCustomClaims) // by default claims is of type `jwt.MapClaims`
	if !ok {
		return nil
	}
	return entity.User{
		ID:        claims.ID,
		Name:      claims.Name,
		Admin:     claims.Admin,
		Resources: claims.Resources,
	}
}

// HasAccessToResource checks if the current user has access to a specific resource.
func HasAccessToResource(c echo.Context, resource string) bool {
	u := CurrentUser(c)
	if u == nil {
		return false
	}

	if u.IsAdmin() {
		return true
	}

	for _, v := range u.GetResources() {
		if v == resource {
			return true
		}
	}

	return false
}

// MockAuthHandler creates a mock authentication middleware for testing purpose.
// If the request contains an Authorization header whose value is "TEST", then
// it considers the user is authenticated as "Tester" whose ID is "100".
// It fails the authentication otherwise.
func MockAuthHandler() echo.MiddlewareFunc {
	config := echojwt.Config{
		Skipper: func(c echo.Context) bool {
			return (c.Request().Header.Get("Authorization") == "TEST")
		},
		SigningKey: []byte("test"),
	}

	return echojwt.WithConfig(config)
}

// MockAuthHeader returns an HTTP header that can pass the authentication check by MockAuthHandler.
func MockAuthHeader() http.Header {
	header := http.Header{}
	header.Add("Authorization", "TEST")
	return header
}
