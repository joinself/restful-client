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
	rg.POST("/refresh", refresh(service, logger))
}

// Login godoc
// @Summary Authenticate User and Retrieve Tokens
// @Description This endpoint authenticates user credentials (username and password) and, upon successful authentication,
// issues a JWT (JSON Web Token) for accessing protected endpoints. Additionally, a refresh token is provided for generating
// new JWTs once the original token expires. The JWT and refresh token are both necessary for seamless and secure user sessions.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Login credentials object containing 'username' and 'password' fields."
// @Success 200 {object} LoginResponse "Authentication successful: Returns the JWT for API access and a refresh token."
// @Failure 400 {object} response.Error "Bad Request: The request is malformed or the JSON body cannot be parsed."
// @Failure 401 {object} response.Error "Unauthorized: Authentication failed due to invalid credentials or inactive user account."
// @Failure 500 {object} response.Error "Internal Server Error: An unexpected error occurred while processing the authentication request."
// @Router /auth/login [post]
func login(service Service, logger log.Logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req LoginRequest

		if err := c.Bind(&req); err != nil {
			logger.With(c.Request().Context()).Warnf("invalid request: %v", err)
			return c.JSON(response.DefaultBadRequestError())
		}

		if reqErr := req.Validate(); reqErr != nil {
			logger.With(c.Request().Context()).Warnf("invalid request: %s", reqErr.Details)
			return c.JSON(reqErr.Status, reqErr)
		}

		var err error
		var resp LoginResponse
		var ctx = c.Request().Context()

		resp, err = service.Login(ctx, req.Username, req.Password)
		if err != nil {
			logger.With(c.Request().Context()).Warnf("problem logging in - %v", err)
			return c.JSON(http.StatusUnauthorized, response.Error{
				Status:  http.StatusUnauthorized,
				Error:   "You're unauthorized to perform this action",
				Details: "Provided auth credentials are invalid",
			})
		}
		logger.With(c.Request().Context()).Infof("user %s logged in", req.Username)

		return c.JSON(http.StatusOK, resp)
	}
}

// RefreshToken godoc
// @Summary Refresh JWT Token
// @Description This endpoint is used to refresh an expired or about to expire JWT token.
// It requires a valid refresh token to be provided in the request body. Upon validation
// of the refresh token, a new JWT token is issued for continued API access.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param refreshToken body RefreshRequest true "Contains the refresh token that needs to be validated and exchanged for a new JWT token."
// @Success 200 {object} LoginResponse "A new JWT token is successfully generated and returned along with its expiry information."
// @Failure 400 {object} response.Error "Bad Request: The request is invalid or malformed. The error message provides more details."
// @Failure 401 {object} response.Error "Unauthorized: The provided refresh token is invalid or expired, and a new JWT token cannot be issued."
// @Failure 500 {object} response.Error "Internal Server Error: An unexpected error occurred while processing the refresh token request."
// @Router /auth/refresh [post]
func refresh(service Service, logger log.Logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req RefreshRequest

		if err := c.Bind(&req); err != nil {
			logger.With(c.Request().Context()).Warnf("invalid request: %v", err)
			return c.JSON(response.DefaultBadRequestError())
		}

		if reqErr := req.Validate(); reqErr != nil {
			logger.With(c.Request().Context()).Warnf("invalid request: %s", reqErr.Details)
			return c.JSON(reqErr.Status, reqErr)
		}

		var err error
		var resp LoginResponse
		var ctx = c.Request().Context()

		resp, err = service.Refresh(ctx, req.RefreshToken)
		if err != nil {
			logger.With(c.Request().Context()).Warnf("problem refreshing token - %v", err)
			return c.JSON(http.StatusUnauthorized, response.Error{
				Status:  http.StatusUnauthorized,
				Error:   "You're unauthorized to perform this action",
				Details: "You've provided a refresh_token, but it's not valid",
			})
		}

		logger.With(c.Request().Context()).Info("successful token refresh")

		return c.JSON(http.StatusOK, resp)
	}
}
