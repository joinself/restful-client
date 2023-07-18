package fact

import (
	"net/http"

	"github.com/joinself/restful-client/internal/connection"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/pagination"
	"github.com/labstack/echo/v4"
)

// RegisterHandlers sets up the routing of the HTTP handlers.
func RegisterHandlers(r *echo.Group, service Service, cService connection.Service, authHandler echo.MiddlewareFunc, logger log.Logger) {
	res := resource{service, cService, logger}

	r.Use(authHandler)

	r.GET("/apps/:app_id/connections/:connection_id/facts", res.query)
	r.POST("/apps/:app_id/connections/:connection_id/facts", res.create)
	r.GET("/apps/:app_id/connections/:connection_id/facts/:id", res.get)
	r.DELETE("/apps/:app_id/connections/:connection_id/facts/:id", res.delete)

	// r.PUT("/apps/:app_id/connections/:connection_id/facts/:id", res.update)
}

type resource struct {
	service  Service
	cService connection.Service
	logger   log.Logger
}

// GetConnection godoc
// @Summary      Get fact details.
// @Description  Get fact details by fact request id.
// @Tags         facts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        app_id   path      string  true  "App id"
// @Param        connection_id   path      int  true  "Connection id"
// @Param        id   path      int  true  "Fact request id"
// @Success      200  {object}  Fact
// @Router       /apps/:app_id/connections/{connection_id}/facts/{id} [get]
func (r resource) get(c echo.Context) error {
	fact, err := r.service.Get(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, fact)
}

func (r resource) query(c echo.Context) error {
	ctx := c.Request().Context()

	// Get the connection id
	conn, err := r.cService.Get(c.Request().Context(), c.Param("app_id"), c.Param("connection_id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}

	cid := conn.ID
	sid := c.QueryParam("source")
	fid := c.QueryParam("fact")

	count, err := r.service.Count(ctx, cid, sid, fid)
	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}
	pages := pagination.NewFromRequest(c.Request(), count)
	facts, err := r.service.Query(ctx, cid, sid, fid, pages.Offset(), pages.Limit())
	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}
	pages.Items = facts
	return c.JSON(http.StatusOK, pages)
}

// WARNING: Do not use for code purposes, this is only used to generate
// the documentation for the openapi, which seems to be broken for nested
// structs.
type CreateFactRequestDoc struct {
	Facts []struct {
		Key    string `json:"key"`
		Value  string `json:"value"`
		Source string `json:"source"`
		Group  *struct {
			Name string `json:"name"`
			Icon string `json:"icon"`
		} `json:"group,omitempty"`
		Type string `json:"type,omitempty"`
	} `json:"facts"`
}

// CreateConnection godoc
// @Summary         Issues a fact.
// @Description  	Issues a fact to one of your connections.
// @Tags            facts
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           app_id   path      string  true  "App id"
// @Param           connection_id  path string  true  "Connection id"
// @Param           request body CreateFactRequestDoc true "query params"
// @Success         200
// @Router          /apps/{app_id}/connections/{connection_id}/facts [post]
func (r resource) create(c echo.Context) error {
	var input CreateFactRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return c.JSON(http.StatusBadRequest, "")
	}

	ctx := c.Request().Context()
	// Get the connection id
	conn, err := r.cService.Get(ctx, c.Param("app_id"), c.Param("connection_id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}

	err = r.service.Create(ctx, c.Param("app_id"), c.Param("connection_id"), conn.ID, input)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, ``)
}

func (r resource) update(c echo.Context) error {
	var input UpdateFactRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return c.JSON(http.StatusBadRequest, "")
	}

	fact, err := r.service.Update(c.Request().Context(), c.Param("id"), input)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, fact)
}

func (r resource) delete(c echo.Context) error {
	fact, err := r.service.Delete(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, fact)
}
