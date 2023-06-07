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

	r.GET("/apps/:app_id/connections/:id", res.get)
	r.GET("/apps/:app_id/connections", res.query)

	r.POST("/apps/:app_id/connections", res.create)
	r.PUT("/apps/:app_id/connections/:id", res.update)
	r.DELETE("/apps/:app_id/connections/:id", res.delete)
}

type resource struct {
	service Service
	logger  log.Logger
}

// GetConnection godoc
// @Summary      Get connection details.
// @Description  Get connection details by selfID.
// @Tags         connections
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Self ID"
// @Success      200  {object}  connection.Connection
// @Router       /connections/{id} [get]
func (r resource) get(c echo.Context) error {
	connection, err := r.service.Get(c.Request().Context(), c.Param("app_id"), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, connection)
}

// ListConnections godoc
// @Summary        List connections.
// @Description    List connections matching the specified filters.
// @Tags           connections
// @Accept         json
// @Produce        json
// @Security       BearerAuth
// @Param          page query int false "page number"
// @Param          per_page query int false "number of elements per page"
// @Success        200  {array}  connection.Connection
// @Router         /connections [get]
func (r resource) query(c echo.Context) error {
	ctx := c.Request().Context()
	count, err := r.service.Count(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	pages := pagination.NewFromRequest(c.Request(), count)
	connections, err := r.service.Query(ctx, c.Param("app_id"), pages.Offset(), pages.Limit())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	pages.Items = connections
	return c.JSON(http.StatusOK, pages)
}

// CreateConnection godoc
// @Summary         Creates a new connection.
// @Description  	Creates a new connection and sends a request for public information.
// @Tags            connections
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           request body CreateConnectionRequest true "query params"
// @Success         200  {object}  connection.Connection
// @Router          /connections [post]
func (r resource) create(c echo.Context) error {
	var input CreateConnectionRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return c.JSON(http.StatusBadRequest, "")
	}

	connection, err := r.service.Create(c.Request().Context(), c.Param("app_id"), input)
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

	connection, err := r.service.Update(c.Request().Context(), c.Param("app_id"), c.Param("id"), input)
	if err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, connection)
}

func (r resource) delete(c echo.Context) error {
	connection, err := r.service.Delete(c.Request().Context(), c.Param("app_id"), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, connection)
}
