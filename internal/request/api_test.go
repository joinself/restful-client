package request

import (
	"net/http"
	"testing"

	"github.com/joinself/restful-client/internal/test"
	"github.com/joinself/restful-client/pkg/acl"
	"github.com/joinself/restful-client/pkg/filter"
	"github.com/joinself/restful-client/pkg/log"
)

func TestGetRequestAPIEndpointAsPlainWithPermissions(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsPlainMiddleware([]string{"GET /apps/app_id/requests/request_jti"}))
	rg.Use(acl.NewMiddleware(filter.NewChecker()).TokenAndAccessCheckMiddleware)
	RegisterHandlers(rg, mockService{}, mockConnectionService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "GET",
			URL:          "/apps/app_id/requests/request_jti",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusOK,
			WantResponse: `{"app_id":"app_id", "id":""}`,
		},
		{
			Name:         "connection not found",
			Method:       "GET",
			URL:          "/apps/app_id/connections/not_found_id/requests/request_jti",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusNotFound,
			WantResponse: `{"status":404,"error":"Not found","details":"The requested resource does not exist, or you don't have permissions to access it"}`,
		},
		{
			Name:         "connection not found",
			Method:       "GET",
			URL:          "/apps/app_id/requests/not_found_id",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusNotFound,
			WantResponse: `{"status":404,"error":"Not found","details":"The requested resource does not exist, or you don't have permissions to access it"}`,
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}

func TestGetRequestAPIEndpointAsPlainWithoutPermissions(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsPlainMiddleware([]string{}))
	rg.Use(acl.NewMiddleware(filter.NewChecker()).TokenAndAccessCheckMiddleware)
	RegisterHandlers(rg, mockService{}, mockConnectionService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "GET",
			URL:          "/apps/app_id/requests/request_jti",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusNotFound,
			WantResponse: `{"status":404,"error":"Not found","details":"The requested resource does not exist, or you don't have permissions to access it"}`,
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}

func TestCreateRequestAPIEndpointAsPlainWithoutPermissions(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsPlainMiddleware([]string{}))
	rg.Use(acl.NewMiddleware(filter.NewChecker()).TokenAndAccessCheckMiddleware)
	RegisterHandlers(rg, mockService{}, mockConnectionService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "not found",
			Method:       "POST",
			URL:          "/apps/app_id/requests",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusNotFound,
			WantResponse: `{"status":404,"error":"Not found","details":"The requested resource does not exist, or you don't have permissions to access it"}`,
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}

func TestCreateRequestAPIEndpoint(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsAdminMiddleware())
	rg.Use(acl.NewMiddleware(filter.NewChecker()).TokenAndAccessCheckMiddleware)
	RegisterHandlers(rg, mockService{}, mockConnectionService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "POST",
			URL:          "/apps/app_id/requests",
			Body:         `{"type":"auth","facts":[{"name":"myfact"}]}`,
			Header:       nil,
			WantStatus:   http.StatusAccepted,
			WantResponse: ``,
		},
		{
			Name:         "invalid input",
			Method:       "POST",
			URL:          "/apps/app_id/requests",
			Body:         `{]`,
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"The provided body is not valid", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "validation error fact",
			Method:       "POST",
			URL:          "/apps/app_id/requests",
			Body:         `{"type":"auth","facts":[{"name":""}]}`,
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"name: cannot be blank.", "error":"Invalid input", "status":400}`,
		},

		{
			Name:         "validation error",
			Method:       "POST",
			URL:          "/apps/app_id/requests",
			Body:         `{}`,
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"type: cannot be blank.", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "creation error",
			Method:       "POST",
			URL:          "/apps/error/requests",
			Body:         `{"type":"auth","facts":[{"name":"myfact"}]}`,
			Header:       nil,
			WantStatus:   http.StatusInternalServerError,
			WantResponse: `There was a problem with your request. *`,
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}
