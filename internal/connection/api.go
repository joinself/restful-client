package connection

import (
	"net/http"

	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/pagination"
	"github.com/joinself/restful-client/pkg/response"
	"github.com/labstack/echo/v4"
)

// RegisterHandlers sets up the routing of the HTTP handlers.
func RegisterHandlers(r *echo.Group, service Service, logger log.Logger) {
	res := resource{service, logger}

	r.GET("/:app_id/connections/:id", res.get)
	r.GET("/:app_id/connections", res.query)

	r.POST("/:app_id/connections", res.create)
	r.PUT("/:app_id/connections/:id", res.update)
	r.DELETE("/:app_id/connections/:id", res.delete)
}

type resource struct {
	service Service
	logger  log.Logger
}

// GetConnection godoc
// @Summary      Retrieve connection details
// @Description  Retrieves the details of a connection using the given selfID and app_id. Ensure you have sufficient permissions to access this information.
// @Tags         connections
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        app_id   path   string  true  "Unique Identifier for the App"
// @Param        id   path  int  true  "Unique Identifier for the connection"
// @Success      200  {object}  ExtConnection  "Successful retrieval of connection details"
// @Failure      404  {object}  response.Error "Unable to find the requested resource or lack of permissions to access it"
// @Router       /apps/{app_id}/connections/{id} [get]
func (r resource) get(c echo.Context) error {
	conn, err := r.service.Get(c.Request().Context(), c.Param("app_id"), c.Param("id"))
	if err != nil {
		return c.JSON(response.DefaultNotFoundError())
	}

	return c.JSON(http.StatusOK, ExtConnection{
		ID:        conn.SelfID,
		AppID:     conn.AppID,
		Name:      conn.Name,
		CreatedAt: conn.CreatedAt,
		UpdatedAt: conn.UpdatedAt,
	})
}

// ListConnections godoc
// @Summary        Retrieve a list of connections
// @Description    Retrieves a list of connections for a given app_id, matching the specified filters. Pagination is supported with optional page and per_page parameters.
// @Tags           connections
// @Accept         json
// @Produce        json
// @Security       BearerAuth
// @Param          app_id   path   string  true  "Unique Identifier for the App"
// @Param          page query int false "Page number for pagination. Default is 1."
// @Param          per_page query int false "Number of elements per page for pagination. Default is 10."
// @Success        200  {object}  ExtListResponse  "Successful retrieval of connections list"
// @Failure        500  {object}  response.Error "Internal server error occurred during the request"
// @Router         /apps/{app_id}/connections [get]
func (r resource) query(c echo.Context) error {
	ctx := c.Request().Context()
	count, err := r.service.Count(ctx, c.Param("app_id"))
	if err != nil {
		return c.JSON(response.DefaultInternalError(c, r.logger, err.Error()))
	}

	pages := pagination.NewFromRequest(c.Request(), count)
	connections, err := r.service.Query(ctx,
		c.Param("app_id"),
		pages.Offset(),
		pages.Limit())
	if err != nil {
		return c.JSON(response.DefaultInternalError(c, r.logger, err.Error()))
	}

	conns := []ExtConnection{}
	for _, conn := range connections {
		conns = append(conns, ExtConnection{
			ID:        conn.SelfID,
			AppID:     conn.AppID,
			Name:      conn.Name,
			CreatedAt: conn.CreatedAt,
			UpdatedAt: conn.UpdatedAt,
		})
	}

	pages.Items = conns
	return c.JSON(http.StatusOK, pages)
}

// CreateConnection godoc
// @Summary Create a new connection
// @Description This API endpoint creates a new connection by taking the application ID and request body as input. It sends a request for public information once the connection is created.
// @Tags connections
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param app_id path string true "Unique identifier of the application"
// @Param request body CreateConnectionRequest true "Body containing details of the connection to be created"
// @Success 201 {object} ExtConnection "Successfully created a new connection and returns the details of the new connection"
// @Failure 400 {object} response.Error "Returns when the provided input is invalid"
// @Failure 500 {object} response.Error "Returns when there is an internal server error"
// @Router /apps/{app_id}/connections [post]
func (r resource) create(c echo.Context) error {
	var input CreateConnectionRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return c.JSON(response.DefaultBadRequestError())
	}

	if reqErr := input.Validate(); reqErr != nil {
		return c.JSON(reqErr.Status, reqErr)
	}

	conn, err := r.service.Create(c.Request().Context(), c.Param("app_id"), input)
	if err != nil {
		return c.JSON(response.DefaultInternalError(c, r.logger, err.Error()))
	}

	return c.JSON(http.StatusCreated, ExtConnection{
		ID:        conn.SelfID,
		AppID:     conn.AppID,
		Name:      conn.Name,
		CreatedAt: conn.CreatedAt,
		UpdatedAt: conn.UpdatedAt,
	})
}

// UpdateConnection godoc
// @Summary Update a specific connection
// @Description This endpoint updates the properties of an existing connection using the provided app_id, connection id, and the request body.
// @Tags connections
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param app_id path string true "Unique Identifier for the Application"
// @Param id path string true "Unique Identifier for the Connection to be updated"
// @Param request body UpdateConnectionRequest true "Body containing updated details of the connection"
// @Success 200 {object} ExtConnection "Successfully updated the connection and returns the updated connection details"
// @Failure 400 {object} response.Error "Returns when the provided input is invalid"
// @Failure 500 {object} response.Error "There was a problem with your request. Please try again"
// @Router /apps/{app_id}/connections/{id} [put]
func (r resource) update(c echo.Context) error {
	var input UpdateConnectionRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return c.JSON(response.DefaultBadRequestError())
	}

	if reqErr := input.Validate(); reqErr != nil {
		return c.JSON(reqErr.Status, reqErr)
	}

	conn, err := r.service.Update(c.Request().Context(), c.Param("app_id"), c.Param("id"), input)
	if err != nil {
		return c.JSON(response.DefaultInternalError(c, r.logger, err.Error()))
	}

	return c.JSON(http.StatusOK, ExtConnection{
		ID:        conn.SelfID,
		AppID:     conn.AppID,
		Name:      conn.Name,
		CreatedAt: conn.CreatedAt,
		UpdatedAt: conn.UpdatedAt,
	})
}

// DeleteConnection godoc
// @Summary Delete a specific connection
// @Description This endpoint deletes an existing connection using the provided app_id and connection id. After deletion, it sends a request for public information and stops incoming communications from that connection.
// @Tags connections
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param app_id path string true "Unique Identifier for the Application"
// @Param id path string true "Unique Identifier for the Connection to be deleted"
// @Success 200 {object} ExtConnection "Successfully deleted the connection and returns the deleted connection details"
// @Failure 404 {object} response.Error "The requested resource could not be found or you don't have permission to access it"
// @Router /apps/{app_id}/connections/{id} [delete]
func (r resource) delete(c echo.Context) error {
	conn, err := r.service.Delete(c.Request().Context(), c.Param("app_id"), c.Param("id"))
	if err != nil {
		return c.JSON(response.DefaultNotFoundError())
	}

	return c.JSON(http.StatusOK, ExtConnection{
		ID:        conn.SelfID,
		AppID:     conn.AppID,
		Name:      conn.Name,
		CreatedAt: conn.CreatedAt,
		UpdatedAt: conn.UpdatedAt,
	})
}
