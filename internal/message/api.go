package message

import (
	"net/http"
	"strconv"

	"github.com/joinself/restful-client/internal/connection"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/pagination"
	"github.com/joinself/restful-client/pkg/response"
	"github.com/labstack/echo/v4"
)

// RegisterHandlers sets up the routing of the HTTP handlers.
func RegisterHandlers(r *echo.Group, service Service, cService connection.Service, logger log.Logger) {
	res := resource{service, cService, logger}

	r.GET("/:app_id/connections/:connection_id/messages/:id", res.get)
	r.GET("/:app_id/connections/:connection_id/messages", res.query)
	r.POST("/:app_id/connections/:connection_id/messages", res.create)
	r.PUT("/:app_id/connections/:connection_id/messages/:id", res.update)
	r.DELETE("/:app_id/connections/:connection_id/messages/:id", res.delete)
	r.POST("/:app_id/connections/:connection_id/messages/:id/read", res.read)
	r.POST("/:app_id/connections/:connection_id/messages/:id/received", res.received)
}

var (
	// LastMessage specifies the message id from what you want to get new messages.
	LastMessage = "messages_since"
)

type resource struct {
	service  Service
	cService connection.Service
	logger   log.Logger
}

// GetMessage    godoc
// @Summary      Gets a message.
// @Description  Retrieves details of a specific message identified by its JTI, within the context of a specific app and connection. Requires Bearer authentication.
// @Tags         messages
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        app_id   path      string  true  "Application ID"
// @Param        connection_id   path      string  true  "Connection ID"
// @Param        id   path      string  true  "Message ID"
// @Success      200  {object}  Message "Successful retrieval of message details"
// @Failure      404  {object}  response.Error "Resource not found or unauthorized access"
// @Router       /apps/{app_id}/connections/{connection_id}/messages/{id} [get]
func (r resource) get(c echo.Context) error {
	conn, err := r.cService.Get(c.Request().Context(), c.Param("app_id"), c.Param("connection_id"))
	if err != nil {
		return c.JSON(response.DefaultNotFoundError())
	}

	message, err := r.service.Get(c.Request().Context(), conn.ID, c.Param("id"))
	if err != nil {
		return c.JSON(response.DefaultNotFoundError())
	}

	return c.JSON(http.StatusOK, message)
}

// ListMessages    godoc
// @Summary        List conversation messages.
// @Description    Retrieves all messages for a specific connection within an app. Supports pagination and can filter messages since a specific message ID.
// @Tags           messages
// @Accept         json
// @Produce        json
// @Security       BearerAuth
// @Param          messages_since query int false "Return elements since a message ID"
// @Param          page query int false "Page number for results pagination"
// @Param          per_page query int false "Number of results per page for pagination"
// @Param          app_id   path      string  true  "Application ID"
// @Param          connection_id path string  true  "Connection ID"
// @Success        200  {object}  ExtListResponse "Successfully retrieved list of messages"
// @Failure        404  {object}  response.Error "Resource not found or unauthorized access"
// @Failure        500  {object}  response.Error "Internal server error"
// @Router         /apps/{app_id}/connections/{connection_id}/messages [get]
func (r resource) query(c echo.Context) error {
	ctx := c.Request().Context()

	// Get the connection id
	conn, err := r.cService.Get(c.Request().Context(), c.Param("app_id"), c.Param("connection_id"))
	if err != nil {
		return c.JSON(response.DefaultNotFoundError())
	}

	messagesSince, err := strconv.Atoi(c.Request().URL.Query().Get(LastMessage))
	if err != nil {
		messagesSince = 0
	}

	// Get the total of entries.
	count, err := r.service.Count(ctx, conn.ID, messagesSince)
	if err != nil {
		return c.JSON(response.DefaultInternalError(c, r.logger, err.Error()))
	}

	// Get the messages
	pages := pagination.NewFromRequest(c.Request(), count)
	messages, err := r.service.Query(ctx, conn.ID, messagesSince, pages.Offset(), pages.Limit())
	if err != nil {
		return c.JSON(response.DefaultInternalError(c, r.logger, err.Error()))
	}

	pages.Items = messages
	return c.JSON(http.StatusOK, pages)
}

// SendMessage    godoc
// @Summary       Sends a message.
// @Description   Sends a message to a specific connection within an app. Requires Bearer authentication.
// @Tags          messages
// @Accept        json
// @Produce       json
// @Security      BearerAuth
// @Param         app_id   path      string  true  "Application ID"
// @Param         connection_id   path      string  true  "Connection ID"
// @Param         request body CreateMessageRequest true "Request to create a message"
// @Success       202  {object}  Message "Successfully sent message"
// @Failure       400  {object}  response.Error "Invalid input"
// @Failure       404  {object}  response.Error "Resource not found or unauthorized access"
// @Failure       500  {object}  response.Error "Internal server error"
// @Router        /apps/{app_id}/connections/{connection_id}/messages [post]
func (r resource) create(c echo.Context) error {
	var input CreateMessageRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return c.JSON(response.DefaultBadRequestError())
	}

	if reqErr := input.Validate(); reqErr != nil {
		return c.JSON(reqErr.Status, reqErr)
	}

	// Get the connection id
	connection, err := r.cService.Get(c.Request().Context(), c.Param("app_id"), c.Param("connection_id"))
	if err != nil {
		return c.JSON(response.DefaultNotFoundError())
	}

	// Create the message
	message, err := r.service.Create(c.Request().Context(), c.Param("app_id"), c.Param("connection_id"), connection.ID, input)
	if err != nil {
		return c.JSON(response.DefaultInternalError(c, r.logger, err.Error()))
	}

	return c.JSON(http.StatusAccepted, message)
}

