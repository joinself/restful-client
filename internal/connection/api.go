package connection

import (
	"net/http"

	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/pagination"
	"github.com/labstack/echo/v4"
)

// RegisterHandlers sets up the routing of the HTTP handlers.
func RegisterHandlers(r *echo.Group, service Service, authHandler echo.MiddlewareFunc, logger log.Logger) {
	res := resource{service, logger}

	// the following endpoints require a valid JWT
	r.Use(authHandler)

	r.GET("/connections/:id", res.get)
	r.GET("/connections", res.query)

	r.POST("/connections", res.create)
	r.PUT("/connections/:id", res.update)
	r.DELETE("/connections/:id", res.delete)
}

type resource struct {
	service Service
	logger  log.Logger
}

func (r resource) get(c echo.Context) error {
	connection, err := r.service.Get(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, connection)
}

func (r resource) query(c echo.Context) error {
	ctx := c.Request().Context()
	count, err := r.service.Count(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	pages := pagination.NewFromRequest(c.Request(), count)
	connections, err := r.service.Query(ctx, pages.Offset(), pages.Limit())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	pages.Items = connections
	return c.JSON(http.StatusOK, pages)
}

func (r resource) create(c echo.Context) error {
	var input CreateConnectionRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return c.JSON(http.StatusBadRequest, "")
	}

	connection, err := r.service.Create(c.Request().Context(), input)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, connection)
}

func (r resource) update(c echo.Context) error {
	var input UpdateConnectionRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	connection, err := r.service.Update(c.Request().Context(), c.Param("id"), input)
	if err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, connection)
}

func (r resource) delete(c echo.Context) error {
	connection, err := r.service.Delete(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, connection)
}
