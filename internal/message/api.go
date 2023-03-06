package message

import (
	"net/http"
	"strconv"

	routing "github.com/go-ozzo/ozzo-routing/v2"
	"github.com/joinself/restful-client/internal/errors"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/pagination"
)

// RegisterHandlers sets up the routing of the HTTP handlers.
func RegisterHandlers(r *routing.RouteGroup, service Service, authHandler routing.Handler, logger log.Logger) {
	res := resource{service, logger}

	// the following endpoints require a valid JWT
	r.Use(authHandler)
	r.Get("/connections/<connection_id>/messages/<id>", res.get)
	r.Get("/connections/<connection_id>/messages", res.query)
	r.Post("/connections/<connection_id>/messages", res.create)
	r.Put("/connections/<connection_id>/messages/<id>", res.update)
	r.Delete("/connections/<connection_id>/messages/<id>", res.delete)
}

var (
	// LastMessage specifies the message id from what you want to get new messages.
	LastMessage = "last_message_id"
)

type resource struct {
	service Service
	logger  log.Logger
}

func (r resource) get(c *routing.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}

	message, err := r.service.Get(c.Request.Context(), id)
	if err != nil {
		return err
	}

	return c.Write(message)
}

func (r resource) query(c *routing.Context) error {
	ctx := c.Request.Context()
	count, err := r.service.Count(ctx)
	if err != nil {
		return err
	}

	messagesSince, err := strconv.Atoi(c.Request.URL.Query().Get(LastMessage))
	if err != nil {
		messagesSince = 0
	}

	pages := pagination.NewFromRequest(c.Request, count)
	messages, err := r.service.Query(ctx, c.Param("connection_id"), messagesSince, pages.Offset(), pages.Limit())
	if err != nil {
		return err
	}

	pages.Items = messages
	return c.Write(pages)
}

func (r resource) create(c *routing.Context) error {
	var input CreateMessageRequest
	if err := c.Read(&input); err != nil {
		r.logger.With(c.Request.Context()).Info(err)
		return errors.BadRequest("")
	}
	message, err := r.service.Create(c.Request.Context(), c.Param("connection_id"), input)
	if err != nil {
		return err
	}

	return c.WriteWithStatus(message, http.StatusCreated)
}

func (r resource) update(c *routing.Context) error {
	var input UpdateMessageRequest
	if err := c.Read(&input); err != nil {
		r.logger.With(c.Request.Context()).Info(err)
		return errors.BadRequest("")
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}
	message, err := r.service.Update(c.Request.Context(), id, input)
	if err != nil {
		return err
	}

	return c.Write(message)
}

func (r resource) delete(c *routing.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}

	message, err := r.service.Delete(c.Request.Context(), id)
	if err != nil {
		return err
	}

	return c.Write(message)
}
