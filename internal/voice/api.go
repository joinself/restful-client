package voice

import (
	"net/http"

	"github.com/joinself/restful-client/internal/connection"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/response"
	"github.com/labstack/echo/v4"
)

// RegisterHandlers sets up the routing of the HTTP handlers.
func RegisterHandlers(r *echo.Group, service Service, cService connection.Service, logger log.Logger) {
	res := resource{service, cService, logger}

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

func (r resource) stop(c echo.Context) error {
	// Get the connection id
	_, err := r.cService.Get(c.Request().Context(), c.Param("app_id"), c.Param("connection_id"))
	if err != nil {
		return c.JSON(response.DefaultNotFoundError())
	}

	r.service.Stop(c.Request().Context(), c.Param("app_id"), c.Param("connection_id"), c.Param("id"))

	return c.JSON(http.StatusOK, c.Param("id"))
}

func (r resource) busy(c echo.Context) error {
	// Get the connection id
	_, err := r.cService.Get(c.Request().Context(), c.Param("app_id"), c.Param("connection_id"))
	if err != nil {
		return c.JSON(response.DefaultNotFoundError())
	}

	r.service.Busy(c.Request().Context(), c.Param("app_id"), c.Param("connection_id"), c.Param("id"))

	return c.JSON(http.StatusOK, c.Param("id"))
}
