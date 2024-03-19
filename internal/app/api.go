package app

import (
	"net/http"

	"github.com/joinself/restful-client/pkg/acl"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/pagination"
	"github.com/joinself/restful-client/pkg/response"
	"github.com/labstack/echo/v4"
)

// RegisterHandlers sets up the routing of the HTTP handlers.
func RegisterHandlers(r *echo.Group, s Service, logger log.Logger) {
	res := resource{logger, s}

	r.GET("", res.list)
	r.POST("", res.create)
	r.DELETE("/:id", res.delete)
	r.POST("/:id/config", res.config)
}

type resource struct {
	logger  log.Logger
	service Service
}

// ListApps godoc
// @Summary        Lists all configured apps.
// @Description    Retrieves and lists all the configured apps for the restful client. You must be authenticated as an admin.
// @Tags           apps
// @Accept         json
// @Produce        json
// @Security       BearerAuth
// @Success        200 {object} ExtListResponse "Successful operation"
// @Failure        404 {object} response.Error "Not found - The requested resource does not exist, or you don't have permissions to access it"
// @Router         /apps [get]
func (r resource) list(c echo.Context) error {
	apps := r.service.List(c.Request().Context())
	pages := pagination.NewFromRequest(c.Request(), len(apps))
	pages.Items = apps

	return c.JSON(http.StatusOK, pages)
}

// CreateApp godoc
// @Summary         Creates a new app.
// @Description     Creates a new app with the given parameters. You must be authenticated as an admin.
// @Tags            apps
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           request body CreateAppRequest true "Details of the new app to create"
// @Success         201  {object}  ExtApp "Successfully created app details"
// @Failure         400 {object} response.Error "Bad request - The provided body is not valid"
// @Failure         404 {object} response.Error "Not found - The requested resource does not exist, or you don't have permissions to access it"
// @Failure         500 {object} response.Error "Internal error - There was a problem with your request"
// @Router          /apps [post]
func (r resource) create(c echo.Context) error {
	var input CreateAppRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return c.JSON(response.DefaultBadRequestError())
	}

	if reqErr := input.Validate(); reqErr != nil {
		return c.JSON(reqErr.Status, reqErr)
	}

	a, err := r.service.Create(c.Request().Context(), input)
	if err != nil {
		return c.JSON(response.DefaultInternalError(c, r.logger, err.Error()))
	}

	return c.JSON(http.StatusOK, ExtApp{
		ID:     a.ID,
		Name:   a.Name,
		Status: a.Status,
		Env:    a.Env,
	})
}

// DeleteApp godoc
// @Summary         Deletes an existing app.
// @Description     Deletes an existing app and sends a request for public information and avoids incoming comms from that app. You must be authenticated as an admin.
// @Tags            apps
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           id   path      int  true  "ID of the app to delete"
// @Success         204  {string} string  "No Content"
// @Failure         404 {object} response.Error "Not found - The requested resource does not exist, or you don't have permissions to access it"
// @Router          /apps/{id} [delete]
func (r resource) delete(c echo.Context) error {
	user := acl.CurrentUser(c)
	if user == nil || !user.IsAdmin() {
		return c.JSON(response.DefaultNotFoundError())
	}

	if _, err := r.service.Delete(c.Request().Context(), c.Param("id")); err != nil {
		return c.JSON(response.DefaultNotFoundError())
	}

	return c.NoContent(http.StatusOK)
}

// SetAppConfig godoc
// @Summary Update the configuration of a specific app
// @Description This endpoint allows you to update the configuration of a specific app. The user must be an admin to perform this action.
// @Tags apps
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param id path int true "ID of the app to be configured. This is a unique identifier."
// @Param config body AppConfig true "Configuration details for the app. Contains the new settings that will be applied."
// @Success 204 {string} string "Successfully updated the app configuration. No Content is returned."
// @Failure 400 {object} response.Error "Bad Request - The request could not be understood or was missing required parameters."
// @Failure 404 {object} response.Error "Not Found - The app with the specified ID does not exist, or the user does not have the required permissions to access it."
// @Router /apps/{id}/config [post]
func (r resource) config(c echo.Context) error {
	user := acl.CurrentUser(c)
	if user == nil || !user.IsAdmin() {
		return c.JSON(response.DefaultNotFoundError())
	}

	var input AppConfig
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return c.JSON(response.DefaultBadRequestError())
	}

	r.service.SetConfig(c.Request().Context(), c.Param("id"), input)

	return c.NoContent(http.StatusAccepted)
}
