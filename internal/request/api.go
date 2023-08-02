package request

import (
	"net/http"

	"github.com/joinself/restful-client/internal/connection"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/labstack/echo/v4"
)

// RegisterHandlers sets up the routing of the HTTP handlers.
func RegisterHandlers(r *echo.Group, service Service, cService connection.Service, authHandler echo.MiddlewareFunc, logger log.Logger) {
	res := resource{service, cService, logger}

	r.Use(authHandler)

	r.GET("/apps/:app_id/connections/:connection_id/requests/:id", res.get)
	r.POST("/apps/:app_id/connections/:connection_id/requests", res.create)
}

type resource struct {
	service  Service
	cService connection.Service
	logger   log.Logger
}

// GetConnection godoc
// @Summary      Get request details.
// @Description  Get request details by request request id.
// @Tags         requests
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        app_id   path      string  true  "App id"
// @Param        connection_id   path      string  true  "Connection id"
// @Param        id   path      int  true  "Request request id"
// @Success      200  {object}  Request
// @Router       /apps/{app_id}/connections/{connection_id}/requests/{id} [get]
func (r resource) get(c echo.Context) error {
	request, err := r.service.Get(c.Request().Context(), c.Param("app_id"), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, request)
}

// CreateConnection godoc
// @Summary         Sends a request request.
// @Description  	Sends a request request to the specified self user.
// @Tags            requests
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           app_id   path      string  true  "App id"
// @Param           connection_id  path string  true  "Connection id"
// @Param           request body CreateRequest true "query params"
// @Success         200  {object}  connection.Connection
// @Router          /apps/{app_id}/connections/{connection_id}/requests [post]
func (r resource) create(c echo.Context) error {
	var input CreateRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return c.JSON(http.StatusBadRequest, "")
	}

	// Get the connection id
	conn, err := r.cService.Get(c.Request().Context(), c.Param("app_id"), c.Param("connection_id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, "connection not found, create a new connection first")
	}

	request, err := r.service.Create(c.Request().Context(), c.Param("app_id"), c.Param("connection_id"), conn.ID, input)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, request)
}
