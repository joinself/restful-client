package fact

import (
	"net/http"

	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/pagination"
	"github.com/labstack/echo/v4"
)

// RegisterHandlers sets up the routing of the HTTP handlers.
func RegisterHandlers(r *echo.Group, service Service, authHandler echo.MiddlewareFunc, logger log.Logger) {
	res := resource{service, logger}

	r.Use(authHandler)

	r.GET("/connections/:connection_id/facts/:id", res.get)
	r.GET("/connections/:connection_id/facts", res.query)

	r.POST("/connections/:connection_id/facts", res.create)
	r.PUT("/connections/:connection_id/facts/:id", res.update)
	r.DELETE("/connections/:connection_id/facts/:id", res.delete)
}

type resource struct {
	service Service
	logger  log.Logger
}

// GetConnection godoc
// @Summary      Get fact details.
// @Description  Get fact details by fact request id.
// @Tags         facts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        connection_id   path      int  true  "Connection id"
// @Param        id   path      int  true  "Fact request id"
// @Success      200  {object}  Fact
// @Router       /connections/{connection_id}/facts/{id} [get]
func (r resource) get(c echo.Context) error {
	fact, err := r.service.Get(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, fact)
}

// ListFacts       godoc
// @Summary        List facts.
// @Description    List facts matching the specified filters.
// @Tags           facts
// @Accept         json
// @Produce        json
// @Security       BearerAuth
// @Param          connection_id   path      int  true  "Connection id"
// @Param          source query string  false  "source"  Format(string)
// @Param          fact query string  false  "fact name"  Format(string)
// @Success        200  {array}  connection.Connection
// @Router         /connections/{connection_id}/facts [get]
func (r resource) query(c echo.Context) error {
	ctx := c.Request().Context()

	query := QueryParams{
		Connection: c.Param("connection_id"),
		Source:     c.QueryParam("source"),
		Fact:       c.QueryParam("fact"),
	}

	count, err := r.service.Count(ctx, query)
	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}
	pages := pagination.NewFromRequest(c.Request(), count)
	facts, err := r.service.Query(ctx, query, pages.Offset(), pages.Limit())
	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}
	pages.Items = facts
	return c.JSON(http.StatusOK, pages)
}

// CreateConnection godoc
// @Summary         Sends a fact request.
// @Description  	Sends a fact request to the specified self user.
// @Tags            facts
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           page query int false "page number"
// @Param           per_page query int false "number of elements per page"
// @Param           connection_id  path string  true  "Connection id"
// @Param           request body CreateFactRequest true "query params"
// @Success         200  {object}  connection.Connection
// @Router          /connections/{connection_id}/facts [post]
func (r resource) create(c echo.Context) error {
	var input CreateFactRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return c.JSON(http.StatusBadRequest, "")
	}
	fact, err := r.service.Create(c.Request().Context(), c.Param("connection_id"), input)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, fact)
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
