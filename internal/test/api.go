package test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// APITestCase represents the data needed to describe an API test case.
type APITestCase struct {
	Name         string
	Method, URL  string
	Body         string
	Header       http.Header
	WantStatus   int
	WantResponse string
}

// Endpoint tests an HTTP endpoint using the given APITestCase spec.
func Endpoint(t *testing.T, router *echo.Echo, tc APITestCase) {
	t.Run(tc.Name, func(t *testing.T) {
		req, _ := http.NewRequest(tc.Method, tc.URL, bytes.NewBufferString(tc.Body))
		if tc.Header != nil {
			req.Header = tc.Header
		}
		res := httptest.NewRecorder()
		if req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", "application/json")
		}

		//-------
		// token, _ := buildAdminJwt("test")
		// req.Header.Set("Authorization", "Bearer "+token)
		//-------

		router.ServeHTTP(res, req)
		assert.Equal(t, tc.WantStatus, res.Code, "status mismatch")
		if tc.WantResponse != "" {
			pattern := strings.Trim(tc.WantResponse, "*")
			if pattern != tc.WantResponse {
				assert.Contains(t, res.Body.String(), pattern, "response mismatch")
			} else {
				assert.JSONEq(t, tc.WantResponse, res.Body.String(), "response mismatch")
			}
		}
	})
}

type jwtCustomClaims struct {
	ID                       string   `json:"id"`
	Name                     string   `json:"name"`
	Admin                    bool     `json:"admin"`
	Resources                []string `json:"resources"`
	IsPasswordChangeRequired bool     `json:"change_password"`
	jwt.RegisteredClaims
}

func buildAdminJwt(key string) (string, error) {
	tokenExpiration := 1000

	// Set custom claims
	claims := &jwtCustomClaims{
		"0",
		"admin",
		true,
		[]string{},
		false,
		jwt.RegisteredClaims{
			Subject:   "0",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(tokenExpiration))),
		},
	}

	// Create token with claims
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(key)
}
