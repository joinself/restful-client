package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/internal/test"
	"github.com/joinself/restful-client/pkg/acl"
	"github.com/joinself/restful-client/pkg/filter"
	"github.com/joinself/restful-client/pkg/log"
)

type mockService struct{}

func (m mockService) List(ctx context.Context) []entity.App {
	return []entity.App{}
}

func (m mockService) ListByStatus(ctx context.Context, statuses []string) ([]entity.App, error) {
	return []entity.App{}, nil
}

func (m mockService) Get(ctx context.Context, id string) (App, error) {
	if id == erroredUUID {
		return App{}, errors.New("expected error")
	}
	return App{}, nil
}

func (m mockService) Create(ctx context.Context, input CreateAppRequest) (App, error) {
	if input.ID == erroredUUID {
		return App{}, errors.New("expected error")
	}

	return App{
		entity.App{
			Name:   "test",
			ID:     "test",
			Status: "testing",
			Env:    "test",
		},
	}, nil
}

func (m mockService) Delete(ctx context.Context, id string) (App, error) {
	if id == "error" {
		return App{}, errors.New("expected error")
	}
	return App{}, nil
}

func (m mockService) Update(ctx context.Context, id string, input UpdateAppRequest) (App, error) {
	if id == "error" {
		return App{}, errors.New("expected error")
	}
	return App{}, nil
}

func TestListAppsAPIEndpointAsAdmin(t *testing.T) {
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
			URL:          "/apps",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusOK,
			WantResponse: `{"items":[], "page":1, "page_count":0, "per_page":100, "total_count":0}`,
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}

