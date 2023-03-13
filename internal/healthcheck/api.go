package healthcheck

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// RegisterHandlers registers the handlers that perform healthchecks.
func RegisterHandlers(r *echo.Group, version string) {
	r.GET("/healthcheck", healthcheck)
}

// Healthcheck endpoint
//
//	@Summary		healthcheck endpoint
//	@Description	check the service is up and running
//	@Tags			healthcheck
//	@Accept			json
//	@Produce		json
//	@Success		200	{string}	string	"OK"
//	@Failure		400	{string}	string	"ok"
//	@Failure		404	{string}	string	"ok"
//	@Failure		500	{string}	string	"ok"
//	@Router			/healthcheck [get]
func healthcheck(c echo.Context) error {
	return c.JSON(http.StatusOK, "OK")
}
