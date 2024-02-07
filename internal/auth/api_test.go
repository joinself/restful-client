package auth

import (
	"context"
	"net/http"
	"testing"

	"github.com/joinself/restful-client/internal/errors"
	"github.com/joinself/restful-client/internal/test"
	"github.com/joinself/restful-client/pkg/log"
)

type mockService struct{}

func (m mockService) Login(ctx context.Context, username, password string) (LoginResponse, error) {
	if username == "test" && password == "pass" {
		return LoginResponse{AccessToken: "token-100", RefreshToken: "r-token-100"}, nil
	}
	return LoginResponse{"", ""}, errors.Unauthorized("")
}

func (m mockService) Refresh(ctx context.Context, token string) (LoginResponse, error) {
	if token == "test" {
		return LoginResponse{"token-100", "r-token-100"}, nil
	}
	return LoginResponse{AccessToken: "", RefreshToken: ""}, errors.Unauthorized("")
}

func TestAPI(t *testing.T) {
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
			WantResponse: "",
		},
		{
			Name:         "bad json",
			Method:       "POST",
			URL:          "/login",
			Body:         `"username":"test_larger","password":"wrong pass"}`,
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: "",
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}
