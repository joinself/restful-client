package app

import (
	"net/http"

	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/pagination"
	"github.com/labstack/echo/v4"
)

// RegisterHandlers sets up the routing of the HTTP handlers.
func RegisterHandlers(r *echo.Group, s Service, logger log.Logger) {
	res := resource{logger, s}

	r.GET("", res.list)
	r.POST("", res.create)
	r.DELETE("/:id", res.delete)
}

type resource struct {
	logger  log.Logger
	service Service
}

type app struct {
	ID string `json:"id"`
}

type response struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	PageCount  int   `json:"page_count"`
	TotalCount int   `json:"total_count"`
	Items      []app `json:"items"`
}

// ListApps godoc
// @Summary        List apps.
// @Description    List restful client configured apps. You must be authenticated as an admin.
// @Tags           apps
// @Accept         json
// @Produce        json
// @Security       BearerAuth
// @Success        200  {object} response
// @Router         /apps [get]
func (r resource) list(c echo.Context) error {
	apps := r.service.List(c.Request().Context())
	pages := pagination.NewFromRequest(c.Request(), len(apps))
	pages.Items = apps

	return c.JSON(http.StatusOK, pages)
}

// CreateApp godoc
// @Summary         Creates a new app.
// @Description  	Creates a new app and sends a request for public information. You must be authenticated as an admin.
// @Tags            app
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           request body CreateAppRequest true "query params"
// @Success         200  {object}  entity.App
// @Router          /apps [post]
func (r resource) create(c echo.Context) error {
	var input CreateAppRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return c.JSON(http.StatusBadRequest, "")
	}

	app, err := r.service.Create(c.Request().Context(), input)
	if err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return c.JSON(http.StatusBadRequest, "")
	}

	return c.JSON(http.StatusOK, app)
}

// CreateApp godoc
// @Summary         Deletes an existing app.
// @Description  	Deletes an existing app and sends a request for public information and avoids incoming comms from that app. You must be authenticated as an admin.
// @Tags            apps
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           id   path      int  true  "current app id"
// @Param           request body CreateAppRequest true "query params"
// @Success         200  {object}  app.App
// @Router          /apps/{id} [delete]
func (r resource) delete(c echo.Context) error {
	_, err := r.service.Delete(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, "success")
}
