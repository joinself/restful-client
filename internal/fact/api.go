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
}

type resource struct {
	service  Service
	cService connection.Service
	logger   log.Logger
}

// GetFact godoc
// @Summary         Retrieve Fact Details
// @Description     Retrieves the details of a specific fact, identified by the app_id, connection_id, and fact id.
// @Tags            Facts
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           app_id        path string true "The unique identifier (ID) of the application associated with the connection."
// @Param           connection_id path string true "The unique identifier (ID) of the connection associated with the fact."
// @Param           id            path string true "The unique identifier (ID) of the fact to be retrieved."
// @Success         200 {object}  ExtFact "Successfully retrieved the fact details."
// @Failure         404 {object}  response.Error "Not Found - The requested fact does not exist, or the authenticated user does not have the necessary permissions."
// @Router          /apps/{app_id}/connections/{connection_id}/facts/{id} [get]
func (r resource) get(c echo.Context) error {
	ctx := c.Request().Context()
	conn, err := r.cService.Get(ctx, c.Param("app_id"), c.Param("connection_id"))
	if err != nil {
		r.logger.With(ctx).Warnf("error retrieving fact: %s", err.Error())
		return c.JSON(response.DefaultNotFoundError())
	}

	f, err := r.service.Get(c.Request().Context(), conn.ID, c.Param("id"))
	if err != nil {
		r.logger.With(ctx).Warnf("error retrieving connection: %s", err.Error())
		return c.JSON(response.DefaultNotFoundError())
	}

	return c.JSON(http.StatusOK, NewExtFact(f))
}

// ListFacts godoc
// @Summary         Retrieve Facts with Filters
// @Description     Retrieves a list of facts for a specific connection, identified by the app_id and connection_id, with optional filters for pagination, source, and fact.
// @Tags            Facts
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           app_id        path string true  "The unique identifier (ID) of the application associated with the connection."
// @Param           connection_id path string true  "The unique identifier (ID) of the connection."
// @Param           page          query int    false "Page number for the results pagination."
// @Param           per_page      query int    false "Number of results per page."
// @Param           source        query string false "Filter by the source of the fact."
// @Param           fact          query string false "Filter by the fact."
// @Success         200 {object}  ExtListResponse "Successfully retrieved the list of facts."
// @Failure         404 {object}  response.Error   "Not Found - The requested resource does not exist, or the authenticated user does not have the necessary permissions."
// @Failure         500 {object}  response.Error   "Internal Server Error - An error occurred while processing the request."
// @Router          /apps/{app_id}/connections/{connection_id}/facts [get]
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
		r.logger.With(ctx).Warnf("error retrieving total facts: %s", err.Error())
		return c.JSON(response.DefaultInternalError(c, r.logger, err.Error()))
	}
	pages := pagination.NewFromRequest(c.Request(), count)
	facts, err := r.service.Query(ctx, cid, sid, fid, pages.Offset(), pages.Limit())
	if err != nil {
		r.logger.With(ctx).Warnf("error retrieving the list of factss: %s", err.Error())
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
// @Summary         Issue a New Fact
// @Description     Issues a new fact to a specific connection, identified by the app_id and connection_id, using the details provided in the request body.
// @Tags            Facts
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           app_id        path string true "The unique identifier (ID) of the application associated with the connection."
// @Param           connection_id path string true "The unique identifier (ID) of the connection to which the fact will be issued."
// @Param           request       body CreateFactRequestDoc true "The body of the request, containing the details of the fact to be issued."
// @Success         201 {string}  string "Fact successfully issued."
// @Failure         400 {object}  response.Error "Bad Request - The body of the request is not valid or incorrectly formatted."
// @Failure         404 {object}  response.Error "Not Found - The requested resource does not exist, or the authenticated user does not have the necessary permissions."
// @Failure         500 {object}  response.Error "Internal Server Error - An error occurred while processing the request."
// @Router          /apps/{app_id}/connections/{connection_id}/facts [post]
func (r resource) create(c echo.Context) error {
	ctx := c.Request().Context()

	var input CreateFactRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(ctx).Warnf("invalid input for creating a fact: %s", err.Error())
		return c.JSON(response.DefaultBadRequestError())
	}

	if err := input.Validate(); err != nil {
		r.logger.With(ctx).Warnf("error validating input for creating a fact: %s", err.Error)
		return c.JSON(err.Status, err)
	}

	// Get the connection id
	conn, err := r.cService.Get(ctx, c.Param("app_id"), c.Param("connection_id"))
	if err != nil {
		r.logger.With(ctx).Warnf("connection not found: %s", err.Error())
		return c.JSON(response.DefaultNotFoundError())
	}

	err = r.service.Create(ctx, c.Param("app_id"), c.Param("connection_id"), conn.ID, input)
	if err != nil {
		r.logger.With(ctx).Warnf("error creating a fact: %s", err.Error())
		return c.JSON(response.DefaultInternalError(c, r.logger, err.Error()))
	}

	return c.NoContent(http.StatusAccepted)
}

// DeleteFact godoc
// @Summary         Delete a Fact
// @Description     Deletes an existing fact for a specific connection, identified by the app_id, connection_id, and the fact's id. The fact is permanently removed from the system.
// @Tags            Facts
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           app_id        path string true "The unique identifier (ID) of the application associated with the connection."
// @Param           connection_id path string true "The unique identifier (ID) of the connection associated with the fact."
// @Param           id            path string true "The unique identifier (ID) of the fact to be deleted."
// @Success         204 {string}  string "No Content - The fact was successfully deleted."
// @Failure         404 {object}  response.Error "Not Found - The requested resource does not exist, or the authenticated user does not have the necessary permissions."
// @Router          /apps/{app_id}/connections/{connection_id}/facts/{id} [delete]
func (r resource) delete(c echo.Context) error {
	ctx := c.Request().Context()
	conn, err := r.cService.Get(ctx, c.Param("app_id"), c.Param("connection_id"))
	if err != nil {
		r.logger.With(ctx).Warnf("error retrieving connection: %s", err.Error())
		return c.JSON(response.DefaultNotFoundError())
	}

	err = r.service.Delete(c.Request().Context(), conn.ID, c.Param("id"))
	if err != nil {
		r.logger.With(ctx).Warnf("error deleting fact: %s", err.Error())
		return c.JSON(response.DefaultNotFoundError())
	}

	return c.NoContent(http.StatusNoContent)
}
