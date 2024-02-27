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
// @Summary      Retrieve specific apikey details
// @Description  Retrieves the specifics of an apikey by using the provided app_id and apikey id.
//
//	The caller must have sufficient permissions to access this information.
//	The method employs Bearer Authentication for secure access.
//
// @Tags         apikeys
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        app_id   path   string  true  "Unique Identifier (UUID) that represents a specific App"
// @Param        id   path  int  true  "Unique Identifier (integer) that represents a specific apikey"
// @Success      200  {object}  ExtApiKey  "Successful retrieval of apikey details will return a 200 status and the apikey object"
// @Failure      404  {object}  response.Error "Failure scenarios include inability to locate the requested resource or insufficient permissions to access it. These will return a 404 status and an error object"
// @Router       /apps/{app_id}/apikeys/{id} [get]
func (r resource) get(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(response.DefaultNotFoundError())
	}

	ak, err := r.service.Get(c.Request().Context(), c.Param("app_id"), id)
	if err != nil {
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
// @Summary        Retrieve a paginated list of apikeys
// @Description    Retrieves a paginated list of apikeys for a specific app_id, matching the specified filters.
//
//	Pagination is provided through optional page and per_page parameters.
//	If not provided, the defaults are page 1 and per_page 10.
//
// @Tags           apikeys
// @Accept         json
// @Produce        json
// @Security       BearerAuth
// @Param          app_id   path   string  true  "Unique Identifier (UUID) for the App"
// @Param          page query int false "The page number for pagination. If not provided, the default is 1."
// @Param          per_page query int false "The number of apikeys to return per page for pagination. If not provided, the default is 10."
// @Success        200  {object}  ExtListResponse  "Successful retrieval of apikeys list will return a 200 status and a list of apikeys"
// @Failure        500  {object}  response.Error "In case of an internal server error during the request, a 500 status and an error object will be returned"
// @Router         /apps/{app_id}/apikeys [get]
func (r resource) query(c echo.Context) error {
	ctx := c.Request().Context()
	count, err := r.service.Count(ctx, c.Param("app_id"))
	if err != nil {
		return c.JSON(response.DefaultInternalError(c, r.logger, err.Error()))
	}

	pages := pagination.NewFromRequest(c.Request(), count)
	apikeys, err := r.service.Query(ctx,
		c.Param("app_id"),
		pages.Offset(),
		pages.Limit())
	if err != nil {
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
// @Summary Create a new apikey
// @Description This API endpoint creates a new apikey by taking the application ID and request body as input.
// The input CreateApiKeyRequest.source should be of a type FULL, MESSAGING, or REQUESTS.
// It sends a request for public information once the apikey is created.
// @Tags apikeys
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param app_id path string true "Unique identifier of the application"
// @Param request body CreateApiKeyRequest true "Body containing details of the apikey to be created. The source should be of a type FULL, MESSAGING, or REQUESTS."
// @Success 201 {object} ExtApiKey "Successfully created a new apikey and returns the details of the new apikey"
// @Failure 400 {object} response.Error "Returns when the provided input is invalid"
// @Failure 500 {object} response.Error "Returns when there is an internal server error"
// @Router /apps/{app_id}/apikeys [post]
func (r resource) create(c echo.Context) error {
	var input CreateApiKeyRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return c.JSON(response.DefaultBadRequestError())
	}

	if reqErr := input.Validate(); reqErr != nil {
		return c.JSON(reqErr.Status, reqErr)
	}

	ak, err := r.service.Create(c.Request().Context(), c.Param("app_id"), input, acl.CurrentUser(c))
	if err != nil {
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
// @Summary Update an apikey
// @Description This endpoint updates the properties of an existing apikey using the provided app_id and apikey id. This includes changes to the apikey's name and other modifiable attributes. The updated details must be provided in the request body.
// @Tags apikeys
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param app_id path string true "Unique Identifier (UUID) for the Application. This is used to ensure that the request is made within the context of a specific application."
// @Param id path string true "Unique Identifier for the ApiKey that is to be updated. This allows the server to identify the specific apikey that needs updating."
// @Param request body UpdateApiKeyRequest true "Request body that must contain the updated details of the apikey. This includes any changes to the apikey's attributes that are allowed to be modified."
// @Success 200 {object} ExtApiKey "Upon successful update, the apikey's updated details are returned in the response body."
// @Failure 400 {object} response.Error "A 400 status code is returned when the inputs provided in the request are invalid or malformed."
// @Failure 500 {object} response.Error "A 500 status code is returned when an internal server error occurs during the processing of the request."
// @Router /apps/{app_id}/apikeys/{id} [put]
func (r resource) update(c echo.Context) error {
	var input UpdateApiKeyRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return c.JSON(response.DefaultBadRequestError())
	}

	if reqErr := input.Validate(); reqErr != nil {
		return c.JSON(reqErr.Status, reqErr)
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(response.DefaultNotFoundError())
	}

	conn, err := r.service.Update(c.Request().Context(), c.Param("app_id"), id, input)
	if err != nil {
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

// UpdateApiKey godoc
// @Summary Update an apikey
// @Description This endpoint updates the properties of an existing apikey using the provided app_id and apikey id. This includes changes to the apikey's name and other modifiable attributes. The updated details must be provided in the request body.
// @Tags apikeys
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param app_id path string true "Unique Identifier (UUID) for the Application. This is used to ensure that the request is made within the context of a specific application."
// @Param id path string true "Unique Identifier for the ApiKey that is to be updated. This allows the server to identify the specific apikey that needs updating."
// @Param request body UpdateApiKeyRequest true "Request body that must contain the updated details of the apikey. This includes any changes to the apikey's attributes that are allowed to be modified."
// @Success 200 {object} ExtApiKey "Upon successful update, the apikey's updated details are returned in the response body."
// @Failure 400 {object} response.Error "A 400 status code is returned when the inputs provided in the request are invalid or malformed."
// @Failure 500 {object} response.Error "A 500 status code is returned when an internal server error occurs during the processing of the request."
// @Router /apps/{app_id}/apikeys/{id} [put]
func (r resource) delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(response.DefaultNotFoundError())
	}
	conn, err := r.service.Delete(c.Request().Context(), c.Param("app_id"), id)
	if err != nil {
		return c.JSON(response.DefaultNotFoundError())
	}

	return c.JSON(http.StatusOK, ExtApiKey{
		ID:        conn.ID,
		AppID:     conn.AppID,
		Name:      conn.Name,
		CreatedAt: conn.CreatedAt,
		UpdatedAt: conn.UpdatedAt,
	})
}
