package apikey

import (
	"net/http"
	"strconv"

	"github.com/joinself/restful-client/pkg/acl"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/pagination"
	"github.com/joinself/restful-client/pkg/response"
	"github.com/labstack/echo/v4"
)

// RegisterHandlers sets up the routing of the HTTP handlers.
func RegisterHandlers(r *echo.Group, service Service, logger log.Logger) {
	res := resource{service, logger}

	r.GET("/:app_id/apikeys/:id", res.get)
	r.GET("/:app_id/apikeys", res.query)

	r.POST("/:app_id/apikeys", res.create)
	r.PUT("/:app_id/apikeys/:id", res.update)
	r.DELETE("/:app_id/apikeys/:id", res.delete)
}

type resource struct {
	service Service
	logger  log.Logger
}

// GetApiKey godoc
// @Summary      Retrieve specific API key details
// @Description  Retrieves the specifics of an API key using the provided app_id and API key id.
//
//	The caller must have sufficient permissions to access this information.
//	This operation is secured using Bearer Authentication.
//	If the operation is successful, the API key details are returned.
//	If the operation fails, an error response is returned.
//
// @Tags         apikeys
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        app_id   path   string  true  "Unique Identifier (UUID) that represents a specific App"
// @Param        id   path  int  true  "Unique Identifier (integer) that represents a specific API key"
// @Success      200  {object}  ExtApiKey  "Upon successful operation, the API key details are returned."
// @Failure      400  {object}  response.Error "If the app_id or id parameters are not valid, a bad request error is returned."
// @Failure      403  {object}  response.Error "If the user does not have sufficient permissions to access the API key details, a forbidden error is returned."
// @Failure      404  {object}  response.Error "If the API key does not exist, a not found error is returned."
// @Failure      500  {object}  response.Error "If there is any internal server error, a generic error is returned."
// @Router       /apps/{app_id}/apikeys/{id} [get]
func (r resource) get(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		r.logger.With(c.Request().Context()).Info("invalid input when getting an apikey - %v", err)
		return c.JSON(response.DefaultNotFoundError())
	}

	ak, err := r.service.Get(c.Request().Context(), c.Param("app_id"), id)
	if err != nil {
		r.logger.With(c.Request().Context()).Info("invalid input when retrieving an apikey - %v", err)
		return c.JSON(response.DefaultNotFoundError())
	}

	return c.JSON(http.StatusOK, ExtApiKey{
		ID:        ak.ID,
		AppID:     ak.AppID,
		Name:      ak.Name,
		Token:     ak.Token,
		CreatedAt: ak.CreatedAt,
		UpdatedAt: ak.UpdatedAt,
	})
}

