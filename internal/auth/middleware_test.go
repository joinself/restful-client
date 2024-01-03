package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestCurrentUser(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}"))
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	assert.Nil(t, CurrentUser(ctx))
	WithUser(ctx, buildJwt(entity.User{
		ID:        "100",
		Name:      "test",
		Admin:     false,
		Resources: []string{"app1"},
	}))

	identity := CurrentUser(ctx)
	assert.NotNil(t, identity)
	assert.Equal(t, "100", identity.GetID())
	assert.Equal(t, "test", identity.GetName())
	assert.Equal(t, false, identity.IsAdmin())
	rs := identity.GetResources()
	assert.Equal(t, 1, len(rs))
	assert.Equal(t, "app1", rs[0])
}

func testHasAccessToResource(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}"))
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	assert.Nil(t, CurrentUser(ctx))
	WithUser(ctx, buildJwt(entity.User{
		ID:        "100",
		Name:      "test",
		Admin:     false,
		Resources: []string{"app1"},
	}))
	assert.False(t, HasAccessToResource(ctx, "app2"))
	assert.False(t, HasAccessToResource(ctx, "app1"))

	WithUser(ctx, buildJwt(entity.User{
		ID:        "100",
		Name:      "test",
		Admin:     true,
		Resources: []string{},
	}))
	assert.True(t, HasAccessToResource(ctx, "app2"))
	assert.True(t, HasAccessToResource(ctx, "app1"))
}

func TestHandler(t *testing.T) {
	assert.NotNil(t, Handler("test"))
}

func buildJwt(identity Identity) *jwt.Token {
	tokenExpiration := 1000

	// Set custom claims
	claims := &jwtCustomClaims{
		identity.GetID(),
		identity.GetName(),
		identity.IsAdmin(),
		identity.GetResources(),
		jwt.RegisteredClaims{
			Subject:   identity.GetID(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(tokenExpiration))),
		},
	}

	// Create token with claims
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
}
