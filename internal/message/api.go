package message

import (
	"net/http"

	routing "github.com/go-ozzo/ozzo-routing/v2"
	"github.com/qiangxue/go-rest-api/internal/errors"
	"github.com/qiangxue/go-rest-api/pkg/log"
	"github.com/qiangxue/go-rest-api/pkg/pagination"
)

// RegisterHandlers sets up the routing of the HTTP handlers.
func RegisterHandlers(r *routing.RouteGroup, service Service, authHandler routing.Handler, logger log.Logger) {
	res := resource{service, logger}

	r.Get("/connections/<connection_id>/messages/<id>", res.get)
	r.Get("/connections/<connection_id>/messages", res.query)

	r.Use(authHandler)

	// the following endpoints require a valid JWT
	r.Post("/connections/<connection_id>/messages", res.create)
	r.Put("/connections/<connection_id>/messages/<id>", res.update)
	r.Delete("/connections/<connection_id>/messages/<id>", res.delete)
}

type resource struct {
	service Service
	logger  log.Logger
}

func (r resource) get(c *routing.Context) error {
	message, err := r.service.Get(c.Request.Context(), c.Param("id"))
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
	pages := pagination.NewFromRequest(c.Request, count)
	messages, err := r.service.Query(ctx, c.Param("connection_id"), pages.Offset(), pages.Limit())
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

	message, err := r.service.Update(c.Request.Context(), c.Param("id"), input)
	if err != nil {
		return err
	}

	return c.Write(message)
}

func (r resource) delete(c *routing.Context) error {
	message, err := r.service.Delete(c.Request.Context(), c.Param("id"))
	if err != nil {
		return err
	}

	return c.Write(message)
}
