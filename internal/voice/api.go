package voice

import (
	"net/http"
	"strconv"

	"github.com/joinself/restful-client/internal/connection"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/pagination"
	"github.com/joinself/restful-client/pkg/response"
	"github.com/labstack/echo/v4"
)

var (
	// LastCall specifies the call id from what you want to get new calls.
	LastCall = "calls_since"
)

// RegisterHandlers sets up the routing of the HTTP handlers.
func RegisterHandlers(r *echo.Group, service Service, cService connection.Service, logger log.Logger) {
	res := resource{service, cService, logger}

	r.GET("/:app_id/connections/:connection_id/calls", res.query)
	r.GET("/:app_id/connections/:connection_id/calls/:id", res.get)
	r.POST("/:app_id/connections/:connection_id/calls", res.setup)
	r.POST("/:app_id/connections/:connection_id/calls/:id/start", res.start)
	r.POST("/:app_id/connections/:connection_id/calls/:id/stop", res.stop)
	r.POST("/:app_id/connections/:connection_id/calls/:id/accept", res.accept)
	r.POST("/:app_id/connections/:connection_id/calls/:id/busy", res.busy)
}

type resource struct {
	service  Service
	cService connection.Service
	logger   log.Logger
}

// GetMessages godoc
// @Summary Retrieves a paginated list of calls
// @Description Retrieves a list of calls based on the provided Application ID and Connection ID, starting from the call defined in the query parameter 'lastCall'. The list is paginated.
// @Tags calls
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param app_id path string true "Unique identifier of the Application"
// @Param connection_id path string true "Unique identifier of the Connection"
// @Param lastCall query int false "Unique identifier of the last Call from which to start the list. If not provided, the list starts from the most recent call."
// @Success 200 {object} pagination.PagedResult{items=Message} "Successfully retrieved the list of calls. The response contains the paginated list of calls."
// @Failure 404 {object} response.Error "The requested resource could not be found, or the request was unauthorized. Check the Application ID and Connection ID."
// @Router /apps/{app_id}/connections/{connection_id}/calls [get]
func (r resource) query(c echo.Context) error {
	ctx := c.Request().Context()
	aID := c.Param("app_id")
	cID := c.Param("connection_id")

	// Get the connection id
	_, err := r.cService.Get(ctx, aID, cID)
	if err != nil {
		return c.JSON(response.DefaultNotFoundError())
	}

	callsSince, err := strconv.Atoi(c.Request().URL.Query().Get(LastCall))
	if err != nil {
		callsSince = 0
	}

	// Get the total of entries.
	count, err := r.service.Count(ctx, aID, cID, callsSince)
	if err != nil {
		return c.JSON(response.DefaultInternalError(c, r.logger, err.Error()))
	}

	// Get the messages
	pages := pagination.NewFromRequest(c.Request(), count)
	messages, err := r.service.Query(ctx, aID, cID, callsSince, pages.Offset(), pages.Limit())
	if err != nil {
		return c.JSON(response.DefaultInternalError(c, r.logger, err.Error()))
	}

	pages.Items = messages
	return c.JSON(http.StatusOK, pages)
}

// GetMessage godoc
// @Summary Retrieve a specific call's details
// @Description Retrieves detailed information about a specific call based on the provided Application ID, Connection ID, and Call ID.
// @Tags calls
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param app_id path string true "Unique identifier of the Application"
// @Param connection_id path string true "Unique identifier of the Connection"
// @Param id path string true "Unique identifier of the Call"
// @Success 200 {object} Message "Successfully retrieved the call details. The response contains the call's details."
// @Failure 404 {object} response.Error "The requested resource could not be found, or the request was unauthorized. Check the Application ID, Connection ID, and Call ID."
// @Router /apps/{app_id}/connections/{connection_id}/calls/{id} [get]
func (r resource) get(c echo.Context) error {
	ctx := c.Request().Context()
	aID := c.Param("app_id")
	cID := c.Param("connection_id")

	_, err := r.cService.Get(ctx, aID, cID)
	if err != nil {
		println("connection does not exist")
		return c.JSON(response.DefaultNotFoundError())
	}

	call, err := r.service.Get(ctx, aID, cID, c.Param("id"))
	if err != nil {
		println("call does not exist")
		println(c.Param("id"))
		return c.JSON(response.DefaultNotFoundError())
	}

	return c.JSON(http.StatusOK, call)
}

// SetupCall godoc
// @Summary Setup a new call
// @Description Sends a setup request to the specific connection to initiate a call.
// @Tags calls
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param app_id path string true "Identifier of the application"
// @Param connection_id path string true "Identifier of the connection"
// @Param id path string true "Identifier of the call"
// @Success 200 {object} Message "Successfully initiated the call. The response body contains the details of the call."
// @Failure 400 {object} response.Error "Invalid request. The request body or the params contains an error."
// @Failure 404 {object} response.Error "Resource not found. The requested application, connection, or call does not exist."
// @Failure 500 {object} response.Error "Internal server error. An unexpected error occurred on the server."
// @Router /apps/{app_id}/connections/{connection_id}/calls/{id}/setup [post]
func (r resource) setup(c echo.Context) error {
	var input SetupData
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return c.JSON(response.DefaultBadRequestError())
	}

	if reqErr := input.Validate(); reqErr != nil {
		return c.JSON(reqErr.Status, reqErr)
	}

	// Get the connection id
	_, err := r.cService.Get(c.Request().Context(), c.Param("app_id"), c.Param("connection_id"))
	if err != nil {
		return c.JSON(response.DefaultNotFoundError())
	}

	call, err := r.service.Setup(c.Request().Context(), c.Param("app_id"), c.Param("connection_id"), input.Name)
	if err != nil {
		return c.JSON(response.DefaultInternalError(c, r.logger, err.Error()))
	}

	return c.JSON(http.StatusOK, call)
}

