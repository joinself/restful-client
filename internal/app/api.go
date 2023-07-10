package app

import (
	"net/http"

	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/pagination"
	selfsdk "github.com/joinself/self-go-sdk"
	"github.com/labstack/echo/v4"
)

// RegisterHandlers sets up the routing of the HTTP handlers.
func RegisterHandlers(r *echo.Group, clients map[string]*selfsdk.Client, authHandler echo.MiddlewareFunc, logger log.Logger) {
	res := resource{logger, clients}

	// the following endpoints require a valid JWT
	r.Use(authHandler)

	r.GET("/apps", res.list)
}

type resource struct {
	logger  log.Logger
	clients map[string]*selfsdk.Client
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
// @Description    List restful client configured apps.
// @Tags           apps
// @Accept         json
// @Produce        json
// @Security       BearerAuth
// @Success        200  {object} response
// @Router         /apps [get]
func (r resource) list(c echo.Context) error {
	apps := []app{}
	for id := range r.clients {
		apps = append(apps, app{ID: id})
	}

	pages := pagination.NewFromRequest(c.Request(), len(apps))
	pages.Items = apps

	return c.JSON(http.StatusOK, pages)
}
