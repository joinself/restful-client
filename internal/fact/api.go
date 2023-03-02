package fact

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

	r.Use(authHandler)

	r.Get("/connections/<connection_id>/facts/<id>", res.get)
	r.Get("/connections/<connection_id>/facts", res.query)

	// the following endpoints require a valid JWT
	r.Post("/connections/<connection_id>/facts", res.create)
	r.Put("/connections/<connection_id>/facts/<id>", res.update)
	r.Delete("/connections/<connection_id>/facts/<id>", res.delete)
}

type resource struct {
	service Service
	logger  log.Logger
}

func (r resource) get(c *routing.Context) error {
	fact, err := r.service.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		return err
	}

	return c.Write(fact)
}

func (r resource) query(c *routing.Context) error {
	ctx := c.Request.Context()
	count, err := r.service.Count(ctx)
	if err != nil {
		return err
	}
	pages := pagination.NewFromRequest(c.Request, count)
	facts, err := r.service.Query(ctx, c.Param("connection_id"), pages.Offset(), pages.Limit())
	if err != nil {
		return err
	}
	pages.Items = facts
	return c.Write(pages)
}

func (r resource) create(c *routing.Context) error {
	var input CreateFactRequest
	if err := c.Read(&input); err != nil {
		r.logger.With(c.Request.Context()).Info(err)
		return errors.BadRequest("")
	}
	fact, err := r.service.Create(c.Request.Context(), c.Param("connection_id"), input)
	if err != nil {
		return err
	}

	return c.WriteWithStatus(fact, http.StatusCreated)
}

func (r resource) update(c *routing.Context) error {
	var input UpdateFactRequest
	if err := c.Read(&input); err != nil {
		r.logger.With(c.Request.Context()).Info(err)
		return errors.BadRequest("")
	}

	fact, err := r.service.Update(c.Request.Context(), c.Param("id"), input)
	if err != nil {
		return err
	}

	return c.Write(fact)
}

func (r resource) delete(c *routing.Context) error {
	fact, err := r.service.Delete(c.Request.Context(), c.Param("id"))
	if err != nil {
		return err
	}

	return c.Write(fact)
}
