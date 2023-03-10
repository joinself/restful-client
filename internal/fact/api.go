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

func (r resource) get(c echo.Context) error {
	fact, err := r.service.Get(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, fact)
}

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
