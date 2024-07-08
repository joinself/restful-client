package signature

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
	// LastSignature specifies the signature id from what you want to get new signatures.
	LastSignature = "signatures_since"
)

// RegisterHandlers sets up the routing of the HTTP handlers.
func RegisterHandlers(r *echo.Group, service Service, cService connection.Service, logger log.Logger) {
	res := resource{service, cService, logger}

	r.GET("/:app_id/connections/:connection_id/signatures", res.query)
	r.GET("/:app_id/connections/:connection_id/signatures/:id", res.get)
	r.POST("/:app_id/connections/:connection_id/signatures", res.create)
}

type resource struct {
	service  Service
	cService connection.Service
	logger   log.Logger
}

// GetSignatures godoc
// @Summary Retrieves a paginated list of sent document signatures
// @Description Retrieves a list of document signatures sent based on the provided Application ID and Connection ID, starting from the signature defined in the query parameter 'lastSignature'. The list is paginated.
// @Tags signatures
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param app_id path string true "Unique identifier of the Application"
// @Param connection_id path string true "Unique identifier of the Connection"
// @Param lastSignature query string false "Unique identifier of the last Signature from which to start the list. If not provided, the list starts from the most recent signature."
// @Success 200 {object} pagination.PagedResult{items=Signature} "Successfully retrieved the list of signatures. The response contains the paginated list of signatures."
// @Failure 404 {object} response.Error "The requested resource could not be found, or the request was unauthorized. Check the Application ID and Connection ID."
// @Router /apps/{app_id}/connections/{connection_id}/signatures [get]
func (r resource) query(c echo.Context) error {
	ctx := c.Request().Context()
	aID := c.Param("app_id")
	cID := c.Param("connection_id")

	// Get the connection id
	_, err := r.cService.Get(ctx, aID, cID)
	if err != nil {
		r.logger.With(ctx).Warnf("error retrieving connection: %s", err.Error())
		return c.JSON(response.DefaultNotFoundError())
	}

	signaturesSince, err := strconv.Atoi(c.Request().URL.Query().Get(LastSignature))
	if err != nil {
		signaturesSince = 0
	}

	// Get the total of entries.
	count, err := r.service.Count(ctx, aID, cID, signaturesSince)
	if err != nil {
		r.logger.With(ctx).Warnf("error total signatures: %s", err.Error())
		return c.JSON(response.DefaultInternalError(c, r.logger, err.Error()))
	}

	// Get the messages
	pages := pagination.NewFromRequest(c.Request(), count)
	messages, err := r.service.Query(ctx, aID, cID, signaturesSince, pages.Offset(), pages.Limit())
	if err != nil {
		r.logger.With(ctx).Warnf("error retrieving paginated list: %s", err.Error())
		return c.JSON(response.DefaultInternalError(c, r.logger, err.Error()))
	}

	pages.Items = messages
	return c.JSON(http.StatusOK, pages)
}

// GetSignature godoc
// @Summary Retrieve a specific document signature's details
// @Description Retrieves detailed information about a specific document signature based on the provided Application ID, Connection ID, and Signature ID.
// @Tags signatures
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param app_id path string true "Application ID"
// @Param connection_id path string true "Connection ID"
// @Param id path string true "Signature ID"
// @Success 200 {object} Signature "Successfully retrieved the signature. The response contains the signature's details."
// @Failure 404 {object} response.Error "The requested resource could not be found, or the request was unauthorized. Check the Application ID, Connection ID, and Signature ID."
// @Router /apps/{app_id}/connections/{connection_id}/signatures/{id} [get]
func (r resource) get(c echo.Context) error {
	ctx := c.Request().Context()
	aID := c.Param("app_id")
	cID := c.Param("connection_id")

	_, err := r.cService.Get(ctx, aID, cID)
	if err != nil {
		r.logger.With(ctx).Warnf("error retrieving connection: %s", err.Error())
		return c.JSON(response.DefaultNotFoundError())
	}

	signature, err := r.service.Get(ctx, aID, cID, c.Param("id"))
	if err != nil {
		r.logger.With(ctx).Warnf("error retrieving signature: %s", err.Error())
		return c.JSON(response.DefaultNotFoundError())
	}

	return c.JSON(http.StatusOK, signature)
}

// CreateSignature godoc
// @Summary Create a new document signature request
// @Description Create a new document signature request and send it to the specific connection to initiate the signature process
// @Tags signatures
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param app_id path string true "Application Identifier"
// @Param connection_id path string true "Connection Identifier"
// @Param id path string true "Signature Identifier"
// @Param body body CreateSignatureRequest true "Signature Request Body"
// @Success 200 {object} Signature "Successfully initiated the signature process. The response body contains the details of the signature."
// @Failure 400 {object} response.Error "Invalid request. The request body or the params contains an error."
// @Failure 404 {object} response.Error "Resource not found. The requested application, connection, or signature does not exist."
// @Failure 500 {object} response.Error "Internal server error. An unexpected error occurred on the server."
// @Router /apps/{app_id}/connections/{connection_id}/signatures/{id}/setup [post]
func (r resource) create(c echo.Context) error {
	ctx := c.Request().Context()

	var input CreateSignatureRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(ctx).Warnf("error processing input: %s", err.Error())
		return c.JSON(response.DefaultBadRequestError())
	}

	if reqErr := input.Validate(); reqErr != nil {
		return c.JSON(reqErr.Status, reqErr)
	}

	_, err := r.cService.Get(c.Request().Context(), c.Param("app_id"), c.Param("connection_id"))
	if err != nil {
		r.logger.With(ctx).Warnf("error retrieving connection: %s", err.Error())
		return c.JSON(response.DefaultNotFoundError())
	}

	signature, err := r.service.Create(c.Request().Context(), c.Param("app_id"), c.Param("connection_id"), input)
	if err != nil {
		r.logger.With(ctx).Warnf("error creating signature: %s", err.Error())
		return c.JSON(response.DefaultInternalError(c, r.logger, err.Error()))
	}

	return c.JSON(http.StatusOK, signature)
}
