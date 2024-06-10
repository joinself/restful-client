package account

import (
	"net/http"

	"github.com/joinself/restful-client/pkg/acl"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/pagination"
	"github.com/joinself/restful-client/pkg/response"
	"github.com/labstack/echo/v4"
)

// RegisterHandlers sets up the routing of the HTTP handlers.
func RegisterHandlers(r *echo.Group, service Service, logger log.Logger) {
	res := resource{service, logger}

	r.GET("", res.list)
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
// @Description     This endpoint creates a new account in the system and sends a request to gather public information associated with the account.
//
//		You must be authenticated as an admin to use this endpoint. The account creation process involves validating the input request,
//		creating the account in the database, and setting the initial account status. If the account is successfully created, a success
//		response is returned. Otherwise, an error response is returned.
//	 	Additionally it will start the associated runner and connect it to Self Network
//
// @Tags            accounts
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           request body CreateAccountRequest true "The request body must contain the account details in JSON format."
// @Success         201  {object}  CreateAccountResponse "Upon successful account creation, the response will contain the username, resources associated with the account,
//
//	and a flag indicating whether a password change is required."
//
// @Failure         400  {object}  response.Error "If the request body is invalid, a bad request error is returned."
// @Failure         403  {object}  response.Error "If the user is not authenticated as an admin, a forbidden error is returned."
// @Failure         500  {object}  response.Error "If there is any internal server error, a generic error is returned."
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
		RequiresPasswordChange: (account.RequiresPasswordChange == 1),
	})
}

// DeleteAccount godoc
// @Summary         Deletes an existing account.
// @Description     This endpoint deletes an existing account from the system.
//
//		It also ensures that no further communications are received from the deleted account. You must be authenticated as an admin to use this endpoint.
//		The account deletion process involves validating the account existence, deleting the account from the database, and setting the account status as deleted.
//		If the account is successfully deleted, a success response is returned. Otherwise, an error response is returned.
//	 Additionally all issued tokens for this user will be invalidated.
//
// @Tags            accounts
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           username   path      string  true  "The username of the account to be deleted."
// @Success         204  {string} string  "Upon successful account deletion, a no content response is returned."
// @Failure         404 {object} response.Error "If the account does not exist or the user is not authenticated as an admin, a not found error is returned."
// @Failure         500 {object} response.Error "If there is any internal server error, a generic error is returned."
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
// @Description     This endpoint changes the password for the provided username.
//
//	The user must provide their current password and the new password they wish to use.
//	The process involves validating the user's current password, updating the password in the database,
//	and returning a success response if the operation was successful.
//	If the operation fails for any reason, an error response is returned.
//
// @Tags            accounts
// @Accept          json
// @Produce         json
// @Security        BearerAuth
// @Param           username   path      string  true  "The username of the account for which the password is to be changed."
// @Param           request body ChangePasswordRequest true "The request body must contain the current password and the new password in JSON format."
// @Success         200  {string} string "Upon successful password change, an OK response is returned."
// @Failure         400 {object} response.Error "If the request body is invalid or the current password is incorrect, a bad request error is returned."
// @Failure         404 {object} response.Error "If the user does not exist or is not authenticated, a not found error is returned."
// @Failure         500 {object} response.Error "If there is any internal server error, a generic error is returned."
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

// ListAccounts godoc
// @Summary        Lists all configured accounts.
// @Description    This endpoint retrieves and lists all the accounts that have been configured in the system.
//
//	The accounts are returned in a paginated format, with the total number of accounts and the accounts for the current page included in the response.
//	You must be authenticated as an admin to use this endpoint. If the operation is successful, a list of accounts is returned.
//	If the operation fails, an error response is returned.
//
// @Tags           accounts
// @Accept         json
// @Produce        json
// @Security       BearerAuth
// @Param          page     query    int     false    "The page number to retrieve. Defaults to 1 if not provided."
// @Param          limit    query    int     false    "The number of accounts to retrieve per page. Defaults to 10 if not provided."
// @Success        200 {object} ExtListResponse "Upon successful operation, a paginated list of accounts is returned."
// @Failure        404 {object} response.Error "Not found - The requested resource does not exist, or you don't have permissions to access it"
// @Router         /accounts [get]
func (r resource) list(c echo.Context) error {
	user := acl.CurrentUser(c)
	if user == nil || !user.IsAdmin() {
		return c.JSON(response.DefaultNotFoundError())
	}

	apps := r.service.List(c.Request().Context())
	pages := pagination.NewFromRequest(c.Request(), len(apps))
	pages.Items = apps

	return c.JSON(http.StatusOK, pages)
}
