package test

import (
	"net/http"
	"net/http/httptest"

	"github.com/joinself/restful-client/pkg/log"
	"github.com/labstack/echo/v4"
)

// MockRoutingContext creates a routing.Conext for testing handlers.
func MockRoutingContext(req *http.Request) (echo.Context, *httptest.ResponseRecorder) {
	res := httptest.NewRecorder()
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	e := echo.New()
	ctx := e.NewContext(req, res)

	return ctx, res
}

// MockRouter creates a routing.Router for testing APIs.
func MockRouter(logger log.Logger) *echo.Echo {
	e := echo.New()
	// e.Use(middleware.Recover())

	return e
}
