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
// @Summary         Retrieve request details
// @Description     Get detailed information about a request using the request ID.
// @Tags            Requests
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           app_id  path      string  true  "Application ID"
// @Param           id      path      int     true  "Request ID"
// @Success         200     {object}  ExtRequest       "Successful Response"
// @Failure         404     {object}  response.Error    "Request Not Found"
// @Failure         500     {object}  response.Error    "Internal Server Error"
// @Router          /apps/{app_id}/requests/{id} [get]
func (r resource) get(c echo.Context) error {
	request, err := r.service.Get(c.Request().Context(), c.Param("app_id"), c.Param("id"))
	if err != nil {
		r.logger.With(c.Request().Context()).Warnf("error retrieving request : %s", err.Error())
		return c.JSON(response.DefaultNotFoundError())
	}

	request.AppID = c.Param("app_id")
	return c.JSON(http.StatusOK, request)
}

// CreateConnection godoc
// @Summary         Send a fact or authentication request
// @Description     This endpoint allows you to send a fact or authentication request to a specified self user.
// @Tags            Requests
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           app_id   path      string           true  "Application ID"
// @Param           request  body      CreateRequest    true  "Request Body"
// @Success         200      {object}  ExtRequest       "Successful Response"
// @Failure         400      {object}  response.Error    "Invalid Request"
// @Failure         500      {object}  response.Error    "Internal Server Error"
// @Router          /apps/{app_id}/requests [post]
func (r resource) create(c echo.Context) error {
	ctx := c.Request().Context()
	var input CreateRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(ctx).Warnf("error invalid input: %s", err.Error())
		return c.JSON(response.DefaultBadRequestError())
	}

	if err := input.Validate(); err != nil {
		r.logger.With(ctx).Infof("error invalid input: %s", err.Error)
		return c.JSON(err.Status, err)
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
		r.logger.With(ctx).Warnf("error creating request: %s", err.Error())
		return c.JSON(response.DefaultInternalError(c, r.logger, err.Error()))
	}

	request.AppID = c.Param("app_id")
	return c.JSON(http.StatusAccepted, request)
}
