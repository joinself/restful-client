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
	r.PUT("/:app_id", res.update)
	r.DELETE("/:app_id", res.delete)
	r.GET("/:app_id", res.get)
}

type resource struct {
	logger  log.Logger
	service Service
}

// ListApps godoc
// @Summary         List All Applications
// @Description     Retrieves and lists all the configured applications accessible to the authenticated user. This operation requires the user to be authenticated with administrative privileges.
// @Tags            Applications
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Success         200 {object} ExtListResponse "The operation was successful. The response contains the details of all the configured applications."
// @Failure         404 {object} response.Error "Resource Not Found - The requested resources do not exist, or the authenticated user does not have sufficient permissions to access them."
// @Router          /apps [get]
func (r resource) list(c echo.Context) error {
	apps := r.service.List(c.Request().Context())
	pages := pagination.NewFromRequest(c.Request(), len(apps))
	pages.Items = apps

	return c.JSON(http.StatusOK, pages)
}

// CreateApp godoc
// @Summary         Create an Application
// @Description     Creates a new application with the provided details. This operation requires the user to be authenticated with administrative privileges.
// @Tags            Applications
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           request body CreateAppRequest true "The details of the new application to be created."
// @Success         201  {object}  ExtApp "A successful response returns the details of the newly created application."
// @Failure         400 {object} response.Error "Bad Request - The body of the request is not valid or incorrectly formatted."
// @Failure         404 {object} response.Error "Resource Not Found - The requested resource does not exist, or the authenticated user does not have sufficient permissions to access it."
// @Failure         500 {object} response.Error "Internal Server Error - An error occurred while processing the request."
// @Router          /apps [post]
func (r resource) create(c echo.Context) error {
	var input CreateAppRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Warnf("invalid request: %s", err.Error())
		return c.JSON(response.DefaultBadRequestError())
	}

	if err := input.Validate(); err != nil {
		r.logger.With(c.Request().Context()).Warnf("invalid request: %s", err.Details)
		return c.JSON(err.Status, err)
	}

	a, err := r.service.Create(c.Request().Context(), input)
	if err != nil {
		r.logger.With(c.Request().Context()).Warnf("err creating app - %v", err)
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
// @Summary         Delete an Application
// @Description     Deletes an existing application identified by the provided app_id. This operation will also send a request to update public information and prevent further communications from the deleted application. Only users authenticated with administrative privileges can perform this operation.
// @Tags            Applications
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           app_id   path   int  true  "The unique identifier (ID) of the application to be deleted."
// @Success         204  {string} string  "Successful operation - the application has been deleted, and no content is returned."
// @Failure         404 {object} response.Error "Resource not found - The requested application does not exist, or the authenticated user does not have sufficient permissions to access it."
// @Router          /apps/{app_id} [delete]
func (r resource) delete(c echo.Context) error {
	user := acl.CurrentUser(c)
	if user == nil || !user.IsAdmin() {
		r.logger.With(c.Request().Context()).Info("insufficient permissions for deleting an app")
		return c.JSON(response.DefaultNotFoundError())
	}

	if _, err := r.service.Delete(c.Request().Context(), c.Param("app_id")); err != nil {
		r.logger.With(c.Request().Context()).Warnf("err deleting an api key - %v", err)
		return c.JSON(response.DefaultNotFoundError())
	}

	return c.NoContent(http.StatusOK)
}

// UpdateApp godoc
// @Summary         Updates an Application
// @Description     Updates an application with the provided details.
// @Tags            Applications
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           request body UpdateAppRequest true "The details of the new application to be updated."
// @Success         201  {object}  ExtApp "A successful response returns the details of the newly updated application."
// @Failure         400 {object} response.Error "Bad Request - The body of the request is not valid or incorrectly formatted."
// @Failure         404 {object} response.Error "Resource Not Found - The requested resource does not exist, or the authenticated user does not have sufficient permissions to access it."
// @Failure         500 {object} response.Error "Internal Server Error - An error occurred while processing the request."
// @Router          /apps [post]
func (r resource) update(c echo.Context) error {
	var input UpdateAppRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Warnf("invalid request: %s", err.Error())
		return c.JSON(response.DefaultBadRequestError())
	}

	if err := input.Validate(); err != nil {
		r.logger.With(c.Request().Context()).Warnf("invalid request: %s", err.Details)
		return c.JSON(err.Status, err)
	}

	a, err := r.service.Update(c.Request().Context(), c.Param("app_id"), input)
	if err != nil {
		r.logger.With(c.Request().Context()).Warnf("err creating app - %v", err)
		return c.JSON(response.DefaultNotFoundError())
	}

	return c.JSON(http.StatusOK, ExtApp{
		ID:     a.ID,
		Name:   a.Name,
		Status: a.Status,
		Env:    a.Env,
	})
}

// GetApp godoc
// @Summary         Updates an Application
// @Description     Updates an application with the provided details.
// @Tags            Applications
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           request body UpdateAppRequest true "The details of the new application to be updated."
// @Success         201  {object}  ExtApp "A successful response returns the details of the newly updated application."
// @Failure         400 {object} response.Error "Bad Request - The body of the request is not valid or incorrectly formatted."
// @Failure         404 {object} response.Error "Resource Not Found - The requested resource does not exist, or the authenticated user does not have sufficient permissions to access it."
// @Failure         500 {object} response.Error "Internal Server Error - An error occurred while processing the request."
// @Router          /apps [post]
func (r resource) get(c echo.Context) error {
	a, err := r.service.Get(c.Request().Context(), c.Param("app_id"))
	if err != nil {
		r.logger.With(c.Request().Context()).Warnf("err creating app - %v", err)
		return c.JSON(response.DefaultNotFoundError())
	}

	return c.JSON(http.StatusOK, ExtApp{
		ID:       a.ID,
		Name:     a.Name,
		Status:   a.Status,
		Env:      a.Env,
		Callback: a.Callback,
	})
}
