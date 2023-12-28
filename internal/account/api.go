package account

import (
	"net/http"

	"github.com/joinself/restful-client/internal/auth"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/labstack/echo/v4"
)

// RegisterHandlers sets up the routing of the HTTP handlers.
func RegisterHandlers(r *echo.Group, service Service, authHandler echo.MiddlewareFunc, logger log.Logger) {
	res := resource{service, logger}

	// the following endpoints require a valid JWT
	r.Use(authHandler)

	r.POST("/accounts", res.create)
	r.DELETE("/accounts/:username", res.delete)
}

type resource struct {
	service Service
	logger  log.Logger
}

// CreateAccount godoc
// @Summary         Creates a new account.
// @Description  	Creates a new account and sends a request for public information. You must be authenticated as an admin.
// @Tags            accounts
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           request body CreateAccountRequest true "query params"
// @Success         200  {object}  account.Account
// @Router          /accounts [post]
func (r resource) create(c echo.Context) error {
	user := auth.CurrentUser(c)
	if user == nil || !user.IsAdmin() {
		return c.JSON(http.StatusNotFound, "not found")
	}

	var input CreateAccountRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return c.JSON(http.StatusBadRequest, "")
	}

	account, err := r.service.Create(c.Request().Context(), input)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, account)
}

// CreateAccount godoc
// @Summary         Deletes an existing account.
// @Description  	Deletes an existing account and sends a request for public information and avoids incoming comms from that account. You must be authenticated as an admin.
// @Tags            accounts
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           username   path      int  true  "current account username"
// @Param           request body CreateAccountRequest true "query params"
// @Success         200  {object}  account.Account
// @Router          /accounts/{username} [delete]
func (r resource) delete(c echo.Context) error {
	user := auth.CurrentUser(c)
	if user == nil || !user.IsAdmin() {
		return c.JSON(http.StatusNotFound, "not found")
	}

	err := r.service.Delete(c.Request().Context(), c.Param("username"))
	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, "success")
}
