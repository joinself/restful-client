package request

import (
	"net/http"

	"github.com/joinself/restful-client/internal/connection"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/response"
	"github.com/labstack/echo/v4"
)

// RegisterHandlers sets up the routing of the HTTP handlers.
func RegisterHandlers(r *echo.Group, service Service, cService connection.Service, logger log.Logger) {
	res := resource{service, cService, logger}

	r.GET("/:app_id/requests/:id", res.get)
	r.POST("/:app_id/requests", res.create)
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
// @Param        id   path      int  true  "Request request id"
// @Success      200  {object}  ExtRequest
// @Router       /apps/{app_id}/requests/{id} [get]
func (r resource) get(c echo.Context) error {
	request, err := r.service.Get(c.Request().Context(), c.Param("app_id"), c.Param("id"))
	if err != nil {
		return c.JSON(response.DefaultNotFoundError())
	}

	request.AppID = c.Param("app_id")
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
// @Param           request body CreateRequest true "query params"
// @Success         200  {object}  ExtRequest
// @Router          /apps/{app_id}/requests [post]
func (r resource) create(c echo.Context) error {
	var input CreateRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return c.JSON(response.DefaultBadRequestError())
	}

	if reqErr := input.Validate(); reqErr != nil {
		return c.JSON(reqErr.Status, reqErr)
	}

	// Get the connection id
	var co entity.Connection
	conn, err := r.cService.Get(c.Request().Context(), c.Param("app_id"), input.SelfID)
	if err == nil {
		co = entity.Connection{
			ID:     conn.ID,
			SelfID: conn.SelfID,
		}
	}

	request, err := r.service.Create(c.Request().Context(), c.Param("app_id"), &co, input)
	if err != nil {
		return c.JSON(response.DefaultInternalError(c, r.logger, err.Error()))
	}

	request.AppID = c.Param("app_id")
	return c.JSON(http.StatusAccepted, request)
}
