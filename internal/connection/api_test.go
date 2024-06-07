package connection

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

func (m mockService) Get(ctx context.Context, appid, selfid string) (Connection, error) {
	if selfid == "not_found_id" {
		return Connection{}, errors.New("expected not found")
	}
	return Connection{
		entity.Connection{
			SelfID: "selfid",
			AppID:  appid,
			Name:   "name",
		},
	}, nil
}

func (m mockService) Query(ctx context.Context, appid string, offset, limit int) ([]Connection, error) {
	conns := []Connection{}
	if appid == "query_error" {
		return []Connection{}, errors.New("expected_error_query")
	}
	conns = append(conns, Connection{
		entity.Connection{
			SelfID: "selfid",
			AppID:  appid,
			Name:   "name",
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

func (m mockService) Create(ctx context.Context, appid string, input CreateConnectionRequest) (Connection, error) {
	if input.SelfID == "controlled_error" {
		return Connection{}, errors.New("controlled error")
	}
	return Connection{}, nil
}

func (m mockService) Update(ctx context.Context, appid, selfid string, input UpdateConnectionRequest) (Connection, error) {
	if input.Name == "controlled_error" {
		return Connection{}, errors.New("controlled error")
	}
	return Connection{}, nil
}

func (m mockService) Delete(ctx context.Context, appid, selfid string) (Connection, error) {
	if selfid == "controlled_error" {
		return Connection{}, errors.New("controlled error")
	}
	return Connection{}, nil
}

func TestListConnectionsAPIEndpointAsAdmin(t *testing.T) {
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
			URL:          "/apps/app_id/connections",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusOK,
			WantResponse: `{"items":[{"app_id":"app_id", "created_at":"0001-01-01T00:00:00Z", "id":"selfid", "name":"name", "updated_at":"0001-01-01T00:00:00Z"}], "page":1, "page_count":1, "per_page":100, "total_count":1}`,
		},
		{
			Name:         "internal error on count",
			Method:       "GET",
			URL:          "/apps/count_error/connections",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusInternalServerError,
			WantResponse: `There was a problem with your request. *`,
		},
		{
			Name:         "internal error on query",
			Method:       "GET",
			URL:          "/apps/query_error/connections",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusInternalServerError,
			WantResponse: `There was a problem with your request. *`,
		},
		{
			Name:         "internal error on query",
			Method:       "GET",
			URL:          "/apps/query_error/connections",
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

func TestListConnectionsAPIEndpointAsPlain(t *testing.T) {
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
			URL:          "/apps/app_id/connections",
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

func TestGetConnectionAPIEndpointAsPlainWithPermissions(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsPlainMiddleware([]string{"GET /apps/app_id/connections/conn_id"}))
	rg.Use(acl.NewMiddleware(filter.NewChecker()).TokenAndAccessCheckMiddleware)
	RegisterHandlers(rg, mockService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "GET",
			URL:          "/apps/app_id/connections/conn_id",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusOK,
			WantResponse: `{"app_id":"app_id", "created_at":"0001-01-01T00:00:00Z", "id":"selfid", "name":"name", "updated_at":"0001-01-01T00:00:00Z"}`,
		},
		{
			Name:         "success",
			Method:       "GET",
			URL:          "/apps/app_id/connections/not_found_id",
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

func TestGetConnectionAPIEndpointAsPlainWithoutPermissions(t *testing.T) {
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
			URL:          "/apps/app_id/connections/conn_id",
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

func TestCreateConnectionAPIEndpointAsPlainWithoutPermissions(t *testing.T) {
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
			URL:          "/apps/app_id/connections",
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

func TestCreateConnectionAPIEndpoint(t *testing.T) {
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
			URL:          "/apps/app_id/connections",
			Body:         `{"selfid":"selfid"}`,
			Header:       nil,
			WantStatus:   http.StatusCreated,
			WantResponse: `{"app_id":"", "created_at":"0001-01-01T00:00:00Z", "id":"", "name":"", "updated_at":"0001-01-01T00:00:00Z"}`,
		},
		{
			Name:         "invalid input",
			Method:       "POST",
			URL:          "/apps/app_id/connections",
			Body:         `{]`,
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"The provided body is not valid", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "validation error",
			Method:       "POST",
			URL:          "/apps/app_id/connections",
			Body:         `{}`,
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"selfid: cannot be blank.", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "creation error",
			Method:       "POST",
			URL:          "/apps/app_id/connections",
			Body:         `{"selfid":"controlled_error"}`,
			Header:       nil,
			WantStatus:   http.StatusInternalServerError,
			WantResponse: `There was a problem with your request. *`,
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}

func TestUpdateConnectionAPIEndpointAsPlainWithoutPermissions(t *testing.T) {
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
			URL:          "/apps/app_id/connections/conn_id",
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

func TestUpdateConnectionAPIEndpoint(t *testing.T) {
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
			URL:          "/apps/app_id/connections/conn_id",
			Body:         `{"name":"new_name"}`,
			Header:       nil,
			WantStatus:   http.StatusOK,
			WantResponse: `{"app_id":"", "created_at":"0001-01-01T00:00:00Z", "id":"", "name":"", "updated_at":"0001-01-01T00:00:00Z"}`,
		},
		{
			Name:         "invalid input",
			Method:       "PUT",
			URL:          "/apps/app_id/connections/conn_id",
			Body:         `{]`,
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"The provided body is not valid", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "validation error",
			Method:       "PUT",
			URL:          "/apps/app_id/connections/conn_id",
			Body:         `{"name":""}`,
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"name: cannot be blank.", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "modification error",
			Method:       "PUT",
			URL:          "/apps/app_id/connections/conn_id",
			Body:         `{"name":"controlled_error"}`,
			Header:       nil,
			WantStatus:   http.StatusInternalServerError,
			WantResponse: `There was a problem with your request. *`,
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}

func TestDeleteConnectionAPIEndpointAsPlainWithoutPermissions(t *testing.T) {
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
			URL:          "/apps/app_id/connections/conn_id",
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

func TestDeleteConnectionAPIEndpoint(t *testing.T) {
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
			URL:          "/apps/app_id/connections/conn_id",
			Body:         `{"name":"new_name"}`,
			Header:       nil,
			WantStatus:   http.StatusOK,
			WantResponse: `{"app_id":"", "created_at":"0001-01-01T00:00:00Z", "id":"", "name":"", "updated_at":"0001-01-01T00:00:00Z"}`,
		},
		{
			Name:         "modification error",
			Method:       "DELETE",
			URL:          "/apps/app_id/connections/controlled_error",
			Body:         `{"name":"controlled_error"}`,
			Header:       nil,
			WantStatus:   http.StatusNotFound,
			WantResponse: `{"status":404,"error":"Not found","details":"The requested resource does not exist, or you don't have permissions to access it"}`,
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}
