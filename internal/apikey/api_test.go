package apikey

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/internal/test"
	"github.com/joinself/restful-client/pkg/acl"
	"github.com/joinself/restful-client/pkg/filter"
	"github.com/joinself/restful-client/pkg/log"
)

type mockService struct{}

func (m mockService) Get(ctx context.Context, appid string, id int) (ApiKey, error) {
	if id == 0 {
		return ApiKey{}, errors.New("expected not found")
	}
	return ApiKey{
		entity.Apikey{
			ID:    1,
			AppID: appid,
			Name:  "name",
		},
	}, nil
}

func (m mockService) Query(ctx context.Context, appid string, offset, limit int) ([]ApiKey, error) {
	conns := []ApiKey{}
	if appid == "query_error" {
		return []ApiKey{}, errors.New("expected_error_query")
	}
	conns = append(conns, ApiKey{
		entity.Apikey{
			ID:    1,
			AppID: appid,
			Name:  "name",
		},
	})
	return conns, nil
}

func (m mockService) Count(ctx context.Context, appid string) (int, error) {
	if appid == "count_error" {
		return 0, errors.New("expected_error_count")
	}
	return 1, nil
}

func (m mockService) Create(ctx context.Context, appid string, input CreateApiKeyRequest, user acl.Identity) (ApiKey, error) {
	if input.Name == "error" {
		return ApiKey{}, errors.New("controlled error")
	}
	return ApiKey{}, nil
}

func (m mockService) Update(ctx context.Context, appid string, id int, input UpdateApiKeyRequest) (ApiKey, error) {
	if input.Name == "0" {
		return ApiKey{}, errors.New("controlled error")
	}
	return ApiKey{}, nil
}

func (m mockService) Delete(ctx context.Context, appid string, id int) (ApiKey, error) {
	if id == 0 {
		return ApiKey{}, errors.New("controlled error")
	}
	return ApiKey{}, nil
}

func TestListApiKeysAPIEndpointAsAdmin(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsAdminMiddleware())
	rg.Use(acl.NewMiddleware(filter.NewChecker()).TokenAndAccessCheckMiddleware)
	RegisterHandlers(rg, mockService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "GET",
			URL:          "/apps/app_id/apikeys",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusOK,
			WantResponse: `{"items":[{"app_id":"app_id", "created_at":"0001-01-01T00:00:00Z", "id":1, "name":"name", "token":"", "updated_at":"0001-01-01T00:00:00Z"}], "page":1, "page_count":1, "per_page":100, "total_count":1}`,
		},
		{
			Name:         "internal error on count",
			Method:       "GET",
			URL:          "/apps/count_error/apikeys",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusInternalServerError,
			WantResponse: `There was a problem with your request. *`,
		},
		{
			Name:         "internal error on query",
			Method:       "GET",
			URL:          "/apps/query_error/apikeys",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusInternalServerError,
			WantResponse: `There was a problem with your request. *`,
		},
		{
			Name:         "internal error on query",
			Method:       "GET",
			URL:          "/apps/query_error/apikeys",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusInternalServerError,
			WantResponse: `There was a problem with your request. *`,
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}

