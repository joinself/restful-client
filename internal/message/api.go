package message

import (
	"net/http"
	"strconv"

	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/pagination"
	"github.com/labstack/echo/v4"
)

// RegisterHandlers sets up the routing of the HTTP handlers.
func RegisterHandlers(r *echo.Group, service Service, authHandler echo.MiddlewareFunc, logger log.Logger) {
	res := resource{service, logger}

	// the following endpoints require a valid JWT
	r.Use(authHandler)
	r.GET("/connections/:connection_id/messages/:id", res.get)
	r.GET("/connections/:connection_id/messages", res.query)
	r.POST("/connections/:connection_id/messages", res.create)
	r.PUT("/connections/:connection_id/messages/:id", res.update)
	r.DELETE("/connections/:connection_id/messages/:id", res.delete)
}

var (
	// LastMessage specifies the message id from what you want to get new messages.
	LastMessage = "last_message_id"
)

type resource struct {
	service Service
	logger  log.Logger
}

// GetMessage    godoc
// @Summary      Gets a message.
// @Description  Get message details
// @Tags         messages
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        connection_id   path      int  true  "Connection id"
// @Param        id   path      int  true  "Message id"
// @Success      200  {object}  Message
// @Router       /connections/{connection_id}/messages/{id} [get]
func (r resource) get(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	message, err := r.service.Get(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, message)
}

// ListMessages    godoc
// @Summary        List conversation messages.
// @Description    List conversation messages with a specific connection.
// @Tags           messages
// @Accept         json
// @Produce        json
// @Security       BearerAuth
// @Param          page query int false "page number"
// @Param          per_page query int false "number of elements per page"
// @Param          connection_id path string  true  "Connection ID"
// @Success        200  {array}  connection.Connection
// @Router         /connections/{connection_id}/messages [get]
func (r resource) query(c echo.Context) error {
	ctx := c.Request().Context()
	count, err := r.service.Count(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	messagesSince, err := strconv.Atoi(c.Request().URL.Query().Get(LastMessage))
	if err != nil {
		messagesSince = 0
	}

	pages := pagination.NewFromRequest(c.Request(), count)
	messages, err := r.service.Query(ctx, c.Param("connection_id"), messagesSince, pages.Offset(), pages.Limit())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	pages.Items = messages
	return c.JSON(http.StatusOK, pages)
}

// SendMessage      godoc
// @Summary         Sends a message.
// @Description  	Sends a message to the specified connection.
// @Tags            messages
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           connection_id   path      int  true  "Connection id"
// @Param           request body CreateMessageRequest true "message request"
// @Success         200  {object}  Message
// @Router          /connections/{connection_id}/messages [post]
func (r resource) create(c echo.Context) error {
	var input CreateMessageRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return c.JSON(http.StatusBadRequest, "")
	}
	message, err := r.service.Create(c.Request().Context(), c.Param("connection_id"), input)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, message)
}

func (r resource) update(c echo.Context) error {
	var input UpdateMessageRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return c.JSON(http.StatusBadRequest, "")
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	message, err := r.service.Update(c.Request().Context(), id, input)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, message)
}

func (r resource) delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	message, err := r.service.Delete(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, message)
}