// StartCall godoc
// @Summary Start a new call
// @Description Sends a start request to the specific connection to initiate a call.
// @Tags calls
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param app_id path string true "Identifier of the application"
// @Param connection_id path string true "Identifier of the connection"
// @Param id path string true "Identifier of the call"
// @Success 200 {object} Message "Successfully initiated the call. The response body contains the details of the call."
// @Failure 400 {object} response.Error "Invalid request. The request body or the params contains an error."
// @Failure 404 {object} response.Error "Resource not found. The requested application, connection, or call does not exist."
// @Failure 500 {object} response.Error "Internal server error. An unexpected error occurred on the server."
// @Router /apps/{app_id}/connections/{connection_id}/calls/{id}/start [post]
func (r resource) start(c echo.Context) error {
	var input ProceedData
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return c.JSON(response.DefaultBadRequestError())
	}

	if reqErr := input.Validate(); reqErr != nil {
		return c.JSON(reqErr.Status, reqErr)
	}

	// Get the connection id
	_, err := r.cService.Get(c.Request().Context(), c.Param("app_id"), c.Param("connection_id"))
	if err != nil {
		return c.JSON(response.DefaultNotFoundError())
	}

	r.service.Start(c.Request().Context(), c.Param("app_id"), c.Param("connection_id"), c.Param("id"), input)

	return c.JSON(http.StatusOK, ``)
}

// AcceptCall godoc
// @Summary Accept a new call
// @Description Sends an accept request to the specific connection to initiate a call.
// @Tags calls
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param app_id path string true "Identifier of the application"
// @Param connection_id path string true "Identifier of the connection"
// @Param id path string true "Identifier of the call"
// @Success 200 {object} Message "Successfully initiated the call. The response body contains the details of the call."
// @Failure 400 {object} response.Error "Invalid request. The request body or the params contains an error."
// @Failure 404 {object} response.Error "Resource not found. The requested application, connection, or call does not exist."
// @Failure 500 {object} response.Error "Internal server error. An unexpected error occurred on the server."
// @Router /apps/{app_id}/connections/{connection_id}/calls/{id}/accept [post]
func (r resource) accept(c echo.Context) error {
	var input ProceedData
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return c.JSON(response.DefaultBadRequestError())
	}

	if reqErr := input.Validate(); reqErr != nil {
		return c.JSON(reqErr.Status, reqErr)
	}

	// Get the connection id
	_, err := r.cService.Get(c.Request().Context(), c.Param("app_id"), c.Param("connection_id"))
	if err != nil {
		return c.JSON(response.DefaultNotFoundError())
	}

	r.service.Accept(c.Request().Context(), c.Param("app_id"), c.Param("connection_id"), c.Param("id"), input)

	return c.JSON(http.StatusOK, c.Param("id"))
}

// StopCall godoc
// @Summary Stop a new call
// @Description Sends a stop request to the specific connection to initiate a call.
// @Tags calls
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param app_id path string true "Identifier of the application"
// @Param connection_id path string true "Identifier of the connection"
// @Param id path string true "Identifier of the call"
// @Success 200 {object} Message "Successfully initiated the call. The response body contains the details of the call."
// @Failure 400 {object} response.Error "Invalid request. The request body or the params contains an error."
// @Failure 404 {object} response.Error "Resource not found. The requested application, connection, or call does not exist."
// @Failure 500 {object} response.Error "Internal server error. An unexpected error occurred on the server."
// @Router /apps/{app_id}/connections/{connection_id}/calls/{id}/stop [post]
func (r resource) stop(c echo.Context) error {
	// Get the connection id
	_, err := r.cService.Get(c.Request().Context(), c.Param("app_id"), c.Param("connection_id"))
	if err != nil {
		return c.JSON(response.DefaultNotFoundError())
	}

	r.service.Stop(c.Request().Context(), c.Param("app_id"), c.Param("connection_id"), c.Param("id"))

	return c.JSON(http.StatusOK, c.Param("id"))
}

// BusyCall godoc
// @Summary Busy a new call
// @Description Sends a busy request to the specific connection to initiate a call.
// @Tags calls
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param app_id path string true "Identifier of the application"
// @Param connection_id path string true "Identifier of the connection"
// @Param id path string true "Identifier of the call"
// @Success 200 {object} Message "Successfully initiated the call. The response body contains the details of the call."
// @Failure 400 {object} response.Error "Invalid request. The request body or the params contains an error."
// @Failure 404 {object} response.Error "Resource not found. The requested application, connection, or call does not exist."
// @Failure 500 {object} response.Error "Internal server error. An unexpected error occurred on the server."
// @Router /apps/{app_id}/connections/{connection_id}/calls/{id}/busy [post]
func (r resource) busy(c echo.Context) error {
	// Get the connection id
	_, err := r.cService.Get(c.Request().Context(), c.Param("app_id"), c.Param("connection_id"))
	if err != nil {
		return c.JSON(response.DefaultNotFoundError())
	}

	r.service.Busy(c.Request().Context(), c.Param("app_id"), c.Param("connection_id"), c.Param("id"))

	return c.JSON(http.StatusOK, c.Param("id"))
}
