package object

import (
	"net/http"

	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/response"
	"github.com/labstack/echo/v4"
)

// RegisterHandlers sets up the routing of the HTTP handlers.
func RegisterHandlers(r *echo.Group, service Service, logger log.Logger) {
	res := resource{service, logger}

	r.POST("/:app_id/objects", res.create)
}

type resource struct {
	service Service
	logger  log.Logger
}

// SendMessage    godoc
// @Summary       Creates an object.
// @Description   Creates a Self shareable object to be used on your chat messages.
// @Tags          objects
// @Accept        json
// @Produce       json
// @Security      BearerAuth
// @Param         app_id   path      string  true  "Application ID"
// @Param         request body CreateObjectRequest true "Request to create an object"
// @Success       200  {object}  Message "Successfully created object"
// @Failure       400  {object}  response.Error "Invalid input"
// @Failure       404  {object}  response.Error "Resource not found or unauthorized access"
// @Failure       500  {object}  response.Error "Internal server error"
// @Router        /apps/{app_id}/objects [post]
func (r resource) create(c echo.Context) error {
	var input CreateObjectRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return c.JSON(response.DefaultBadRequestError())
	}

	if reqErr := input.Validate(); reqErr != nil {
		return c.JSON(reqErr.Status, reqErr)
	}

	resp := CreateObjectResponse{
		Objects: make([]ExtObject, 0),
	}

	for _, o := range input.Objects {
		obj, err := r.service.BuildObject(c.Request().Context(), c.Param("app_id"), o)
		if err != nil {
			return c.JSON(response.DefaultInternalError(c, r.logger, err.Error()))
		}
		resp.Objects = append(resp.Objects, *obj)
	}

	return c.JSON(http.StatusOK, resp)
}
