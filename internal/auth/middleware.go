package auth

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/joinself/restful-client/pkg/acl"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

// Handler returns a JWT-based authentication middleware.
func Handler(verificationKey string) echo.MiddlewareFunc {
	config := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(acl.JWTCustomClaims)
		},
		SigningKey: []byte(verificationKey),
	}

	return echojwt.WithConfig(config)
}

type contextKey int

const (
	userKey contextKey = iota
)

// WithUser returns a context that contains the user identity from the given JWT.
func WithUser(ctx echo.Context, token *jwt.Token) {
	ctx.Set("user", token)
}
