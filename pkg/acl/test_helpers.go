package acl

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/joinself/restful-client/internal/entity"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

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

func AuthAsAdminMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("user", BuildJwt(entity.User{
				ID:        "100",
				Name:      "admin",
				Admin:     true,
				Resources: []string{},
			}))

			return next(c)
		}
	}
}

func AuthAsPlainMiddleware(resources []string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("user", BuildJwt(entity.User{
				ID:        "100",
				Name:      "john",
				Admin:     false,
				Resources: resources,
			}))

			return next(c)
		}
	}
}

func BuildJwt(identity Identity) *jwt.Token {
	tokenExpiration := 1000

	// Set custom claims
	claims := &JWTCustomClaims{
		identity.GetID(),
		identity.GetName(),
		identity.IsAdmin(),
		identity.GetResources(),
		0,
		identity.IsPasswordChangeRequired(),
		jwt.RegisteredClaims{
			Subject:   identity.GetID(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(tokenExpiration))),
		},
	}
	// Create token with claims
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
}