// EditMessage    godoc
// @Summary       Edits a message.
// @Description   Updates an existing message in a specific connection within an app. Requires Bearer authentication.
// @Tags          messages
// @Accept        json
// @Produce       json
// @Security      BearerAuth
// @Param         app_id   path      string  true  "Application ID"
// @Param         connection_id   path      string  true  "Connection ID"
// @Param         id   path      string  true  "Message ID"
// @Param         request body UpdateMessageRequest true "Request to update a message"
// @Success       200  {object}  Message "Successfully updated message"
// @Failure       400  {object}  response.Error "Invalid input"
// @Failure       404  {object}  response.Error "Resource not found or unauthorized access"
// @Failure       500  {object}  response.Error "Internal server error"
// @Router        /apps/{app_id}/connections/{connection_id}/messages/{id} [put]
func (r resource) update(c echo.Context) error {
	var input UpdateMessageRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return c.JSON(response.DefaultBadRequestError())
	}

	if reqErr := input.Validate(); reqErr != nil {
		return c.JSON(reqErr.Status, reqErr)
	}

	// Get the connection id
	connection, err := r.cService.Get(c.Request().Context(), c.Param("app_id"), c.Param("connection_id"))
	if err != nil {
		return c.JSON(response.DefaultNotFoundError())
	}

	message, err := r.service.Update(
		c.Request().Context(),
		c.Param("app_id"),
		connection.ID,
		c.Param("connection_id"),
		c.Param("id"),
		input)
	if err != nil {
		return c.JSON(response.DefaultInternalError(c, r.logger, err.Error()))
	}

	return c.JSON(http.StatusOK, message)
}

// DeleteMessage    godoc
// @Summary         Deletes a message.
// @Description     Deletes a specific message from a specific connection within an app.
// @Tags            messages
// @Security        BearerAuth
// @Param           app_id   path   string  true  "Application ID"
// @Param           connection_id   path   string  true  "Connection ID"
// @Param           id   path   string  true  "Message ID"
// @Success         204  {object}  nil "Successfully deleted message, no content returned"
// @Failure         404  {object}  response.Error "Resource not found or unauthorized access"
// @Router          /apps/{app_id}/connections/{connection_id}/messages/{id} [delete]
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

// MarkMessageAsRead    godoc
// @Summary         Marks a specific message as read
// @Description     This endpoint allows you to mark a specific message as read in a given application by its connection ID and message ID.
// @Tags            messages
// @Security        BearerAuth
// @Accept          json
// @Produce         json
// @Param           app_id   path   string  true  "Unique identifier of the application"
// @Param           connection_id   path   string  true  "Unique identifier of the connection"
// @Param           id   path   string  true  "Unique identifier of the message to be marked as read"
// @Success         204  {object}  nil "Successfully marked the message as read. No content is returned."
// @Failure         404  {object}  response.Error "The requested resource could not be found, or you're not authorized to access it."
// @Failure         500  {object}  response.Error "An error occurred while processing your request. Please try again."
// @Router          /apps/{app_id}/connections/{connection_id}/messages/{id}/read [post]
func (r resource) read(c echo.Context) error {
	connection, err := r.cService.Get(c.Request().Context(), c.Param("app_id"), c.Param("connection_id"))
	if err != nil {
		return c.JSON(response.DefaultNotFoundError())
	}

	r.service.MarkAsRead(c.Request().Context(), c.Param("app_id"), c.Param("connection_id"), c.Param("id"), connection.ID)
	if err != nil {
		return c.JSON(response.DefaultInternalError(c, r.logger, err.Error()))
	}

	return c.NoContent(http.StatusOK)
}

// MarkMessageAsReceived godoc
// @Summary         Marks a message as received
// @Description     Updates the status of the specified message to 'received'.
// @Tags            messages
// @Security        BearerAuth
// @Param           app_id   path   string  true  "Application ID"
// @Param           connection_id   path   string  true  "Connection ID"
// @Param           id   path   string  true  "Message ID"
// @Success         200  {object}  nil "Successfully updated message status to received"
// @Failure         404  {object}  response.Error "Message not found or unauthorized access"
// @Failure         500  {object}  response.Error "Internal server error while processing your request"
// @Router          /apps/{app_id}/connections/{connection_id}/messages/{id}/received [post]
func (r resource) received(c echo.Context) error {
	connection, err := r.cService.Get(c.Request().Context(), c.Param("app_id"), c.Param("connection_id"))
	if err != nil {
		return c.JSON(response.DefaultNotFoundError())
	}

	err = r.service.MarkAsReceived(c.Request().Context(), c.Param("app_id"), c.Param("connection_id"), c.Param("id"), connection.ID)
	if err != nil {
		return c.JSON(response.DefaultInternalError(c, r.logger, err.Error()))
	}

	return c.NoContent(http.StatusOK)
}
