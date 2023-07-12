package test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
		res := sendRequest(router, tc.Method, tc.URL, tc.Body, tc.Header)
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

func SendRequest(router *echo.Echo, method, url, body string, header http.Header) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, url, bytes.NewBufferString(body))
	if header != nil {
		req.Header = header
	}
	res := httptest.NewRecorder()
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	router.ServeHTTP(res, req)

	return res
}