// ListApiKeys godoc
// @Summary        Retrieve a paginated list of API keys
// @Description    Retrieves a paginated list of API keys for a specific app_id, matching the specified filters.
//
//	Pagination is provided through optional page and per_page parameters.
//	If not provided, the defaults are page 1 and per_page 10.
//	The caller must have sufficient permissions to access this information.
//	This operation is secured using Bearer Authentication.
//	If the operation is successful, a paginated list of API keys is returned.
//	If the operation fails, an error response is returned.
//
// @Tags           apikeys
// @Accept         json
// @Produce        json
// @Security       BearerAuth
// @Param          app_id   path   string  true  "Unique Identifier (UUID) for the App"
// @Param          page query int false "The page number for pagination. If not provided, the default is 1."
// @Param          per_page query int false "The number of API keys to return per page for pagination. If not provided, the default is 10."
// @Success        200  {object}  ExtListResponse  "Upon successful operation, a paginated list of API keys is returned."
// @Failure        400  {object}  response.Error "If the app_id, page, or per_page parameters are not valid, a bad request error is returned."
// @Failure        403  {object}  response.Error "If the user does not have sufficient permissions to access the API keys, a forbidden error is returned."
// @Failure        404  {object}  response.Error "If the app does not exist, a not found error is returned."
// @Failure        500  {object}  response.Error "If there is any internal server error, a generic error is returned."
// @Router         /apps/{app_id}/apikeys [get]
func (r resource) query(c echo.Context) error {
	ctx := c.Request().Context()
	count, err := r.service.Count(ctx, c.Param("app_id"))
	if err != nil {
		r.logger.With(c.Request().Context()).Warnf("err counting apikey - %v", err)
		return c.JSON(response.DefaultInternalError(c, r.logger, err.Error()))
	}

	pages := pagination.NewFromRequest(c.Request(), count)
	apikeys, err := r.service.Query(ctx,
		c.Param("app_id"),
		pages.Offset(),
		pages.Limit())
	if err != nil {
		r.logger.With(c.Request().Context()).Warnf("error listing apikeys - %v", err)
		return c.JSON(response.DefaultInternalError(c, r.logger, err.Error()))
	}

	aks := []ExtApiKey{}
	for _, ak := range apikeys {
		aks = append(aks, ExtApiKey{
			ID:        ak.ID,
			AppID:     ak.AppID,
			Name:      ak.Name,
			Token:     ak.Token,
			CreatedAt: ak.CreatedAt,
			UpdatedAt: ak.UpdatedAt,
		})
	}

	pages.Items = aks
	return c.JSON(http.StatusOK, pages)
}

// CreateApiKey godoc
// @Summary Create a new API key
// @Description This API endpoint creates a new API key for a specific app.
//
//	The app ID and request body containing the API key details are required as input.
//	The source field in the request body should be one of the following types: FULL, MESSAGING, or REQUESTS.
//	Once the API key is created, a request is sent for public information.
//	The caller must have sufficient permissions to create an API key.
//	This operation is secured using Bearer Authentication.
//	If the operation is successful, the details of the new API key are returned.
//	If the operation fails, an error response is returned.
//
// @Tags apikeys
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param app_id path string true "Unique identifier of the app"
// @Param request body CreateApiKeyRequest true "Body containing details of the API key to be created. The source should be one of the following types: FULL, MESSAGING, or REQUESTS."
// @Success 201 {object} ExtApiKey "Upon successful operation, the details of the new API key are returned."
// @Failure 400 {object} response.Error "If the app_id or request body parameters are not valid, a bad request error is returned."
// @Failure 403 {object} response.Error "If the user does not have sufficient permissions to create an API key, a forbidden error is returned."
// @Failure 404 {object} response.Error "If the app does not exist, a not found error is returned."
// @Failure 500 {object} response.Error "If there is any internal server error, a generic error is returned."
// @Router /apps/{app_id}/apikeys [post]
func (r resource) create(c echo.Context) error {
	var input CreateApiKeyRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Warnf("invalid request: %v", err)
		return c.JSON(response.DefaultBadRequestError())
	}

	if err := input.Validate(); err != nil {
		r.logger.With(c.Request().Context()).Warnf("invalid request: %s", err.Details)
		return c.JSON(err.Status, err)
	}

	ak, err := r.service.Create(c.Request().Context(), c.Param("app_id"), input, acl.CurrentUser(c))
	if err != nil {
		r.logger.With(c.Request().Context()).Warnf("err creating apikey - %v", err)
		return c.JSON(response.DefaultInternalError(c, r.logger, err.Error()))
	}

	return c.JSON(http.StatusCreated, ExtApiKey{
		ID:        ak.ID,
		AppID:     ak.AppID,
		Name:      ak.Name,
		Token:     ak.Token,
		CreatedAt: ak.CreatedAt,
		UpdatedAt: ak.UpdatedAt,
	})
}

