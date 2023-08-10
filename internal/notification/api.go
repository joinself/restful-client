package notification

import (
	"errors"
	"net/http"

	"github.com/joinself/restful-client/pkg/log"
	"github.com/labstack/echo/v4"
)

// RegisterHandlers sets up the routing of the HTTP handlers.
func RegisterHandlers(r *echo.Group, service Service, authHandler echo.MiddlewareFunc, logger log.Logger) {
	res := resource{service, logger}

	// the following endpoints require a valid JWT
	r.Use(authHandler)

	r.POST("/apps/:app_id/connections/:connection_id/notify", res.create)
}

type resource struct {
	service Service
	logger  log.Logger
}

// CreateNotification godoc
// @Summary         Sends a system notification.
// @Description  	Sends a system notification to the given connection
// @Tags            notifications
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           app_id   path      string  true  "App id"
// @Param           connection_id   path      string  true  "Connection id"
// @Param           request body SystemNotificationData true "system notification"
// @Success         200
// @Router          /apps/{app_id}/connections/{connection_id} [post]
func (r resource) create(c echo.Context) error {
	var input SystemNotificationData
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return c.JSON(http.StatusBadRequest, errors.New("invalid input"))
	}

	err := r.service.Send(c.Request().Context(), c.Param("app_id"), c.Param("connection_id"), input)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusCreated, "")
}