func TestListApiKeysAPIEndpointAsPlain(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsPlainMiddleware([]string{}))
	rg.Use(acl.NewMiddleware(filter.NewChecker()).TokenAndAccessCheckMiddleware)
	RegisterHandlers(rg, mockService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "not found",
			Method:       "GET",
			URL:          "/apps/app_id/apikeys",
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

func TestGetApiKeyAPIEndpointAsPlainWithPermissions(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsPlainMiddleware([]string{"GET /apps/app_id/apikeys/1"}))
	rg.Use(acl.NewMiddleware(filter.NewChecker()).TokenAndAccessCheckMiddleware)
	RegisterHandlers(rg, mockService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "GET",
			URL:          "/apps/app_id/apikeys/1",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusOK,
			WantResponse: `{"app_id":"app_id", "created_at":"0001-01-01T00:00:00Z", "id":1, "name":"name", "token":"", "updated_at":"0001-01-01T00:00:00Z"}`,
		},
		{
			Name:         "success",
			Method:       "GET",
			URL:          "/apps/app_id/apikeys/0",
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

func TestGetApiKeyAPIEndpointAsPlainWithoutPermissions(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsPlainMiddleware([]string{}))
	rg.Use(acl.NewMiddleware(filter.NewChecker()).TokenAndAccessCheckMiddleware)
	RegisterHandlers(rg, mockService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "GET",
			URL:          "/apps/app_id/apikeys/1",
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

func TestCreateApiKeyAPIEndpointAsPlainWithoutPermissions(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsPlainMiddleware([]string{}))
	rg.Use(acl.NewMiddleware(filter.NewChecker()).TokenAndAccessCheckMiddleware)
	RegisterHandlers(rg, mockService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "not found",
			Method:       "POST",
			URL:          "/apps/app_id/apikeys",
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

func TestCreateApiKeyAPIEndpoint(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsAdminMiddleware())
	rg.Use(acl.NewMiddleware(filter.NewChecker()).TokenAndAccessCheckMiddleware)
	RegisterHandlers(rg, mockService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "POST",
			URL:          "/apps/app_id/apikeys",
			Body:         `{"name":"non-blank-name", "scope":"FULL"}`,
			Header:       nil,
			WantStatus:   http.StatusCreated,
			WantResponse: `{"app_id":"", "created_at":"0001-01-01T00:00:00Z", "id":0, "name":"", "token":"", "updated_at":"0001-01-01T00:00:00Z"}`,
		},
		{
			Name:         "invalid input",
			Method:       "POST",
			URL:          "/apps/app_id/apikeys",
			Body:         `{]`,
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"The provided body is not valid", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "validation error",
			Method:       "POST",
			URL:          "/apps/app_id/apikeys",
			Body:         `{}`,
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"name: cannot be blank; scope: cannot be blank.", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "creation error",
			Method:       "POST",
			URL:          "/apps/app_id/apikeys",
			Body:         `{"name":"error", "scope":"FULL"}`,
			Header:       nil,
			WantStatus:   http.StatusInternalServerError,
			WantResponse: `There was a problem with your request. *`,
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}

func TestUpdateApiKeyAPIEndpointAsPlainWithoutPermissions(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsPlainMiddleware([]string{}))
	rg.Use(acl.NewMiddleware(filter.NewChecker()).TokenAndAccessCheckMiddleware)
	RegisterHandlers(rg, mockService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "not found",
			Method:       "PUT",
			URL:          "/apps/app_id/apikeys/1",
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

func TestUpdateApiKeyAPIEndpoint(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsAdminMiddleware())
	rg.Use(acl.NewMiddleware(filter.NewChecker()).TokenAndAccessCheckMiddleware)
	RegisterHandlers(rg, mockService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "PUT",
			URL:          "/apps/app_id/apikeys/1",
			Body:         `{"name":"new_name"}`,
			Header:       nil,
			WantStatus:   http.StatusOK,
			WantResponse: `{"app_id":"", "created_at":"0001-01-01T00:00:00Z", "id":0, "name":"", "token":"", "updated_at":"0001-01-01T00:00:00Z"}`,
		},
		{
			Name:         "invalid input",
			Method:       "PUT",
			URL:          "/apps/app_id/apikeys/1",
			Body:         `{]`,
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"The provided body is not valid", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "validation error",
			Method:       "PUT",
			URL:          "/apps/app_id/apikeys/1",
			Body:         `{"name":""}`,
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"name: cannot be blank.", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "modification error",
			Method:       "PUT",
			URL:          "/apps/app_id/apikeys/1",
			Body:         `{"name":"0"}`,
			Header:       nil,
			WantStatus:   http.StatusInternalServerError,
			WantResponse: `There was a problem with your request. *`,
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}

func TestDeleteApiKeyAPIEndpointAsPlainWithoutPermissions(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsPlainMiddleware([]string{}))
	rg.Use(acl.NewMiddleware(filter.NewChecker()).TokenAndAccessCheckMiddleware)
	RegisterHandlers(rg, mockService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "not found",
			Method:       "DELETE",
			URL:          "/apps/app_id/apikeys/1",
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

func TestDeleteApiKeyAPIEndpoint(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsAdminMiddleware())
	rg.Use(acl.NewMiddleware(filter.NewChecker()).TokenAndAccessCheckMiddleware)
	RegisterHandlers(rg, mockService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "DELETE",
			URL:          "/apps/app_id/apikeys/1",
			Body:         `{"name":"new_name"}`,
			Header:       nil,
			WantStatus:   http.StatusOK,
			WantResponse: `{"app_id":"", "created_at":"0001-01-01T00:00:00Z", "id":0, "name":"", "token":"", "updated_at":"0001-01-01T00:00:00Z"}`,
		},
		{
			Name:         "modification error",
			Method:       "DELETE",
			URL:          "/apps/app_id/apikeys/0",
			Body:         `{"name":"0"}`,
			Header:       nil,
			WantStatus:   http.StatusNotFound,
			WantResponse: `{"status":404,"error":"Not found","details":"The requested resource does not exist, or you don't have permissions to access it"}`,
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}
