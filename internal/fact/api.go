package fact

import (
	"net/http"

	"github.com/joinself/restful-client/internal/connection"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/pagination"
	"github.com/joinself/restful-client/pkg/response"
	"github.com/labstack/echo/v4"
)

// RegisterHandlers sets up the routing of the HTTP handlers.
func RegisterHandlers(r *echo.Group, service Service, cService connection.Service, logger log.Logger) {
	res := resource{service, cService, logger}

	r.GET("/:app_id/connections/:connection_id/facts", res.query)
	r.POST("/:app_id/connections/:connection_id/facts", res.create)
	r.GET("/:app_id/connections/:connection_id/facts/:id", res.get)
	r.DELETE("/:app_id/connections/:connection_id/facts/:id", res.delete)

	// r.PUT("/apps/:app_id/connections/:connection_id/facts/:id", res.update)
}

type resource struct {
	service  Service
	cService connection.Service
	logger   log.Logger
}

// GetFact godoc
// @Summary Retrieve specific fact details
// @Description This endpoint retrieves the details of a specific fact using the provided app_id, connection_id and fact request id.
// @Tags facts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param app_id path string true "Unique Identifier for the Application"
// @Param connection_id path string true "Unique Identifier for the Connection"
// @Param id path string true "Unique Identifier for the Fact Request"
// @Success 200 {object} ExtFact "Successfully retrieved the fact details"
// @Failure 404 {object} response.Error "The requested fact could not be found or you don't have permission to access it"
// @Router /apps/{app_id}/connections/{connection_id}/facts/{id} [get]
func (r resource) get(c echo.Context) error {
	conn, err := r.cService.Get(c.Request().Context(), c.Param("app_id"), c.Param("connection_id"))
	if err != nil {
		return c.JSON(response.DefaultNotFoundError())
	}

	f, err := r.service.Get(c.Request().Context(), conn.ID, c.Param("id"))
	if err != nil {
		println("......")
		println(err.Error())
		println("......")
		return c.JSON(response.DefaultNotFoundError())
	}

	return c.JSON(http.StatusOK, NewExtFact(f))
}

// ListFacts godoc
// @Summary Retrieve facts based on filters
// @Description This endpoint retrieves a list of facts using the provided app_id, connection_id, and other optional filters. The results can be paginated using page and per_page parameters.
// @Tags facts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param app_id path string true "Unique Identifier for the Application"
// @Param connection_id path string true "Unique Identifier for the Connection"
// @Param page query int false "Page number for the results pagination"
// @Param per_page query int false "Number of results per page"
// @Param source query string false "Filter by source of the fact"
// @Param fact query string false "Filter by fact"
// @Success 200 {object} ExtListResponse "Successfully retrieved the list of facts"
// @Failure 404 {object} response.Error "The requested resource could not be found or you don't have permission to access it"
// @Failure 500 {object} response.Error "There was a problem with your request. Please try again"
// @Router /apps/{app_id}/connections/{connection_id}/facts [get]
func (r resource) query(c echo.Context) error {
	ctx := c.Request().Context()

	// Get the connection id
	conn, err := r.cService.Get(c.Request().Context(), c.Param("app_id"), c.Param("connection_id"))
	if err != nil {
		return c.JSON(response.DefaultNotFoundError())
	}

	cid := conn.ID
	sid := c.QueryParam("source")
	fid := c.QueryParam("fact")

	count, err := r.service.Count(ctx, cid, sid, fid)
	if err != nil {
		return c.JSON(response.DefaultInternalError(c, r.logger, err.Error()))
	}
	pages := pagination.NewFromRequest(c.Request(), count)
	facts, err := r.service.Query(ctx, cid, sid, fid, pages.Offset(), pages.Limit())
	if err != nil {
		return c.JSON(response.DefaultInternalError(c, r.logger, err.Error()))
	}
	items := []ExtFact{}
	for _, f := range facts {
		items = append(items, NewExtFact(f))
	}
	pages.Items = items
	return c.JSON(http.StatusOK, pages)
}

// IssueFact godoc
// @Summary Issue a new fact to a connection
// @Description This endpoint issues a new fact to a specific connection using the provided app_id, connection_id and the request body.
// @Tags facts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param app_id path string true "Unique Identifier for the Application"
// @Param connection_id path string true "Unique Identifier for the Connection"
// @Param request body CreateFactRequestDoc true "Body containing the details of the fact to be issued"
// @Success 201 {string} string "Fact successfully issued"
// @Failure 400 {object} response.Error "Invalid input - the provided body is not valid"
// @Failure 404 {object} response.Error "Not found - the requested resource does not exist, or you don't have permissions to access it"
// @Failure 500 {object} response.Error "Internal error - there was a problem with your request. Please try again"
// @Router /apps/{app_id}/connections/{connection_id}/facts [post]
func (r resource) create(c echo.Context) error {
	var input CreateFactRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return c.JSON(response.DefaultBadRequestError())
	}

	if reqErr := input.Validate(); reqErr != nil {
		return c.JSON(reqErr.Status, reqErr)
	}

	ctx := c.Request().Context()
	// Get the connection id
	conn, err := r.cService.Get(ctx, c.Param("app_id"), c.Param("connection_id"))
	if err != nil {
		return c.JSON(response.DefaultNotFoundError())
	}

	err = r.service.Create(ctx, c.Param("app_id"), c.Param("connection_id"), conn.ID, input)
	if err != nil {
		return c.JSON(response.DefaultInternalError(c, r.logger, err.Error()))
	}

	return c.NoContent(http.StatusAccepted)
}

// TODO : Consider removing this endpoint
func (r resource) update(c echo.Context) error {
	var input UpdateFactRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return c.JSON(response.DefaultBadRequestError())
	}

	if reqErr := input.Validate(); reqErr != nil {
		return c.JSON(reqErr.Status, reqErr)
	}

	conn, err := r.cService.Get(c.Request().Context(), c.Param("app_id"), c.Param("connection_id"))
	if err != nil {
		return c.JSON(response.DefaultNotFoundError())
	}

	fact, err := r.service.Update(c.Request().Context(), conn.ID, c.Param("id"), input)
	if err != nil {
		return c.JSON(response.DefaultInternalError(c, r.logger, err.Error()))
	}

	return c.JSON(http.StatusOK, fact)
}

// DeleteFact godoc
// @Summary Deletes a fact
// @Description Deletes an existing fact for a specific connection identified by app_id, connection_id and the id of the fact to be deleted.
// @Tags facts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param app_id path string true "Unique Identifier for the Application"
// @Param connection_id path string true "Unique Identifier for the Connection"
// @Param id path string true "Unique Identifier for the Fact to be deleted"
// @Success 204 {string} string "Fact successfully deleted"
// @Failure 404 {object} response.Error "The requested resource does not exist, or you don't have permissions to access it"
// @Router /apps/{app_id}/connections/{connection_id}/facts/{id} [delete]
func (r resource) delete(c echo.Context) error {
	conn, err := r.cService.Get(c.Request().Context(), c.Param("app_id"), c.Param("connection_id"))
	if err != nil {
		return c.JSON(response.DefaultNotFoundError())
	}

	err = r.service.Delete(c.Request().Context(), conn.ID, c.Param("id"))
	if err != nil {
		return c.JSON(response.DefaultNotFoundError())
	}

	return c.NoContent(http.StatusNoContent)
}
