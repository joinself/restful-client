package auth

import (
	"context"
	e "errors"
	"net/http"
	"testing"

	"github.com/joinself/restful-client/internal/errors"
	"github.com/joinself/restful-client/internal/test"
	"github.com/joinself/restful-client/pkg/log"
)

type mockService struct{}

func (m mockService) Login(ctx context.Context, username, password string) (LoginResponse, error) {
	if username == "test_larger" && password == "pass_larger" {
		return LoginResponse{AccessToken: "token-100", RefreshToken: "r-token-100"}, nil
	}
	return LoginResponse{"", ""}, errors.Unauthorized("")
}

func (m mockService) Refresh(ctx context.Context, token string) (LoginResponse, error) {
	if token == "test_token_test_token_test_token_test_token_test_token_" {
		return LoginResponse{"token-100", "r-token-100"}, nil
	}
	if token == "test_token_test_token_test_token_test_token_test_token_error" {
		return LoginResponse{}, e.New("error message")
	}
	return LoginResponse{AccessToken: "", RefreshToken: ""}, errors.Unauthorized("")
}

func TestLoginAPIEndpoint(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	RegisterHandlers(router.Group(""), mockService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "POST",
			URL:          "/login",
			Body:         `{"username":"test_larger","password":"pass_larger"}`,
			Header:       nil,
			WantStatus:   http.StatusOK,
			WantResponse: `{"token":"token-100","refresh_token":"r-token-100"}`,
		},
		{
			Name:         "bad credential",
			Method:       "POST",
			URL:          "/login",
			Body:         `{"username":"test_larger","password":"wrong pass"}`,
			Header:       nil,
			WantStatus:   http.StatusUnauthorized,
			WantResponse: `{"details":"Provided auth credentials are invalid", "error":"You're unauthorized to perform this action", "status":401}`,
		},
		{
			Name:         "bad json",
			Method:       "POST",
			URL:          "/login",
			Body:         `"username":"test_larger","password":"wrong pass"}`,
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"The provided body is not valid", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "invalid data",
			Method:       "POST",
			URL:          "/login",
			Body:         `{"username":"","password":""}`,
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"password: cannot be blank; username: cannot be blank.", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "invalid username",
			Method:       "POST",
			URL:          "/login",
			Body:         `{"username":"foo","password":"pass_larger"}`,
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"username: the length must be between 5 and 128.", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "invalid password",
			Method:       "POST",
			URL:          "/login",
			Body:         `{"username":"test_larger","password":"foo"}`,
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"password: the length must be between 5 and 128.", "error":"Invalid input", "status":400}`,
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}

func TestRefreshAPIEndpoint(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	RegisterHandlers(router.Group(""), mockService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "POST",
			URL:          "/refresh",
			Body:         `{"refresh_token":"test_token_test_token_test_token_test_token_test_token_"}`,
			Header:       nil,
			WantStatus:   http.StatusOK,
			WantResponse: `{"token":"token-100","refresh_token":"r-token-100"}`,
		},
		{
			Name:         "bad json",
			Method:       "POST",
			URL:          "/refresh",
			Body:         `{]`,
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"The provided body is not valid", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "invalid data",
			Method:       "POST",
			URL:          "/refresh",
			Body:         `{"refresh_token":""}`,
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"You must provide a refresh_token", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "invalid refresh token",
			Method:       "POST",
			URL:          "/refresh",
			Body:         `{"refresh_token":"test_token_test_token_test_token_test_token_test_token_error"}`,
			Header:       nil,
			WantStatus:   http.StatusUnauthorized,
			WantResponse: `{"details":"You've provided a refresh_token, but it's not valid", "error":"You're unauthorized to perform this action", "status":401}`,
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}
