package auth

import (
	"net/http"

	"github.com/joinself/restful-client/pkg/log"
	"github.com/labstack/echo/v4"
)

type AuthRequest struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	RefreshToken string `json:"refresh_token"`
}

type AuthResponse struct {
	AccessToken  string `json:"token"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// RegisterHandlers registers handlers for different HTTP requests.
func RegisterHandlers(rg *echo.Group, service Service, logger log.Logger) {
	rg.POST("/login", login(service, logger))
}

// Login 	    godoc
// @Summary     Authenticate.
// @Description Get a temporary JWT token to interact with the api.
// @Tags        login
// @Accept      json
// @Produce     json
// @Param       request   body      AuthRequest  true  "Self ID"
// @Success     200  {object}  AuthResponse
// @Router      /login [post]
func login(service Service, logger log.Logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req AuthRequest

		if err := c.Bind(&req); err != nil {
			logger.With(c.Request().Context()).Errorf("invalid request: %v", err)
			return c.JSON(http.StatusBadRequest, "")
		}

		// If refresh token is present validate and return a valid auth token
		var err error
		var resp AuthResponse
		ctx := c.Request().Context()

		if req.RefreshToken != "" { // Is a refresh based auth workflow.
			resp, err = service.Refresh(ctx, req.RefreshToken)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, err)
			}
		} else { // Is a basic auth workflow.
			resp, err = service.Login(ctx, req.Username, req.Password)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, err)
			}
		}

		return c.JSON(http.StatusOK, resp)
	}
}
