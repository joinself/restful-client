package healthcheck

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// RegisterHandlers registers the handlers that perform healthchecks.
func RegisterHandlers(r *echo.Echo, version string) {
	r.GET("/healthcheck", healthcheck)
}

// healthcheck responds to a healthcheck request.
func healthcheck(c echo.Context) error {
	return c.JSON(http.StatusOK, "OK")
}
