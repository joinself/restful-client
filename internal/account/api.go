package account

import (
	"net/http"

	"github.com/joinself/restful-client/pkg/acl"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/response"
	"github.com/labstack/echo/v4"
)

// RegisterHandlers sets up the routing of the HTTP handlers.
func RegisterHandlers(r *echo.Group, service Service, logger log.Logger) {
	res := resource{service, logger}

	r.POST("", res.create)
	r.DELETE("/:username", res.delete)
	r.PUT("/:username/password", res.changePassword)
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
// @Success         200  {object}  CreateAccountResponse
// @Router          /accounts [post]
func (r resource) create(c echo.Context) error {
	user := acl.CurrentUser(c)
	if user == nil || !user.IsAdmin() {
		return c.JSON(response.DefaultNotFoundError())
	}

	var input CreateAccountRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return c.JSON(response.DefaultBadRequestError())
	}

	if reqErr := input.Validate(); reqErr != nil {
		r.logger.With(c.Request().Context()).Info(reqErr)
		return c.JSON(reqErr.Status, reqErr)
	}

	account, err := r.service.Create(c.Request().Context(), input)
	if err != nil {
		return c.JSON(response.DefaultInternalError(c, r.logger, err.Error()))
	}

	return c.JSON(http.StatusCreated, CreateAccountResponse{
		UserName:               account.UserName,
		Resources:              account.Resources,
		RequiresPasswordChange: account.RequiresPasswordChange,
	})
}

// DeleteAccount godoc
// @Summary         Deletes an existing account.
// @Description     Deletes an existing account and sends a request for public information and avoids incoming comms from that account. You must be authenticated as an admin.
// @Tags            accounts
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           username   path      string  true  "Username of the account to delete"
// @Success         204  {string} string  "No Content"
// @Failure         404 {object} response.Error "Not found - The requested resource does not exist, or you don't have permissions to access it"
// @Router          /accounts/{username} [delete]
func (r resource) delete(c echo.Context) error {
	user := acl.CurrentUser(c)
	if user == nil || !user.IsAdmin() {
		return c.JSON(response.DefaultNotFoundError())
	}

	err := r.service.Delete(c.Request().Context(), c.Param("username"))
	if err != nil {
		return c.JSON(response.DefaultNotFoundError())
	}

	return c.NoContent(http.StatusNoContent)
}

// ChangePassword godoc
// @Summary         Changes the password for the current user.
// @Description     Changes the password for the current user. You must be authenticated.
// @Tags            accounts
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           username   path      string  true  "Username of the account to change the password"
// @Param           request body ChangePasswordRequest true "Password change details"
// @Success         200  {string} string "No Content"
// @Failure         400 {object} response.Error "Bad request - The provided body is not valid"
// @Failure         404 {object} response.Error "Not found - The requested resource does not exist, or you don't have permissions to access it"
// @Failure         500 {object} response.Error "Internal error - There was a problem with your request"
// @Router          /accounts/{username}/password [put]
func (r resource) changePassword(c echo.Context) error {
	ctx := c.Request().Context()
	user := acl.CurrentUser(c)
	if user == nil {
		return c.JSON(response.DefaultNotFoundError())
	}

	if user.GetName() != c.Param("username") {
		r.logger.With(ctx).Info("update username not matching")
		return c.JSON(response.DefaultNotFoundError())
	}

	var i ChangePasswordRequest
	if err := c.Bind(&i); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return c.JSON(response.DefaultBadRequestError())
	}

	if reqErr := i.Validate(); reqErr != nil {
		r.logger.With(ctx).Info(reqErr.Error)
		return c.JSON(reqErr.Status, reqErr)
	}

	err := r.service.SetPassword(ctx, c.Param("username"), i.Password, i.NewPassword)
	if err != nil {
		return c.JSON(response.DefaultInternalError(c, r.logger, err.Error()))
	}

	return c.NoContent(http.StatusOK)
}
