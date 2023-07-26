package group

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/joinself/restful-client/internal/connection"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/pagination"
	"github.com/labstack/echo/v4"
)

type resource struct {
	service  Service
	cService connection.Service
	logger   log.Logger
}

func RegisterHandlers(r *echo.Group, s Service, cs connection.Service, am echo.MiddlewareFunc, logger log.Logger) {
	res := resource{
		service:  s,
		cService: cs,
		logger:   logger,
	}

	r.Use(am)

	r.GET("/apps/:app_id/groups", res.query)
	// Get group details
	r.POST("/apps/:app_id/groups", res.create)
	// Create a group
	r.GET("/apps/:app_id/groups/:id", res.get)
	// Join a group / change details
	r.PUT("/apps/:app_id/groups/:id", res.update)
	// Leave a group
	r.DELETE("/apps/:app_id/groups/:id", res.delete)
}

// ListMessages    godoc
// @Summary        List app groups.
// @Description    List all app groups.
// @Tags           groups
// @Accept         json
// @Produce        json
// @Security       BearerAuth
// @Param          page query int false "page number"
// @Param          per_page query int false "number of elements per page"
// @Param          app_id   path      string  true  "App id"
// @Success        200  {object}  response
// @Router         /apps/:app_id/groups [get]
func (r resource) query(c echo.Context) error {
	ctx := c.Request().Context()
	count, err := r.service.Count(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	pages := pagination.NewFromRequest(c.Request(), count)
	messages, err := r.service.Query(ctx, c.Param("app_id"), pages.Offset(), pages.Limit())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	pages.Items = messages
	return c.JSON(http.StatusOK, pages)
}

// SendMessage      godoc
// @Summary         Create a group
// @Description  	Creates a new group
// @Tags            groups
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           app_id   path      string  true  "App id"
// @Param           request body CreateGroupRequest true "message request"
// @Success         200  {object}  Message
// @Router          /apps/:app_id/groups [post]
func (r resource) create(c echo.Context) error {
	var input CreateGroupRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return c.JSON(http.StatusBadRequest, "invalid input")
	}

	group, err := r.service.Create(c.Request().Context(), c.Param("app_id"), input)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	// TODO: Send an invitation.

	return c.JSON(http.StatusCreated, group)
}

// GetConnection godoc
// @Summary      Get group details.
// @Description  Get group details by group ID..
// @Tags         groups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        app_id   path      string  true  "App id"
// @Param        id   path      int  true  "current group id"
// @Success      200  {object}  Group
// @Router       /apps/:app_id/groups/{id} [get]
func (r resource) get(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, errors.New("invalid group id"))
	}

	group, err := r.service.Get(c.Request().Context(), c.Param("app_id"), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, group)
}

// SendMessage      godoc
// @Summary         Edits a group.
// @Description  	Updates a group
// @Tags            groups
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           app_id   path      string  true  "App id"
// @Param           group_id   path      int  true  "Group id"
// @Param           request body UpdateGroupRequest true "message request"
// @Success         200  {object}  Message
// @Router          /apps/:app_id/groups/{group_id} [put]
func (r resource) update(c echo.Context) error {
	var input UpdateGroupRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return c.JSON(http.StatusBadRequest, "invalid input")
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, errors.New("invalid group id"))
	}

	// TODO: Update the group details
	group, err := r.service.Update(c.Request().Context(), c.Param("app_id"), id, input)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	return c.JSON(http.StatusOK, group)
}

// CreateConnection godoc
// @Summary         Deletes a group.
// @Description  	Deletes a group from the database and sends a leave group request.
// @Tags            groups
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           app_id   path      string  true  "App id"
// @Param           id   path      int  true  "group id to be removed"
// @Success         200  {object}  connection.Connection
// @Router          /apps/:app_id/groups/{id} [delete]
func (r resource) delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	err = r.service.Delete(c.Request().Context(), c.Param("app_id"), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, "")
}
