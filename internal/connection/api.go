package connection

import (
	"net/http"

	"github.com/gofrs/uuid"
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
// @Summary      Get connection details.
// @Description  Get connection details by selfID.
// @Tags         connections
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        app_id   path      string  true  "App id"
// @Param        id   path      int  true  "current connection id"
// @Success      200  {object}  ExtConnection
// @Router       /apps/{app_id}/connections/{id} [get]
func (r resource) get(c echo.Context) error {
	conn, err := r.service.Get(c.Request().Context(), c.Param("app_id"), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, response.Error{
			Status:  http.StatusNotFound,
			Error:   "Not found",
			Details: "The requested resource does not exist, or you don't have permissions to access it",
		})
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
// @Summary        List connections.
// @Description    List connections matching the specified filters.
// @Tags           connections
// @Accept         json
// @Produce        json
// @Security       BearerAuth
// @Param          app_id   path      string  true  "App id"
// @Param          page query int false "page number"
// @Param          per_page query int false "number of elements per page"
// @Success        200  {object}  ExtListResponse
// @Router         /apps/{app_id}/connections [get]
func (r resource) query(c echo.Context) error {
	ctx := c.Request().Context()
	count, err := r.service.Count(ctx, c.Param("app_id"))
	if err != nil {
		errorCode, _ := uuid.NewV4()
		r.logger.With(c.Request().Context()).Info("[%s] %s", errorCode, err.Error())
		return c.JSON(http.StatusInternalServerError, response.Error{
			Status:  http.StatusInternalServerError,
			Error:   "Internal error",
			Details: "There was a problem with your request. Error code [" + errorCode.String() + "]",
		})
	}

	pages := pagination.NewFromRequest(c.Request(), count)
	connections, err := r.service.Query(ctx,
		c.Param("app_id"),
		pages.Offset(),
		pages.Limit())
	if err != nil {
		errorCode, _ := uuid.NewV4()
		r.logger.With(c.Request().Context()).Info("[%s] %s", errorCode, err.Error())
		return c.JSON(http.StatusInternalServerError, response.Error{
			Status:  http.StatusInternalServerError,
			Error:   "Internal error",
			Details: "There was a problem with your request. Error code [" + errorCode.String() + "]",
		})
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
// @Summary         Creates a new connection.
// @Description  	Creates a new connection and sends a request for public information.
// @Tags            connections
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           app_id   path      string  true  "App id"
// @Param           request body CreateConnectionRequest true "query params"
// @Success         200  {object}  ExtConnection
// @Router          /apps/{app_id}/connections [post]
func (r resource) create(c echo.Context) error {
	var input CreateConnectionRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return c.JSON(http.StatusBadRequest, response.Error{
			Status:  http.StatusBadRequest,
			Error:   "Invalid input",
			Details: "The provided body is not valid",
		})
	}

	if reqErr := input.Validate(); reqErr != nil {
		return c.JSON(reqErr.Status, reqErr)
	}

	conn, err := r.service.Create(c.Request().Context(), c.Param("app_id"), input)
	if err != nil {
		errorCode, _ := uuid.NewV4()
		r.logger.With(c.Request().Context()).Info("[%s] %s", errorCode, err.Error())
		return c.JSON(http.StatusInternalServerError, response.Error{
			Status:  http.StatusInternalServerError,
			Error:   "Internal error",
			Details: "There was a problem with your request. Error code [" + errorCode.String() + "]",
		})
	}

	return c.JSON(http.StatusCreated, ExtConnection{
		ID:        conn.SelfID,
		AppID:     conn.AppID,
		Name:      conn.Name,
		CreatedAt: conn.CreatedAt,
		UpdatedAt: conn.UpdatedAt,
	})
}

// CreateConnection godoc
// @Summary         Updates a connection.
// @Description  	Updates the properties of an existing connection..
// @Tags            connections
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           app_id   path      string  true  "App id"
// @Param           id   path      int  true  "current connection id"
// @Param           request body UpdateConnectionRequest true "query params"
// @Success         200  {object}  ExtConnection
// @Router          /apps/{app_id}/connections/{id} [put]
func (r resource) update(c echo.Context) error {
	var input UpdateConnectionRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return c.JSON(http.StatusBadRequest, response.Error{
			Status:  http.StatusBadRequest,
			Error:   "Invalid input",
			Details: "The provided body is not valid",
		})
	}

	if reqErr := input.Validate(); reqErr != nil {
		return c.JSON(reqErr.Status, reqErr)
	}

	conn, err := r.service.Update(c.Request().Context(), c.Param("app_id"), c.Param("id"), input)
	if err != nil {
		errorCode, _ := uuid.NewV4()
		r.logger.With(c.Request().Context()).Info("[%s] %s", errorCode, err.Error())
		return c.JSON(http.StatusInternalServerError, response.Error{
			Status:  http.StatusInternalServerError,
			Error:   "Internal error",
			Details: "There was a problem with your request. Error code [" + errorCode.String() + "]",
		})
	}

	return c.JSON(http.StatusOK, ExtConnection{
		ID:        conn.SelfID,
		AppID:     conn.AppID,
		Name:      conn.Name,
		CreatedAt: conn.CreatedAt,
		UpdatedAt: conn.UpdatedAt,
	})
}

// UpdateConnection godoc
// @Summary         Deletes an existing connection.
// @Description  	Deletes an existing connection and sends a request for public information and avoids incoming comms from that connection.
// @Tags            connections
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           app_id   path      string  true  "App id"
// @Param           id   path      int  true  "current connection id"
// @Param           request body CreateConnectionRequest true "query params"
// @Success         200  {object}  ExtConnection
// @Router          /apps/{app_id}/connections/{id} [delete]
func (r resource) delete(c echo.Context) error {
	conn, err := r.service.Delete(c.Request().Context(), c.Param("app_id"), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, response.Error{
			Status:  http.StatusNotFound,
			Error:   "Not found",
			Details: "The requested resource does not exist, or you don't have permissions to access it",
		})
	}

	return c.JSON(http.StatusOK, ExtConnection{
		ID:        conn.SelfID,
		AppID:     conn.AppID,
		Name:      conn.Name,
		CreatedAt: conn.CreatedAt,
		UpdatedAt: conn.UpdatedAt,
	})
}