func TestListAppsAPIEndpointAsPlain(t *testing.T) {
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
			URL:          "/apps",
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

const (
	validUUID   = "00000000-0000-0000-0000-000000000000"
	invalidUUID = "o"
	erroredUUID = "11111111-1111-1111-1111-111111111111"
	validKey    = "sk_1:0000000000000000000000000000000000000000000"
	invalidKey  = "098120730129783"
	validName   = "name"
	validEnv    = "sandbox"
)

func TestCreateAppAPIEndpointAsAdmin(t *testing.T) {
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
			URL:          "/apps",
			Body:         fmt.Sprintf(`{"id":"%s","secret":"%s","name":"%s","env":"%s"}`, validUUID, validKey, validName, validEnv),
			Header:       nil,
			WantStatus:   http.StatusOK,
			WantResponse: `{"env":"test", "id":"test", "name":"test", "status":"testing"}`,
		},
		{
			Name:         "UUID validation error",
			Method:       "POST",
			URL:          "/apps",
			Body:         fmt.Sprintf(`{"id":"%s","secret":"%s","name":"%s","env":"%s"}`, invalidUUID, validKey, validName, validEnv),
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"id: not valid UUID.", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "Secret validation error",
			Method:       "POST",
			URL:          "/apps",
			Body:         fmt.Sprintf(`{"id":"%s","secret":"%s","name":"%s","env":"%s"}`, validUUID, "foo", validName, validEnv),
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"secret: not valid secret.", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "invalid input",
			Method:       "POST",
			URL:          "/apps",
			Body:         `{]`,
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"The provided body is not valid", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "errored creation",
			Method:       "POST",
			URL:          "/apps",
			Body:         fmt.Sprintf(`{"id":"%s","secret":"%s","name":"%s","env":"%s"}`, erroredUUID, validKey, validName, validEnv),
			Header:       nil,
			WantStatus:   http.StatusInternalServerError,
			WantResponse: `There was a problem with your request. *`,
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}

func TestCreateAppAPIEndpointAsPlain(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsPlainMiddleware([]string{}))
	rg.Use(acl.NewMiddleware(filter.NewChecker()).TokenAndAccessCheckMiddleware)
	RegisterHandlers(rg, mockService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "POST",
			URL:          "/apps",
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

func TestDeleteAppAPIEndpointAsAdmin(t *testing.T) {
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
			URL:          "/apps/app",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusOK,
			WantResponse: ``,
		},
		{
			Name:         "error deleting",
			Method:       "DELETE",
			URL:          "/apps/error",
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
func TestDeleteAppAPIEndpointAsPlain(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsPlainMiddleware([]string{}))
	rg.Use(acl.NewMiddleware(filter.NewChecker()).TokenAndAccessCheckMiddleware)
	RegisterHandlers(rg, mockService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "DELETE",
			URL:          "/apps/app",
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

func TestUpdateAppAPIEndpointAsAdmin(t *testing.T) {
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
			URL:          "/apps/app",
			Body:         `{"callback":"http://localhost","callback_secret":"secret"}`,
			Header:       nil,
			WantStatus:   http.StatusOK,
			WantResponse: ``,
		},
		{
			Name:         "invalid-body",
			Method:       "PUT",
			URL:          "/apps/app",
			Body:         `[}`,
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"The provided body is not valid", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "error updating",
			Method:       "PUT",
			URL:          "/apps/error",
			Body:         `{"callback":"http://localhost","callback_secret":"secret"}`,
			Header:       nil,
			WantStatus:   http.StatusNotFound,
			WantResponse: `{"status":404,"error":"Not found","details":"The requested resource does not exist, or you don't have permissions to access it"}`,
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}

func TestUpdateAppAPIEndpointAsPlain(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsPlainMiddleware([]string{"PUT /apps/app_id"}))
	rg.Use(acl.NewMiddleware(filter.NewChecker()).TokenAndAccessCheckMiddleware)
	RegisterHandlers(rg, mockService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "PUT",
			URL:          "/apps/app_id",
			Body:         `{"callback":"http://localhost","callback_secret":"secret"}`,
			Header:       nil,
			WantStatus:   http.StatusOK,
			WantResponse: ``,
		},
		{
			Name:         "unaccessible-resource",
			Method:       "PUT",
			URL:          "/apps/unaccessible",
			Body:         `{"callback":"http://localhost","callback_secret":"secret"}`,
			Header:       nil,
			WantStatus:   http.StatusNotFound,
			WantResponse: `{"status":404,"error":"Not found","details":"The requested resource does not exist, or you don't have permissions to access it"}`,
		},
		{
			Name:         "invalid-body",
			Method:       "PUT",
			URL:          "/apps/app_id",
			Body:         `[}`,
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"The provided body is not valid", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "error updating",
			Method:       "PUT",
			URL:          "/apps/error",
			Body:         `{"callback":"http://localhost","callback_secret":"secret"}`,
			Header:       nil,
			WantStatus:   http.StatusNotFound,
			WantResponse: `{"status":404,"error":"Not found","details":"The requested resource does not exist, or you don't have permissions to access it"}`,
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}

func TestGetAppAPIEndpointAsAdmin(t *testing.T) {
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
			URL:          "/apps/app",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusOK,
			WantResponse: ``,
		},
		{
			Name:         "error getting",
			Method:       "GET",
			URL:          "/apps/" + erroredUUID,
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
func TestGetAppAPIEndpointAsPlain(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsPlainMiddleware([]string{"GET /apps/app_id"}))
	rg.Use(acl.NewMiddleware(filter.NewChecker()).TokenAndAccessCheckMiddleware)
	RegisterHandlers(rg, mockService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "GET",
			URL:          "/apps/app_id",
			Body:         `{"callback":"http://localhost","callback_secret":"secret"}`,
			Header:       nil,
			WantStatus:   http.StatusOK,
			WantResponse: ``,
		},
		{
			Name:         "unaccessible-resource",
			Method:       "GET",
			URL:          "/apps/unaccessible",
			Body:         `{"callback":"http://localhost","callback_secret":"secret"}`,
			Header:       nil,
			WantStatus:   http.StatusNotFound,
			WantResponse: `{"status":404,"error":"Not found","details":"The requested resource does not exist, or you don't have permissions to access it"}`,
		},
		{
			Name:         "error getting",
			Method:       "GET",
			URL:          "/apps/error",
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
