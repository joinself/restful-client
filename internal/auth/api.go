package auth

import (
	"net/http"

	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/response"
	"github.com/labstack/echo/v4"
)

// RegisterHandlers registers handlers for different HTTP requests.
func RegisterHandlers(rg *echo.Group, service Service, logger log.Logger) {
	rg.POST("/login", login(service, logger))
}

// Login godoc
// @Summary User Authentication
// @Description Authenticates a user and returns a temporary JWT token and refresh token for API interaction.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Authentication request body with your username and password, or a refresh token"
// @Success 200 {object} LoginResponse "Successfully authenticated, JWT token and Refresh JWT token are returned in response"
// @Failure 401,400 {object} response.Error "Returns error details"
// @Router /login [post]
func login(service Service, logger log.Logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req LoginRequest

		if err := c.Bind(&req); err != nil {
			logger.With(c.Request().Context()).Errorf("invalid request: %v", err)
			return c.JSON(response.DefaultBadRequestError())
		}

		if reqErr := req.Validate(); reqErr != nil {
			return c.JSON(reqErr.Status, reqErr)
		}

		var err error
		var resp LoginResponse
		var ctx = c.Request().Context()

		if req.RefreshToken != "" { // Is a refresh based auth workflow.
			resp, err = service.Refresh(ctx, req.RefreshToken)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, response.Error{
					Status:  http.StatusUnauthorized,
					Error:   "You're unauthorized to perform this action",
					Details: "You've provided a refresh_token, but it's not valid",
				})
			}
		} else { // Is a basic auth workflow.
			resp, err = service.Login(ctx, req.Username, req.Password)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, response.Error{
					Status:  http.StatusUnauthorized,
					Error:   "You're unauthorized to perform this action",
					Details: "Provided auth credentials are invalid",
				})
			}
		}

		return c.JSON(http.StatusOK, resp)
	}
}
