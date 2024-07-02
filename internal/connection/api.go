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
// @Summary         Retrieve Connection Details
// @Description     Retrieves the details of a specified connection, identified by the provided selfID and app_id. User must be authenticated and have sufficient permissions to access this information.
// @Tags            Connections
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           app_id   path   string  true  "The unique identifier (ID) of the application associated with the connection."
// @Param           id   path  int  true  "The unique identifier (ID) of the connection to be retrieved."
// @Success         200  {object}  ExtConnection  "Successful operation. The response contains the details of the requested connection."
// @Failure         404  {object}  response.Error "Resource not found - The requested connection does not exist, or the authenticated user does not have sufficient permissions to access it."
// @Router          /apps/{app_id}/connections/{id} [get]
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
// @Summary         List All Connections
// @Description     Retrieves a list of connections associated with the given app_id, matching the specified filters. Pagination is supported and can be controlled via optional page and per_page parameters.
// @Tags            Connections
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           app_id   path   string  true  "The unique identifier (ID) of the application whose connections are to be retrieved."
// @Param           page query int false "The page number for pagination. Defaults to 1 if not specified."
// @Param           per_page query int false "The number of elements to be displayed per page for pagination. Defaults to 10 if not specified."
// @Success         200  {object}  ExtListResponse  "Successful operation. The response contains a list of connections associated with the specified application."
// @Failure         500  {object}  response.Error "Internal Server Error - An error occurred while processing the request."
// @Router          /apps/{app_id}/connections [get]
func (r resource) query(c echo.Context) error {
	ctx := c.Request().Context()
	count, err := r.service.Count(ctx, c.Param("app_id"))
	if err != nil {
		r.logger.With(ctx).Warnf("there was an error retrieving a count of all apps %v", err)
		return c.JSON(response.DefaultInternalError(c, r.logger, err.Error()))
	}

	pages := pagination.NewFromRequest(c.Request(), count)
	connections, err := r.service.Query(ctx,
		c.Param("app_id"),
		pages.Offset(),
		pages.Limit())
	if err != nil {
		r.logger.With(ctx).Warnf("there was an error retrieving the list of apps %v", err)
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
// @Summary         Create a New Connection
// @Description     Creates a new connection for the specified application (app_id). The details of the new connection are provided in the request body. Once the connection is created, a request for public information is sent.
// @Tags            Connections
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           app_id path string true "The unique identifier (ID) of the application for which the new connection is to be created."
// @Param           request body CreateConnectionRequest true "The body of the request, containing the details of the new connection to be created."
// @Success         201 {object} ExtConnection "Successful operation. The response contains the details of the newly created connection."
// @Failure         400 {object} response.Error "Bad Request - The body of the request is not valid or incorrectly formatted."
// @Failure         500 {object} response.Error "Internal Server Error - An error occurred while processing the request."
// @Router          /apps/{app_id}/connections [post]
func (r resource) create(c echo.Context) error {
	ctx := c.Request().Context()
	var input CreateConnectionRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(ctx).Warnf("error mapping create connection input %v", err)
		return c.JSON(response.DefaultBadRequestError())
	}

	if err := input.Validate(); err != nil {
		r.logger.With(ctx).Warnf("invalid input data when creating a connection %v", err)
		return c.JSON(err.Status, err)
	}

	conn, err := r.service.Create(c.Request().Context(), c.Param("app_id"), input)
	if err != nil {
		r.logger.With(ctx).Warnf("error creating a connection %v", err)
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
// @Summary         Update a Connection
// @Description     Updates the properties of an existing connection, identified by the provided app_id and connection id. The updates are passed in the request body.
// @Tags            Connections
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           app_id path string true "The unique identifier (ID) of the application associated with the connection."
// @Param           id path string true "The unique identifier (ID) of the connection to be updated."
// @Param           request body UpdateConnectionRequest true "The body of the request, containing the details of the connection updates."
// @Success         200 {object} ExtConnection "Successful operation. The response contains the details of the updated connection."
// @Failure         400 {object} response.Error "Bad Request - The body of the request is not valid or incorrectly formatted."
// @Failure         500 {object} response.Error "Internal Server Error - An error occurred while processing the request."
// @Router          /apps/{app_id}/connections/{id} [put]
func (r resource) update(c echo.Context) error {
	ctx := c.Request().Context()
	var input UpdateConnectionRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(ctx).Warnf("problem mapping uddate connection input %v", err)
		return c.JSON(response.DefaultBadRequestError())
	}

	if err := input.Validate(); err != nil {
		r.logger.With(ctx).Warnf("problem validating update connection input %v", err)
		return c.JSON(err.Status, err)
	}

	conn, err := r.service.Update(c.Request().Context(), c.Param("app_id"), c.Param("id"), input)
	if err != nil {
		r.logger.With(ctx).Warnf("problem updating connection %v", err)
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
// @Summary         Delete a Connection
// @Description     Deletes an existing connection, identified by the provided app_id and connection id. Once the connection is deleted, a request for public information is sent, and incoming communications from the deleted connection are stopped.
// @Tags            Connections
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           app_id path string true "The unique identifier (ID) of the application associated with the connection."
// @Param           id path string true "The unique identifier (ID) of the connection to be deleted."
// @Success         200 {object} ExtConnection "Successful operation. The response contains the details of the deleted connection."
// @Failure         404 {object} response.Error "Resource Not Found - The requested connection does not exist, or the authenticated user does not have sufficient permissions to delete it."
// @Router          /apps/{app_id}/connections/{id} [delete]
func (r resource) delete(c echo.Context) error {
	conn, err := r.service.Delete(c.Request().Context(), c.Param("app_id"), c.Param("id"))
	if err != nil {
		r.logger.With(c.Request().Context()).Warnf("problem deleting app %v", err)
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