// UpdateApiKey godoc
// @Summary Update an API key
// @Description This endpoint updates the properties of an existing API key for a specific application.
//
//	The application ID, API key ID, and request body containing the updated API key details are required as input.
//	This includes changes to the API key's name and other modifiable attributes.
//	The caller must have sufficient permissions to update an API key.
//	This operation is secured using Bearer Authentication.
//	If the operation is successful, the updated API key details are returned.
//	If the operation fails, an error response is returned.
//
// @Tags apikeys
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param app_id path string true "Unique Identifier (UUID) for the Application. This is used to ensure that the request is made within the context of a specific application."
// @Param id path string true "Unique Identifier for the API key that is to be updated. This allows the server to identify the specific API key that needs updating."
// @Param request body UpdateApiKeyRequest true "Request body that must contain the updated details of the API key. This includes any changes to the API key's attributes that are allowed to be modified."
// @Success 200 {object} ExtApiKey "Upon successful operation, the updated API key details are returned."
// @Failure 400 {object} response.Error "If the app_id, id, or request body parameters are not valid, a bad request error is returned."
// @Failure 403 {object} response.Error "If the user does not have sufficient permissions to update the API key, a forbidden error is returned."
// @Failure 404 {object} response.Error "If the app or API key does not exist, a not found error is returned."
// @Failure 500 {object} response.Error "If there is any internal server error, a generic error is returned."
// @Router /apps/{app_id}/apikeys/{id} [put]
func (r resource) update(c echo.Context) error {
	var input UpdateApiKeyRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Warnf("invalid request: %v", err)
		return c.JSON(response.DefaultBadRequestError())
	}

	if err := input.Validate(); err != nil {
		r.logger.With(c.Request().Context()).Warnf("invalid request: %s", err.Details)
		return c.JSON(err.Status, err)
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(response.DefaultNotFoundError())
	}

	conn, err := r.service.Update(c.Request().Context(), c.Param("app_id"), id, input)
	if err != nil {
		r.logger.With(c.Request().Context()).Warnf("err updating apikey - %v", err)
		return c.JSON(response.DefaultInternalError(c, r.logger, err.Error()))
	}

	return c.JSON(http.StatusOK, ExtApiKey{
		ID:        conn.ID,
		AppID:     conn.AppID,
		Name:      conn.Name,
		CreatedAt: conn.CreatedAt,
		UpdatedAt: conn.UpdatedAt,
	})
}

// DeleteApiKey godoc
// @Summary Delete an API key
// @Description This endpoint deletes an existing API key using the provided app_id and API key id.
//
//	The caller must have sufficient permissions to delete an API key.
//	This operation is secured using Bearer Authentication.
//	If the operation is successful, the details of the deleted API key are returned.
//	If the operation fails, an error response is returned.
//
// @Tags apikeys
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param app_id path string true "Unique Identifier (UUID) for the Application. This is used to ensure that the request is made within the context of a specific application."
// @Param id path string true "Unique Identifier for the API key that is to be deleted. This allows the server to identify the specific API key that needs deleting."
// @Success 200 {object} ExtApiKey "Upon successful operation, the details of the deleted API key are returned."
// @Failure 400 {object} response.Error "If the app_id or id parameters are not valid, a bad request error is returned."
// @Failure 403 {object} response.Error "If the user does not have sufficient permissions to delete the API key, a forbidden error is returned."
// @Failure 404 {object} response.Error "If the app or API key does not exist, a not found error is returned."
// @Failure 500 {object} response.Error "If there is any internal server error, a generic error is returned."
// @Router /apps/{app_id}/apikeys/{id} [delete]
func (r resource) delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(response.DefaultNotFoundError())
	}
	conn, err := r.service.Delete(c.Request().Context(), c.Param("app_id"), id)
	if err != nil {
		r.logger.With(c.Request().Context()).Warnf("err deleting apikey - %v", err)
		return c.JSON(response.DefaultInternalError(c, r.logger, err.Error()))
	}

	return c.JSON(http.StatusOK, ExtApiKey{
		ID:        conn.ID,
		AppID:     conn.AppID,
		Name:      conn.Name,
		CreatedAt: conn.CreatedAt,
		UpdatedAt: conn.UpdatedAt,
	})
}
