package acl

import (
	"net/http"

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
